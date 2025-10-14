package zcago

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/session"
	"github.com/Amrakk/zcago/session/auth"
	"github.com/Amrakk/zcago/version"
)

type Zalo interface {
	Login(ctx context.Context, cred Credentials) (API, error)
	LoginQR(ctx context.Context, opt *LoginQROption, cb LoginQRCallback) (API, error)
}

type zalo struct {
	enableEncryptParam bool
	opts               []session.Option
}

func NewZalo(opts ...session.Option) Zalo {
	z := &zalo{
		enableEncryptParam: true,
		opts:               opts,
	}

	return z
}

// Login authenticates using pre-saved credentials containing cookies from a previous session.
func (z *zalo) Login(ctx context.Context, cred Credentials) (API, error) {
	sc := session.NewContext(z.opts...)

	return z.loginCookie(ctx, sc, cred)
}

// LoginQR performs interactive QR code authentication.
func (z *zalo) LoginQR(ctx context.Context, opts *LoginQROption, cb LoginQRCallback) (API, error) {
	const defaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0"

	options := LoginQROption{
		UserAgent: defaultUA,
		Language:  "vi",
	}
	if opts != nil {
		if opts.UserAgent != "" {
			options.UserAgent = opts.UserAgent
		}
		if opts.Language != "" {
			options.Language = opts.Language
		}
		options.QRPath = opts.QRPath
	}

	sc := session.NewContext(z.opts...)
	sc.SetUserAgent(options.UserAgent)

	res, err := loginQR(ctx, sc, options.QRPath, cb)
	if err != nil {
		logger.Log(sc).Error(err)
		return nil, err
	} else if res == nil {
		return nil, errs.WrapZCA("unable to login with QR code", "zalo.LoginQR", err)
	}

	imei := generateZaloUUID(options.UserAgent)

	if cb != nil {
		cb(auth.EventGotLoginInfo{
			Data: auth.QRGotLoginInfoData{
				IMEI:      imei,
				UserAgent: options.UserAgent,
			},
		})
	}

	cred := Credentials{
		IMEI:      imei,
		UserAgent: options.UserAgent,
		Language:  &options.Language,
		// Cookie field omitted - preserves existing session cookies from QR login
	}

	return z.loginCookie(ctx, sc, cred)
}

func (z *zalo) loginCookie(ctx context.Context, sc session.MutableContext, cred Credentials) (API, error) {
	if ok := cred.IsValid(); !ok {
		return nil, errs.NewZCA("invalid credentials", "zalo.loginCookie")
	}

	lang := "vi"
	if cred.Language != nil && *cred.Language != "" {
		lang = *cred.Language
	}
	sc.SetIMEI(cred.IMEI)
	sc.SetLanguage(lang)
	sc.SetUserAgent(cred.UserAgent)

	// Apply saved cookies if provided
	// Skip if invalid/empty
	if cred.Cookie != nil && cred.Cookie.IsValid() {
		u := url.URL{Scheme: "https", Host: "chat.zalo.me"}
		jar := cred.Cookie.BuildCookieJar(&u)
		sc.SetCookieJar(jar)
	}

	var (
		loginInfo  *session.LoginInfo
		serverInfo *session.ServerInfo
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer func() { _ = recover() }()
		version.CheckUpdate(gctx, sc)
		return nil
	})

	g.Go(func() error {
		li, err := login(gctx, sc, z.enableEncryptParam)
		if err != nil {
			logger.Log(sc).Error("Login failed", err)
			return err
		}
		loginInfo = li
		return nil
	})

	g.Go(func() error {
		si, err := getServerInfo(gctx, sc, z.enableEncryptParam)
		if err != nil {
			logger.Log(sc).Error("Failed to get server info:", err)
			return err
		}
		serverInfo = si
		return nil
	})

	if err := g.Wait(); err != nil || loginInfo == nil || serverInfo == nil {
		if err != nil {
			return nil, err
		}
		logger.Log(sc).Error("Login or server info is empty")
		return nil, errs.NewZCA("Login failed", "zalo.loginCookie")
	}

	secretKey := session.SecretKey(loginInfo.ZPWEnk)

	sc.SealLogin(session.Seal{
		UID:       loginInfo.UID,
		IMEI:      cred.IMEI,
		UserAgent: cred.UserAgent,
		Language:  lang,
		SecretKey: secretKey,
		LoginInfo: loginInfo,
		Settings:  serverInfo.Settings,
		ExtraVer:  serverInfo.ExtraVer,
	})

	logger.Log(sc).Success("Successfully logged in as ", sc.UID())
	return api.New(sc)
}

func generateZaloUUID(userAgent string) string {
	u := uuid.New().String()
	hash := md5.Sum([]byte(userAgent))
	return u + "-" + hex.EncodeToString(hash[:])
}
