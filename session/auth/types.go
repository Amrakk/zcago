package auth

import (
	"context"
)

type LoginQREventType int

const (
	LoginQREventGenerated LoginQREventType = iota
	LoginQREventExpired
	LoginQREventScanned
	LoginQREventDeclined
	LoginQREventGotLoginInfo
)

// ----------------------------------------
// Events
// ----------------------------------------

type LoginQREvent interface {
	Type() LoginQREventType
}

// QRCodeGenerated
type EventQRCodeGenerated struct {
	Data    QRGeneratedData
	Actions GeneratedActions
}

func (EventQRCodeGenerated) Type() LoginQREventType { return LoginQREventGenerated }

// QRCodeExpired
type EventQRCodeExpired struct {
	Actions CommonActions
}

func (EventQRCodeExpired) Type() LoginQREventType { return LoginQREventExpired }

// QRCodeScanned
type EventQRCodeScanned struct {
	Data    QRScannedData
	Actions CommonActions
}

func (EventQRCodeScanned) Type() LoginQREventType { return LoginQREventScanned }

// QRCodeDeclined
type EventQRCodeDeclined struct {
	Data    QRDeclinedData
	Actions CommonActions
}

func (EventQRCodeDeclined) Type() LoginQREventType { return LoginQREventDeclined }

// GotLoginInfo
type EventGotLoginInfo struct {
	Data QRGotLoginInfoData
}

func (EventGotLoginInfo) Type() LoginQREventType { return LoginQREventGotLoginInfo }

// ----------------------------------------
// Actions
// ----------------------------------------

type CommonActions interface {
	Retry(ctx context.Context) error
	Abort(ctx context.Context) error
}

type GeneratedActions interface {
	Retry(ctx context.Context) error
	Abort(ctx context.Context) error
	SaveToFile(ctx context.Context, path string) error
}

type qrController struct {
	retryFn      func(context.Context) error
	abortFn      func(context.Context) error
	saveToFileFn func(context.Context, string) error
}

func (c qrController) Retry(ctx context.Context) error {
	if c.retryFn != nil {
		return c.retryFn(ctx)
	}
	return nil
}

func (c qrController) Abort(ctx context.Context) error {
	if c.abortFn != nil {
		return c.abortFn(ctx)
	}
	return nil
}

func (c qrController) SaveToFile(ctx context.Context, p string) error {
	if c.saveToFileFn != nil {
		return c.saveToFileFn(ctx, p)
	}
	return nil
}

type commonActions struct{ qrController }

// ----------------------------------------
// Data payloads
// ----------------------------------------

type QRGeneratedOptions struct {
	EnabledCheckOCR   bool `json:"enabledCheckOCR"`
	EnabledMultiLayer bool `json:"enabledMultiLayer"`
}

type QRGeneratedData struct {
	Code    string             `json:"code"`
	Image   string             `json:"image"` // base64-encoded image data
	Options QRGeneratedOptions `json:"options"`
}

type QRScannedData struct {
	Avatar      string `json:"avatar"`
	DisplayName string `json:"display_name"`
}

type QRDeclinedData struct {
	Code string
}

type QRGotLoginInfoData struct {
	IMEI      string
	UserAgent string
}

type UserInfo struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type QRUserInfo struct {
	Logged           bool     `json:"logged"`
	SessionChatValid bool     `json:"sessionChatValid"`
	Info             UserInfo `json:"info"`
}
