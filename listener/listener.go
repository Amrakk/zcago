package listener

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/websocketx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

const (
	ZaloManualClosure       int = 1000
	ZaloAbnormalClosure     int = 1006
	ZaloDuplicateConnection int = 3000
	ZaloKickConnection      int = 3003
)

type Listener interface {
	Start(ctx context.Context, retryOnClose bool) error
	Stop() error

	// Channels
	Connected() <-chan struct{}
	Disconnected() <-chan websocketx.CloseInfo
	Closed() <-chan websocketx.CloseInfo
	Error() <-chan error
	// Typing() <-chan Typing
	Message() <-chan model.UserMessage
	// OldMessage() <-chan OldMessages
	// SeenMessage() <-chan []SeenMessage
	// DeliveredMessage() <-chan []DeliveredMessage
	// Reaction() <-chan Reaction
	// OldReaction() <-chan OldReactions
	// UploadAttachment() <-chan UploadAttachment
	Undo() <-chan model.Undo
	// Friend() <-chan Friend
	// Group() <-chan Group
	CipherKey() <-chan string

	SendWS(ctx context.Context, payload WSPayload, requireID bool) error

	RequestOldMessages(ctx context.Context, threadType model.ThreadType, lastMsgID *string) error
	RequestOldReactions(ctx context.Context, threadType model.ThreadType, lastMsgID *string) error
}

type listener struct {
	mu sync.RWMutex

	ch    channels
	reqID uint64

	client websocketx.Client
	sc     session.MutableContext

	urls      []string
	wsURL     string
	cookie    string
	userAgent string

	retryStates map[string]*retryState
	rotateCount int

	cipherKey string

	selfListen  bool
	pingStopper *func()

	ctx context.Context
}

var _ Listener = (*listener)(nil)

func New(sc session.MutableContext, urls []string) (*listener, error) {
	if err := validateInputs(sc, urls); err != nil {
		return nil, err
	}

	wsURL, err := buildWebSocketURL(sc, urls[0])
	if err != nil {
		return nil, err
	}

	cookieStr := buildCookieString(sc.Cookies())
	retryStates := buildRetryStates(sc)
	channels := initializeChannels()

	return &listener{
		sc: sc,

		ch:    channels,
		reqID: 0,

		urls:      urls,
		wsURL:     wsURL,
		cookie:    cookieStr,
		userAgent: sc.UserAgent(),

		retryStates: retryStates,
		rotateCount: 0,

		cipherKey:   "",
		selfListen:  sc.Options().SelfListen,
		pingStopper: nil,
	}, nil
}

func (ln *listener) createWebSocketConnection(ctx context.Context) (websocketx.Client, error) {
	u, err := url.Parse(ln.wsURL)
	if err != nil {
		return nil, errs.NewZCAError("parse websocket URL failed", "listener.Start", &err)
	}

	h := make(http.Header)
	h.Set("accept-encoding", "gzip, deflate, br, zstd")
	h.Set("accept-language", "en-US,en;q=0.9")
	h.Set("cache-control", "no-cache")
	h.Set("host", u.Host)
	h.Set("origin", "https://chat.zalo.me")
	h.Set("pragma", "no-cache")
	h.Set("user-agent", ln.userAgent)
	h.Set("cookie", ln.cookie)

	client, err := websocketx.Dial(ctx, ln.wsURL, &websocketx.Options{
		Header: h,
		Proxy:  ln.sc.Proxy(),
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (ln *listener) Start(ctx context.Context, retryOnClose bool) error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.client != nil {
		return errs.NewZCAError("Already started", "listener.Start", nil)
	}
	if ctx.Err() != nil {
		err := ctx.Err()
		return errs.NewZCAError("context cancelled", "listener.Start", &err)
	}

	ln.ctx = ctx

	client, err := ln.createWebSocketConnection(ctx)
	if err != nil {
		return err
	}

	ln.client = client

	select {
	case ln.ch.Connected <- struct{}{}:
	default:
	}

	go ln.run(retryOnClose)

	return nil
}

func (ln *listener) run(retryOnClose bool) {
	cl := ln.getClient()
	if cl == nil {
		return
	}

	ctx, cancel := context.WithCancel(ln.ctx)
	defer cancel()

	errsCh := cl.Errors()
	msgsCh := cl.Messages()
	closCh := cl.Closed()

	for {
		select {
		case <-ctx.Done():
			return

		case err, ok := <-errsCh:
			if !ok {
				errsCh = nil
				continue
			}
			ln.emitError(ctx, err)

		case msg, ok := <-msgsCh:
			if !ok {
				msgsCh = nil
				continue
			}
			ln.handleWebSocketMessage(ctx, msg)

		case ci, ok := <-closCh:
			if !ok {
				return
			}
			ln.handleConnectionClose(ctx, ci, retryOnClose)
			return
		}

		if errsCh == nil && msgsCh == nil && closCh == nil {
			return
		}
	}
}

func (ln *listener) handleConnectionClose(ctx context.Context, ci websocketx.CloseInfo, retryOnClose bool) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.reset()

	select {
	case ln.ch.Disconnected <- ci:
	case <-ctx.Done():
		return
	default:
	}

	if ln.shouldRetryConnection(ctx, ci, retryOnClose) {
		ln.scheduleReconnection(ci)
	} else {
		ln.emitClosed(ctx, ci)
	}
}

func (ln *listener) Stop() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.client == nil {
		return nil
	}

	client := ln.client
	ln.reset()
	client.Close(ZaloManualClosure, "")

	return nil
}

func (ln *listener) reset() {
	ln.client = nil
	ln.reqID = 0
	ln.cipherKey = ""
	ln.ctx = nil
	if ln.pingStopper != nil {
		(*ln.pingStopper)()
		ln.pingStopper = nil
	}
}

//
// Constructor helpers
//

func validateInputs(sc session.MutableContext, urls []string) error {
	if sc == nil {
		return errs.NewZCAError("context is nil", "listener.New", nil)
	}
	if sc.CookieJar() == nil {
		return errs.NewZCAError("cookie jar is not available", "listener.New", nil)
	}
	if ua := sc.UserAgent(); ua == "" {
		return errs.NewZCAError("user-agent is not available", "listener.New", nil)
	}
	if len(urls) == 0 || urls[0] == "" {
		return errs.NewZCAError("websocket URL list is empty", "listener.New", nil)
	}
	return nil
}

func buildWebSocketURL(sc session.MutableContext, url string) (string, error) {
	wsURL := httpx.MakeURL(sc, url, map[string]any{
		"t": time.Now().UnixMilli(),
	}, true)
	if wsURL == "" {
		return "", errs.NewZCAError("build websocket URL failed", "listener.New", nil)
	}
	return wsURL, nil
}

func buildCookieString(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}
	var b strings.Builder
	first := true
	for _, c := range cookies {
		if !first {
			b.WriteString("; ")
		}
		first = false
		b.WriteString(c.Name)
		b.WriteByte('=')
		b.WriteString(c.Value)
	}
	return b.String()
}

func buildRetryStates(sc session.MutableContext) map[string]*retryState {
	retryStates := make(map[string]*retryState, 8)
	if s := sc.Settings(); s != nil && s.Features.Socket.Retries != nil {
		for reason, cfg := range s.Features.Socket.Retries {
			max := uint(0)
			if cfg.Max != nil {
				max = *cfg.Max
			}

			times := cfg.Times.Slice()
			if len(times) == 0 {
				continue
			}
			retryStates[reason] = &retryState{
				count: 0,
				max:   max,
				times: append([]uint(nil), times...),
			}
		}
	}
	return retryStates
}

func initializeChannels() channels {
	buf := defaultBuffers()
	return channels{
		Connected:    make(chan struct{}, buf.Connected),
		Disconnected: make(chan websocketx.CloseInfo, buf.Disconnected),
		Closed:       make(chan websocketx.CloseInfo, buf.Closed),
		Error:        make(chan error, buf.Error),
		// Typing: make(chan Typing, buf.Typing),
		Message: make(chan model.UserMessage, buf.Message),
		// OldMessages: make(chan OldMessagesEvent, buf.OldMessages),
		// SeenMessages: make(chan []SeenMessage, buf.SeenMessages),
		// DeliveredMessages: make(chan []DeliveredMessage, buf.DeliveredMessages),
		// Reaction: make(chan Reaction, buf.Reaction),
		// OldReactions: make(chan OldReactionsEvent, buf.OldReactions),
		// UploadAttachment: make(chan UploadEventData, buf.UploadAttachment),
		Undo: make(chan model.Undo, buf.Undo),
		// Friend: make(chan FriendEvent, buf.Friend),
		// Group: make(chan GroupEvent, buf.Group),
		CipherKey: make(chan string, buf.CipherKey),
	}
}
