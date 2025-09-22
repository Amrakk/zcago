package zcago

import (
	"github.com/Amrakk/zcago/session"
	"github.com/Amrakk/zcago/session/auth"
)

type (
	Option session.Option

	LoginQROption   auth.LoginQROption
	LoginQRCallback auth.LoginQRCallback
)

func toSessionOptions(opts ...Option) []session.Option {
	sopts := make([]session.Option, len(opts))
	for i, o := range opts {
		sopts[i] = session.Option(o)
	}
	return sopts
}
