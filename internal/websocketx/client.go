package websocketx

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	TextMessage   = websocket.MessageText
	BinaryMessage = websocket.MessageBinary
)

type Client interface {
	Messages() <-chan Message
	Errors() <-chan error
	Closed() <-chan CloseInfo

	Write(ctx context.Context, typ websocket.MessageType, data []byte) error
	WriteText(ctx context.Context, s string) error

	Close(code int, reason string)
}

type Message struct {
	Type websocket.MessageType
	Data []byte
}

type CloseInfo struct {
	Code   int
	Reason string
	Err    error
}

type writeRequest struct {
	ctx    context.Context
	typ    websocket.MessageType
	data   []byte
	result chan error
}

type client struct {
	conn    *websocket.Conn
	connCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once

	msgChan    chan Message
	errChan    chan error
	closedChan chan CloseInfo
	writeQueue chan writeRequest
}

var _ Client = (*client)(nil)

func Dial(ctx context.Context, url string, opt *Options) (*client, error) {
	cfg := defaultOptions()
	if opt != nil {
		if opt.Header != nil {
			cfg.Header = opt.Header
		}
		if opt.MsgBuf > 0 {
			cfg.MsgBuf = opt.MsgBuf
		}
		if opt.ErrBuf > 0 {
			cfg.ErrBuf = opt.ErrBuf
		}
		if opt.WriteBuf > 0 {
			cfg.WriteBuf = opt.WriteBuf
		}
	}

	dialOpts := &websocket.DialOptions{
		HTTPHeader: cfg.Header,
	}

	if cfg.HTTPClient != nil {
		dialOpts.HTTPClient = cfg.HTTPClient
	}

	conn, _, err := websocket.Dial(ctx, url, dialOpts)
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithCancel(ctx)
	cl := &client{
		conn:       conn,
		connCtx:    cctx,
		cancel:     cancel,
		msgChan:    make(chan Message, cfg.MsgBuf),
		errChan:    make(chan error, cfg.ErrBuf),
		closedChan: make(chan CloseInfo, 1),
		writeQueue: make(chan writeRequest, cfg.WriteBuf),
	}

	// Reader
	cl.wg.Add(1)
	go cl.readLoop(cctx)

	// Writer
	cl.wg.Add(1)
	go cl.writeLoop(cctx)

	return cl, nil
}

func (c *client) Messages() <-chan Message { return c.msgChan }
func (c *client) Errors() <-chan error     { return c.errChan }
func (c *client) Closed() <-chan CloseInfo { return c.closedChan }

func (c *client) Write(ctx context.Context, typ websocket.MessageType, data []byte) error {
	req := writeRequest{
		ctx:    ctx,
		typ:    typ,
		data:   data,
		result: make(chan error, 1),
	}

	select {
	case c.writeQueue <- req:
		select {
		case err := <-req.result:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *client) WriteText(ctx context.Context, s string) error {
	return c.Write(ctx, websocket.MessageText, []byte(s))
}

func (c *client) Close(code int, reason string) {
	c.shutdown(CloseInfo{Code: code, Reason: reason}, true)
}

func (c *client) readLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		typ, data, err := c.conn.Read(ctx)
		if err != nil {
			if c.isFatalErr(err) {
				c.shutdown(closeInfoFromErr(err), false)
				return
			} else {
				c.handleErr(err)
				continue
			}
		}

		switch typ {
		case websocket.MessageText, websocket.MessageBinary:
			c.handleMsg(Message{Type: typ, Data: data})
		}
	}
}

func (c *client) writeLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-c.writeQueue:
			err := c.conn.Write(req.ctx, req.typ, req.data)

			select {
			case req.result <- err:
			default:
			}

			if err != nil && c.isFatalErr(err) {
				c.shutdown(closeInfoFromErr(err), false)
				return
			}
		}
	}
}

func (c *client) shutdown(ci CloseInfo, sendCloseFrame bool) {
	c.once.Do(func() {
		c.pushClose(ci)
		if sendCloseFrame {
			_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = c.conn.Close(websocket.StatusCode(ci.Code), ci.Reason)
		} else {
			_ = c.conn.CloseNow()
		}
		c.cancel()
		go func() {
			c.wg.Wait()
			close(c.msgChan)
			close(c.errChan)
			close(c.closedChan)
		}()
	})
}

func (c *client) handleMsg(m Message) {
	select {
	case c.msgChan <- m:
	default:
		select { // drop oldest
		case <-c.msgChan:
		default:
		}
		select { // retry once, non-blocking
		case c.msgChan <- m:
		default:
		}
	}
}

func (c *client) handleErr(err error) {
	select {
	case c.errChan <- err:
	default:
		select { // drop oldest
		case <-c.errChan:
		default:
		}
		select { // retry once, non-blocking
		case c.errChan <- err:
		default:
		}
	}
}

func (c *client) pushClose(ci CloseInfo) {
	for {
		select {
		case c.closedChan <- ci:
			return
		default:
			select {
			case <-c.closedChan:
			default:
			}
		}
	}
}
