package zcago

import (
	"context"
	"net/url"

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
	LoginQR(ctx context.Context, opt *LoginQROption, cb *LoginQRCallback) (API, error)
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

func (z *zalo) Login(ctx context.Context, cred Credentials) (API, error) {
	appCtx := session.NewContext(z.opts...)

	return z.loginCookie(ctx, appCtx, cred)
}

func (z *zalo) LoginQR(ctx context.Context, opt *LoginQROption, cb *LoginQRCallback) (API, error) {
	// if (!options) options = {};
	// if (!options.userAgent)
	// 	options.userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0";
	// if (!options.language) options.language = "vi";

	// const ctx = createContext(this.options.apiType, this.options.apiVersion);
	// Object.assign(ctx.options, this.options);

	// const loginQRResult = await loginQR(
	// 	ctx,
	// 	options as { userAgent: string; language: string; qrPath?: string },
	// 	callback,
	// );
	// if (!loginQRResult) throw new ZaloApiError("Unable to login with QRCode");

	// const imei = generateZaloUUID(options.userAgent);

	// if (callback) {
	// 	// Thanks to @YanCastle for this great suggestion!
	// 	callback({
	// 		type: LoginQRCallbackEventType.GotLoginInfo,
	// 		data: {
	// 			cookie: loginQRResult.cookies,
	// 			imei,
	// 			userAgent: options.userAgent,
	// 		},
	// 		actions: null,
	// 	});
	// }

	// return this.loginCookie(ctx, {
	// 	cookie: loginQRResult.cookies,
	// 	imei,
	// 	userAgent: options.userAgent,
	// 	language: options.language,
	// });
	panic("unimplemented")
}

func (z *zalo) loginCookie(ctx context.Context, sc session.MutableContext, cred Credentials) (API, error) {
	if ok := cred.IsValid(); !ok {
		return nil, errs.NewZCA("invalid credentials", "zalo.loginCookie")
	}

	lang := "vi"
	if cred.Language != nil && *cred.Language != "" {
		lang = *cred.Language
	}
	sc.SetIMEI(cred.Imei)
	sc.SetLanguage(lang)
	sc.SetUserAgent(cred.UserAgent)

	u := url.URL{Scheme: "https", Host: "chat.zalo.me"}
	jar := cred.Cookie.BuildCookieJar(&u)
	sc.SetCookieJar(jar)

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
		li, err := auth.Login(gctx, sc, z.enableEncryptParam)
		if err != nil {
			logger.Log(sc).Error("Login failed", err)
			return err
		}
		loginInfo = li
		return nil
	})

	g.Go(func() error {
		si, err := auth.GetServerInfo(gctx, sc, z.enableEncryptParam)
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
		IMEI:      cred.Imei,
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
