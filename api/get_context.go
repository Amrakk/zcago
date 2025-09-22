package api

import (
	"context"

	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/session"
)

func (a *api) GetContext(ctx context.Context) (session.Context, error) {
	if a.ctx == nil {
		return nil, errs.NewZCAError("API context is not initialized", "GetContext", nil)
	}

	return a.ctx, nil
}
