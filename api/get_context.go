package api

import (
	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/session"
)

func (a *api) GetContext() (session.Context, error) {
	if a.ctx == nil {
		return nil, errs.NewZCAError("API context is not initialized", "GetContext", nil)
	}

	return a.ctx, nil
}
