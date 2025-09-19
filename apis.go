package zcago

import (
	"context"

	"github.com/Amrakk/zcago/session"
)

type API interface {
	GetContext(ctx context.Context) (*session.Context, error)
	// FetchAccountInfo(ctx context.Context) (*somewhere.FetchAccountInfoResponse, error)
}

type api struct {
	Context *session.Context
}

func NewAPI(ctx *session.Context) API {
	return &api{
		Context: ctx,
	}
}

func (a *api) GetContext(ctx context.Context) (*session.Context, error) {
	return a.Context, nil
}

// func (a *api) FetchAccountInfo(ctx context.Context) (*somewhere.FetchAccountInfoResponse, error) {
// 	return somewhere.FetchAccountInfo(ctx)
// }
