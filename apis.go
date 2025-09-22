package zcago

import (
	"context"

	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/session"
)

type API interface {
	GetContext(ctx context.Context) (session.Context, error)
	FetchAccountInfo(ctx context.Context) (api.FetchAccountInfoResponse, error)
}
