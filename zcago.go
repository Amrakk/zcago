package zcago

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/Amrakk/zcago/api"
	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/httpx"
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
	opts               []Option
}

func NewZalo(opts ...Option) Zalo {
	z := &zalo{
		enableEncryptParam: true,
		opts:               opts,
	}

	return z
}

func (z *zalo) Login(ctx context.Context, cred Credentials) (API, error) {
	appCtx := session.NewContext(toSessionOptions(z.opts...)...)

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
	version.CheckUpdate(sc)
	z.validateParams(cred)

	lang := "vi"
	if cred.Language != nil && *cred.Language != "" {
		lang = *cred.Language
	}

	sc.SetIMEI(cred.Imei)
	sc.SetLanguage(lang)
	sc.SetUserAgent(cred.UserAgent)
	sc.SetCookieJar(z.parseCookies(cred.Cookie))

	loginInfo, err := auth.Login(ctx, sc, z.enableEncryptParam)
	if err != nil {
		httpx.Logger(sc).Error("Login failed", err)
		return nil, err
	}
	serverInfo, err := auth.GetServerInfo(ctx, sc, z.enableEncryptParam)
	if err != nil {
		return nil, err
	}

	if loginInfo == nil || serverInfo == nil {
		return nil, errs.NewZCAError("login failed", "Login", nil)
	}

	sc.SealLogin(session.Seal{
		UID:       loginInfo.UID,
		IMEI:      cred.Imei,
		UserAgent: cred.UserAgent,
		Language:  lang,
		SecretKey: loginInfo.ZPWEnk,
		LoginInfo: loginInfo,
		Settings:  serverInfo.Settings,
		ExtraVer:  serverInfo.ExtraVer,
	})

	httpx.Logger(sc).Info("Logged in as ", sc.UID())

	return api.New(sc), nil
}

func (z *zalo) parseCookies(cookie CookieUnion) http.CookieJar {
	cookieArr := cookie.GetCookies()

	for i := range cookieArr {
		if len(cookieArr[i].Domain) > 0 && cookieArr[i].Domain[0] == '.' {
			cookieArr[i].Domain = cookieArr[i].Domain[1:]
		}
	}

	jar, _ := cookiejar.New(nil)
	cookies := make([]*http.Cookie, len(cookieArr))
	url := url.URL{Scheme: "https", Host: "chat.zalo.me"}

	for i, c := range cookieArr {
		cookies[i] = c.ToHTTPCookie()
	}

	jar.SetCookies(&url, cookies)
	return jar
}

func (z *zalo) validateParams(cred Credentials) error {
	if len(cred.Imei) == 0 || !cred.Cookie.IsValid() || len(cred.UserAgent) == 0 {
		return errs.NewZCAError("invalid credentials", "", nil)
	}
	return nil
}
