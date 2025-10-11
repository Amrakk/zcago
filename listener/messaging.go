package listener

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/websocketx"
	"github.com/Amrakk/zcago/model"
)

type WSPayload struct {
	Version uint8
	CMD     uint16
	SubCMD  uint8
	Data    map[string]any
}

type WSMessage[T any] struct {
	Key          *string `json:"key"`
	Encrypt      uint    `json:"encrypt"`
	ErrorCode    int     `json:"error_code"`
	ErrorMessage string  `json:"error_message"`
	Data         T       `json:"data"`
}

type BaseWSMessage = WSMessage[string]

// ----------------------------------------
// WebSocket sending utilities
// ----------------------------------------

func (ln *listener) SendWS(ctx context.Context, p WSPayload, requireID bool) error {
	if err := ln.validateSendRequest(ctx); err != nil {
		return err
	}

	client := ln.getClient()
	if client == nil {
		return errs.NewZCA("listener not started", "listener.SendWS")
	}

	if requireID {
		ln.addRequestID(&p)
	}

	frame, err := encodeFrame(p)
	if err != nil {
		return errs.WrapZCA("failed to encode frame", "listener.SendWS", err)
	}

	client.Write(ctx, websocketx.BinaryMessage, frame)

	return nil
}

func (ln *listener) RequestOldMessages(ctx context.Context, tt model.ThreadType, lastMsgID *string) error {
	cmd := uint16(510)
	if tt == model.ThreadTypeUser {
		cmd = 511
	}
	data := map[string]any{
		"first":  true,
		"lastId": lastMsgID,
		"preIds": []string{},
	}

	return ln.SendWS(ctx, WSPayload{
		Version: 1,
		CMD:     cmd,
		SubCMD:  1,
		Data:    data,
	}, true)
}

func (ln *listener) RequestOldReactions(ctx context.Context, tt model.ThreadType, lastMsgID *string) error {
	cmd := uint16(610)
	if tt == model.ThreadTypeUser {
		cmd = 611
	}
	data := map[string]any{
		"first":  true,
		"lastId": lastMsgID,
		"preIds": []string{},
	}

	return ln.SendWS(ctx, WSPayload{
		Version: 1,
		CMD:     cmd,
		SubCMD:  1,
		Data:    data,
	}, true)
}

func (ln *listener) validateSendRequest(ctx context.Context) error {
	if ln == nil {
		return errs.NewZCA("listener is nil", "listener.validateSendRequest")
	}
	if ctx == nil {
		return errs.NewZCA("context is nil", "listener.validateSendRequest")
	}
	if ctx.Err() != nil {
		err := ctx.Err()
		return errs.WrapZCA("context cancelled", "listener.validateSendRequest", err)
	}
	return nil
}

func (ln *listener) addRequestID(p *WSPayload) {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if p.Data == nil {
		p.Data = map[string]any{}
	}
	p.Data["req_id"] = "req_" + fmt.Sprint(ln.reqID)
	ln.reqID++
}

// ----------------------------------------
// Websocket reading utilities
// ----------------------------------------

func (ln *listener) handleWebSocketMessage(ctx context.Context, msg websocketx.Message) {
	if msg.Type != websocketx.BinaryMessage {
		return
	}

	version, cmd, subCMD, data, err := parseWebSocketMessage(msg.Data)
	if err != nil {
		ln.emitError(ctx, err)
		return
	}

	var parsed BaseWSMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		ln.emitError(ctx, errs.WrapZCA("failed to parse message JSON", "listener.handleWebSocketMessage", err))
		return
	}

	ln.router(ctx, uint(version), uint(cmd), uint(subCMD), parsed)
}

func parseWebSocketMessage(data []byte) (byte, uint16, byte, []byte, error) {
	if len(data) < 4 {
		return 0, 0, 0, nil, errs.NewZCA("message too short", "listener.parseWebSocketMessage")
	}

	header := make([]byte, 4)
	copy(header, data[:4])

	version, cmd, subCMD, err := getMessageHeader(header)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	msgData := data[4:]
	if len(msgData) == 0 {
		return 0, 0, 0, nil, errs.NewZCA("empty message data", "listener.parseWebSocketMessage")
	}

	return version, cmd, subCMD, msgData, nil
}

func getMessageHeader(buffer []byte) (byte, uint16, byte, error) {
	if len(buffer) < 4 {
		return 0, 0, 0, errs.NewZCA("invalid header", "listener.getHeader")
	}

	version := buffer[0]
	cmd := binary.LittleEndian.Uint16(buffer[1:3])
	subCMD := buffer[3]

	return version, cmd, subCMD, nil
}
