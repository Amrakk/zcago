package listener

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Amrakk/zcago/errs"
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
	Stop()

	// Channels
	Connected() <-chan struct{}
	Disconnected() <-chan websocketx.CloseInfo
	Closed() <-chan websocketx.CloseInfo
	Error() <-chan error
	Message() <-chan model.Message
	// OldMessage() <-chan OldMessages
	// Reaction() <-chan Reaction
	// OldReaction() <-chan OldReactions
	// Typing() <-chan Typing
	// DeliveredMessage() <-chan []DeliveredMessage
	// SeenMessage() <-chan []SeenMessage
	Undo() <-chan model.Undo
	UploadAttachment() <-chan model.UploadAttachment
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
	userAgent string

	retryStates map[string]*retryState
	rotateCount int

	cipherKey string

	selfListen  bool
	pingStopper *func()

	cancel context.CancelFunc
	wg     sync.WaitGroup
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

	retryStates := buildRetryStates(sc)
	channels := initializeChannels()

	return &listener{
		sc: sc,

		ch:    channels,
		reqID: 0,

		urls:      urls,
		wsURL:     wsURL,
		userAgent: sc.UserAgent(),

		retryStates: retryStates,
		rotateCount: 0,

		cipherKey:   "",
		selfListen:  sc.Options().SelfListen,
		pingStopper: nil,
	}, nil
}

func (ln *listener) Start(ctx context.Context, retryOnClose bool) error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.client != nil {
		return errs.NewZCA("Already started", "listener.Start")
	}
	if ctx.Err() != nil {
		err := ctx.Err()
		return errs.WrapZCA("context cancelled", "listener.Start", err)
	}

	lctx, cancel := context.WithCancel(ctx)
	ln.cancel = cancel

	client, err := ln.createWebSocketConnection(lctx)
	if err != nil {
		cancel()
		return err
	}

	ln.client = client

	select {
	case ln.ch.Connected <- struct{}{}:
	default:
	}

	ln.wg.Add(1)
	go ln.run(lctx, retryOnClose)

	return nil
}

func (ln *listener) createWebSocketConnection(ctx context.Context) (websocketx.Client, error) {
	u, err := url.Parse(ln.wsURL)
	if err != nil {
		return nil, errs.WrapZCA("parse websocket URL failed", "listener.createWebSocketConnection", err)
	}

	h := make(http.Header)
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Cache-Control", "no-cache")
	h.Set("Host", u.Host)
	h.Set("Origin", "https://chat.zalo.me")
	h.Set("Pragma", "no-cache")
	h.Set("User-Agent", ln.userAgent)

	client, err := websocketx.Dial(ctx, ln.wsURL, &websocketx.Options{
		Header:     h,
		HTTPClient: ln.sc.Client(),
	})
	if err != nil {
		return nil, errs.WrapZCA("websocket dial failed", "listener.createWebSocketConnection", err)
	}

	return client, nil
}

func (ln *listener) run(ctx context.Context, retryOnClose bool) {
	defer ln.wg.Done()

	cl := ln.getClient()
	if cl == nil {
		return
	}

	errsCh := cl.Errors()
	msgsCh := cl.Messages()
	closCh := cl.Closed()
	open := 3

	for open > 0 {
		select {
		case <-ctx.Done():
			return

		case err, ok := <-errsCh:
			if !ok {
				errsCh = nil
				open--
				continue
			}
			ln.emitError(ctx, err)

		case msg, ok := <-msgsCh:
			if !ok {
				msgsCh = nil
				open--
				continue
			}
			ln.handleWebSocketMessage(ctx, msg)

		case ci, ok := <-closCh:
			if !ok {
				closCh = nil
				open--
				continue
			}
			ln.handleConnectionClose(ctx, ci, retryOnClose)
			return
		}
	}
}

func (ln *listener) handleConnectionClose(ctx context.Context, ci websocketx.CloseInfo, retryOnClose bool) {
	ln.reset()

	select {
	case ln.ch.Disconnected <- ci:
	case <-ctx.Done():
		return
	default:
	}

	if delay, ok := ln.shouldRetryConnection(ctx, ci, retryOnClose); ok {
		if err := ln.scheduleReconnection(ctx, ci, delay); err != nil {
			ln.emitError(ctx, errs.WrapZCA("failed to schedule reconnection:", "listener.handleConnectionClose", err))
			ln.emitClosed(ctx, ci)
			ln.cancelActiveContext()
		}
		return
	}

	ln.emitClosed(ctx, ci)
	ln.cancelActiveContext()
}

func (ln *listener) Stop() {
	client := ln.getClient()
	if client == nil {
		return
	}

	ln.cancelActiveContext()
	client.Close(ZaloManualClosure, "")

	ln.wg.Wait()
}

func (ln *listener) reset() {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.pingStopper != nil {
		(*ln.pingStopper)()
		ln.pingStopper = nil
	}

	if ln.client != nil {
		ln.client = nil
	}

	ln.reqID = 0
	ln.cipherKey = ""
}

func (ln *listener) cancelActiveContext() {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	if ln.cancel != nil {
		ln.cancel()
		ln.cancel = nil
	}
}

func (ln *listener) getClient() websocketx.Client {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	return ln.client
}

// ----------------------------------------
// Constructor helpers
// ----------------------------------------

func validateInputs(sc session.MutableContext, urls []string) error {
	if sc == nil {
		return errs.NewZCA("context is nil", "listener.validateInputs")
	}
	if sc.CookieJar() == nil {
		return errs.NewZCA("cookie jar is not available", "listener.validateInputs")
	}
	if ua := sc.UserAgent(); ua == "" {
		return errs.NewZCA("user-agent is not available", "listener.validateInputs")
	}
	if len(urls) == 0 || urls[0] == "" {
		return errs.NewZCA("websocket URL list is empty", "listener.validateInputs")
	}
	return nil
}

func buildWebSocketURL(sc session.MutableContext, url string) (string, error) {
	wsURL := httpx.MakeURL(sc, url, map[string]any{
		"t": time.Now().UnixMilli(),
	}, true)
	if wsURL == "" {
		return "", errs.NewZCA("build websocket URL failed", "listener.buildWebSocketURL")
	}
	return wsURL, nil
}

func buildRetryStates(sc session.MutableContext) map[string]*retryState {
	retryStates := make(map[string]*retryState, 8)
	if s := sc.Settings(); s != nil && s.Features.Socket.Retries != nil {
		for reason, cfg := range s.Features.Socket.Retries {
			max := 0
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
				times: append([]int(nil), times...),
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
		Message:      make(chan model.Message, buf.Message),
		// OldMessages: make(chan OldMessagesEvent, buf.OldMessages),
		// Reaction: make(chan Reaction, buf.Reaction),
		// OldReactions: make(chan OldReactionsEvent, buf.OldReactions),
		// Typing: make(chan Typing, buf.Typing),
		// DeliveredMessages: make(chan []DeliveredMessage, buf.DeliveredMessages),
		// SeenMessages: make(chan []SeenMessage, buf.SeenMessages),
		Undo:             make(chan model.Undo, buf.Undo),
		UploadAttachment: make(chan model.UploadAttachment, buf.UploadAttachment),
		// Friend: make(chan FriendEvent, buf.Friend),
		// Group: make(chan GroupEvent, buf.Group),
		CipherKey: make(chan string, buf.CipherKey),
	}
}
