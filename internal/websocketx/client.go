package websocketx

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
)

type Client interface {
	Messages() <-chan Message
	Errors() <-chan error
	Closed() <-chan CloseInfo

	Write(typ int, data []byte)
	WriteText(s string)

	Close(code int, reason string)
}

type Message struct {
	Type int
	Data []byte
}

type CloseInfo struct {
	Code   int
	Reason string
	Err    error
}

type client struct {
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once

	msgChan    chan Message
	errChan    chan error
	closedChan chan CloseInfo
	writeQueue chan Message
}

var _ Client = (*client)(nil)

func Dial(ctx context.Context, url string, opt *Options) (*client, error) {
	cfg := defaultOptions()
	if opt != nil {
		if opt.Proxy != nil {
			cfg.Proxy = opt.Proxy
		}
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

	dialer := websocket.Dialer{
		Proxy:             cfg.Proxy,
		EnableCompression: true,
	}

	ws, _, err := dialer.Dial(url, cfg.Header)
	if err != nil {
		return nil, err
	}

	ws.SetReadLimit(-1)

	cctx, cancel := context.WithCancel(ctx)
	cl := &client{
		conn:       ws,
		ctx:        cctx,
		cancel:     cancel,
		msgChan:    make(chan Message, cfg.MsgBuf),
		errChan:    make(chan error, cfg.ErrBuf),
		closedChan: make(chan CloseInfo, 1),
		writeQueue: make(chan Message, cfg.WriteBuf),
	}

	// Reader
	cl.wg.Add(1)
	go cl.readLoop()

	// Writer
	cl.wg.Add(1)
	go cl.writeLoop()

	return cl, nil
}

func (c *client) Messages() <-chan Message { return c.msgChan }
func (c *client) Errors() <-chan error     { return c.errChan }
func (c *client) Closed() <-chan CloseInfo { return c.closedChan }

func (c *client) Write(typ int, data []byte) {
	select {
	case c.writeQueue <- Message{Type: typ, Data: data}:
	case <-c.ctx.Done():
		c.handleErr(context.Canceled)
	}
}

func (c *client) WriteText(s string) {
	c.Write(websocket.TextMessage, []byte(s))
}

func (c *client) Close(code int, reason string) {
	c.shutdown(CloseInfo{Code: code, Reason: reason}, true)
}

func (c *client) readLoop() {
	defer c.wg.Done()
	for {
		typ, data, err := c.conn.ReadMessage()
		if err != nil {
			c.shutdown(closeInfoFromErr(err), false)
			return
		}
		switch typ {
		case websocket.TextMessage, websocket.BinaryMessage:
			c.handleMsg(Message{Type: typ, Data: data})
		}
	}
}

func (c *client) writeLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.writeQueue:
			if err := c.conn.WriteMessage(msg.Type, msg.Data); err != nil {
				if isFatalWriteErr(err) {
					c.shutdown(closeInfoFromErr(err), false)
					return
				}
				c.handleErr(err)
			}
		}
	}
}

func (c *client) shutdown(ci CloseInfo, sendCloseFrame bool) {
	c.once.Do(func() {
		c.pushClose(ci)
		if sendCloseFrame {
			_ = c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(ci.Code, ci.Reason), time.Now().Add(2*time.Second))
		}
		_ = c.conn.Close()
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
