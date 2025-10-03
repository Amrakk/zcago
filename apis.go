package zcago

import (
	"context"

	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/listener"
	"github.com/Amrakk/zcago/session"
)

type API interface {
	GetContext() (session.Context, error)
	GetOwnID() string
	Listener() listener.Listener

	//gen:methods

	// GetUserInfo returns the profile for userID.
	//
	// Params:
	//   ctx    — cancel/deadline control
	//   userID — target identifier
	//
	// Errors: errs.ZaloAPIError
	FetchAccountInfo(ctx context.Context) (*api.FetchAccountInfoResponse, error)
	GetUserInfo(ctx context.Context, userID ...string) (*api.GetUserInfoResponse, error)
	// UpdateLanguage sets the user’s language.
	//
	// Note: Calling this endpoint alone will not update the user’s language.
	//
	// Params:
	//   - ctx  — cancel/deadline control
	//   - lang — target language ("VI", "EN")
	//
	// Errors: errs.ZaloAPIError
	UpdateLanguage(ctx context.Context, lang api.Language) (api.UpdateLanguageResponse, error)
}
