package api

import (
	"context"

	"github.com/Amrakk/zcago/model"
)

type FetchAccountInfoResponse = model.User

func (a *api) FetchAccountInfo(ctx context.Context) (FetchAccountInfoResponse, error) {
	panic("not implemented")
}
