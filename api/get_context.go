package api

import (
	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/session"
)

func (a *api) GetContext() (session.Context, error) {
	if a.sc == nil {
		return nil, errs.NewZCA("API context is not initialized", "api.GetContext")
	}

	return a.sc, nil
}
