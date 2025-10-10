package listener

import (
	"context"

	"github.com/Amrakk/zcago/internal/websocketx"
	"github.com/Amrakk/zcago/model"
)

type Buffers struct {
	Connected    int
	Disconnected int
	Closed       int
	Error        int
	// Typing           int
	Message int
	// OldMessage       int
	// SeenMessage      int
	// DeliveredMessage int
	// Reaction         int
	// OldReaction      int
	// UploadAttachment int
	Undo int
	// Friend           int
	// Group            int
	CipherKey int
}

func defaultBuffers() Buffers {
	return Buffers{Connected: 1, Disconnected: 4, Closed: 1, Error: 16, Message: 128, Undo: 32, CipherKey: 4}
}

type channels struct {
	Connected    chan struct{}
	Disconnected chan websocketx.CloseInfo
	Closed       chan websocketx.CloseInfo
	Error        chan error
	// Typing           chan Typing
	Message chan model.UserMessage
	// OldMessage       chan OldMessages
	// SeenMessage      chan []SeenMessage
	// DeliveredMessage chan []DeliveredMessage
	// Reaction         chan Reaction
	// OldReaction      chan OldReactions
	// UploadAttachment chan UploadAttachment
	Undo chan model.Undo
	// Friend           chan Friend
	// Group            chan Group
	CipherKey chan string
}

func (ln *listener) Connected() <-chan struct{}                { return ln.ch.Connected }
func (ln *listener) Disconnected() <-chan websocketx.CloseInfo { return ln.ch.Disconnected }
func (ln *listener) Closed() <-chan websocketx.CloseInfo       { return ln.ch.Closed }
func (ln *listener) Error() <-chan error                       { return ln.ch.Error }

// func (ln *listener) Typing() <-chan Typing                       { return ln.ch.Typing }
func (ln *listener) Message() <-chan model.UserMessage { return ln.ch.Message }

// func (ln *listener) OldMessage() <-chan OldMessages              { return ln.ch.OldMessage }
// func (ln *listener) SeenMessage() <-chan []SeenMessage           { return ln.ch.SeenMessage }
// func (ln *listener) DeliveredMessage() <-chan []DeliveredMessage { return ln.ch.DeliveredMessage }
// func (ln *listener) Reaction() <-chan Reaction                   { return ln.ch.Reaction }
// func (ln *listener) OldReaction() <-chan OldReactions            { return ln.ch.OldReaction }
// func (ln *listener) UploadData() <-chan UploadData               { return ln.ch.UploadData }
func (ln *listener) Undo() <-chan model.Undo { return ln.ch.Undo }

// func (ln *listener) Friend() <-chan Friend                       { return ln.ch.Friend }
// func (ln *listener) Group() <-chan Group                         { return ln.ch.Group }
func (ln *listener) CipherKey() <-chan string { return ln.ch.CipherKey }

func (ln *listener) emitError(ctx context.Context, err error) {
	select {
	case <-ctx.Done():
		return
	case ln.ch.Error <- err:
	default:
	}
}

func (ln *listener) emitClosed(ctx context.Context, ci websocketx.CloseInfo) {
	select {
	case ln.ch.Closed <- ci:
	case <-ctx.Done():
		return
	default:
	}
}

// emit attempts to deliver obj into the provided channel without blocking.
//
// Behavior:
//   - If the channel has buffer space, obj is sent immediately.
//   - If the channel is full, the oldest value in the channel is dropped
//     (non-blocking receive) and emit retries once to enqueue obj.
//   - If the second attempt also fails (channel still full), obj is dropped.
//
// This policy ensures that slow or absent receivers do not block the sender,
// at the cost of possibly overwriting or dropping messages.
func emit[T any](ctx context.Context, ch chan T, obj T) {
	select {
	case <-ctx.Done():
		return
	case ch <- obj:
	default:
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- obj:
		default:
		}
	}
}
