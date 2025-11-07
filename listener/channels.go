package listener

import (
	"context"

	"github.com/Amrakk/zcago/internal/websocketx"
	"github.com/Amrakk/zcago/model"
)

type Buffers struct {
	Connected         int
	Disconnected      int
	Closed            int
	Error             int
	Message           int
	OldMessages       int
	Typing            int
	SeenMessages      int
	DeliveredMessages int
	Reaction          int
	OldReactions      int
	Undo              int
	UploadAttachment  int
	Friend            int
	Group             int
	CipherKey         int
}

func defaultBuffers() Buffers {
	return Buffers{
		Connected:         1,
		Disconnected:      4,
		Closed:            1,
		Error:             16,
		Message:           64,
		OldMessages:       64,
		Typing:            8,
		SeenMessages:      16,
		DeliveredMessages: 16,
		Reaction:          64,
		OldReactions:      64,
		Undo:              32,
		UploadAttachment:  32,
		Friend:            32,
		Group:             32,
		CipherKey:         4,
	}
}

type channels struct {
	Connected         chan struct{}
	Disconnected      chan websocketx.CloseInfo
	Closed            chan websocketx.CloseInfo
	Error             chan error
	Message           chan model.Message
	OldMessages       chan model.OldMessages
	Reaction          chan model.Reaction
	OldReactions      chan model.OldReactions
	Typing            chan model.Typing
	DeliveredMessages chan []model.DeliveredMessage
	SeenMessages      chan []model.SeenMessage
	Undo              chan model.Undo
	UploadAttachment  chan model.UploadAttachment
	Friend            chan model.FriendEvent
	Group             chan model.GroupEvent
	CipherKey         chan string
}

func (ln *listener) Connected() <-chan struct{}                { return ln.ch.Connected }
func (ln *listener) Disconnected() <-chan websocketx.CloseInfo { return ln.ch.Disconnected }
func (ln *listener) Closed() <-chan websocketx.CloseInfo       { return ln.ch.Closed }
func (ln *listener) Error() <-chan error                       { return ln.ch.Error }
func (ln *listener) Message() <-chan model.Message             { return ln.ch.Message }
func (ln *listener) OldMessages() <-chan model.OldMessages     { return ln.ch.OldMessages }
func (ln *listener) Reaction() <-chan model.Reaction           { return ln.ch.Reaction }
func (ln *listener) OldReactions() <-chan model.OldReactions   { return ln.ch.OldReactions }
func (ln *listener) Typing() <-chan model.Typing               { return ln.ch.Typing }

func (ln *listener) DeliveredMessages() <-chan []model.DeliveredMessage {
	return ln.ch.DeliveredMessages
}
func (ln *listener) SeenMessages() <-chan []model.SeenMessage        { return ln.ch.SeenMessages }
func (ln *listener) Undo() <-chan model.Undo                         { return ln.ch.Undo }
func (ln *listener) UploadAttachment() <-chan model.UploadAttachment { return ln.ch.UploadAttachment }
func (ln *listener) Friend() <-chan model.FriendEvent                { return ln.ch.Friend }
func (ln *listener) Group() <-chan model.GroupEvent                  { return ln.ch.Group }
func (ln *listener) CipherKey() <-chan string                        { return ln.ch.CipherKey }

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
