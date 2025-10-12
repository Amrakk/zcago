package zcago

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/session"
	"github.com/Amrakk/zcago/session/auth"
)

type (
	LoginQREvent    auth.LoginQREvent
	LoginQRCallback auth.LoginQRCallback
)

type LoginQROption struct {
	UserAgent string
	Language  string
	QRPath    string
}

func login(ctx context.Context, sc session.MutableContext, encryptParams bool) (*session.LoginInfo, error) {
	return auth.Login(ctx, sc, encryptParams)
}

func getServerInfo(ctx context.Context, sc session.MutableContext, enableEncryptParam bool) (*session.ServerInfo, error) {
	return auth.GetServerInfo(ctx, sc, enableEncryptParam)
}

func loginQR(ctx context.Context, sc session.MutableContext, qrPath string, cb LoginQRCallback) (*auth.LoginQRResult, error) {
	return auth.LoginQR(ctx, sc, qrPath, auth.LoginQRCallback(cb))
}

func WithSelfListen(v bool) session.Option  { return session.WithSelfListen(v) }
func WithCheckUpdate(v bool) session.Option { return session.WithCheckUpdate(v) }
func WithLogging(v bool) session.Option     { return session.WithLogging(v) }

// WithLogLevel sets the logging verbosity for a session.
//
// Accepted values:
//
//	0 — Verbose
//	1 — Debug
//	2 — Info
//	3 — Warn
//	4 — Error
//	5 — Success
func WithLogLevel(level uint8) session.Option      { return session.WithLogLevel(level) }
func WithHTTPClient(c *http.Client) session.Option { return session.WithHTTPClient(c) }
func WithAPIType(t uint) session.Option            { return session.WithAPIType(t) }
func WithAPIVersion(v uint) session.Option         { return session.WithAPIVersion(v) }

func WithImageMetadataGetter(f session.ImageMetadataGetter) session.Option {
	return session.WithImageMetadataGetter(f)
}
