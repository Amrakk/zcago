package zcago

import (
	"net/http"

	"github.com/Amrakk/zcago/session"
	"github.com/Amrakk/zcago/session/auth"
)

type (
	LoginQROption   auth.LoginQROption
	LoginQRCallback auth.LoginQRCallback
)

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
