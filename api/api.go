package api

import (
	"github.com/Amrakk/zcago/session"
)

type api struct {
	ctx session.MutableContext
}

func New(ctx session.MutableContext) *api {
	return &api{
		ctx: ctx,
	}
}
