package zcago

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/session"
	"github.com/Amrakk/zcago/session/auth"
)

type Zalo interface {
	Login(ctx context.Context, cred Credentials) (API, error)
	LoginQR(ctx context.Context, opt *LoginQROption, cb *LoginQRCallback) (API, error)
}

type zalo struct {
	EnableEncryptParam bool
	Options            session.Options
}

func NewZalo(opts *session.Options) Zalo {
	z := &zalo{
		EnableEncryptParam: true,
		Options:            session.ApplyOptions(opts),
	}

	return z
}

func (z *zalo) Login(ctx context.Context, cred Credentials) (API, error) {
	appCtx := session.NewContext(z.Options.APIType, z.Options.APIVersion)
	appCtx.Options = z.Options

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

func (z *zalo) loginCookie(ctx context.Context, sc *session.Context, cred Credentials) (API, error) {
	z.validateParams(cred)

	sc.IMEI = cred.Imei
	sc.Cookie = z.parseCookies(cred.Cookie)
	sc.UserAgent = cred.UserAgent

	if cred.Language != nil {
		sc.Language = *cred.Language
	} else {
		sc.Language = "vi"
	}

	loginInfo, err := auth.Login(ctx, sc, z.EnableEncryptParam)
	if err != nil {
		httpx.Logger(sc).Error("Login failed", err)
		return nil, err
	}
	serverInfo, err := auth.GetServerInfo(ctx, sc, z.EnableEncryptParam)
	if err != nil {
		return nil, err
	}

	if loginInfo == nil || serverInfo == nil {
		return nil, errs.NewZCAError("login failed", "Login", nil)
	}

	sc.SecretKey = loginInfo.ZPWEnk
	sc.UID = loginInfo.UID

	// if settings, ok := serverInfo["settings"].(map[string]any); ok {
	// 	sc.Settings = settings
	// } else if settings, ok := serverInfo["setttings"].(map[string]any); ok {
	// 	sc.Settings = settings
	// } else {
	// 	return nil, errs.NewZCAError("missing settings", "Login", nil)
	// }

	// if extraVer, ok := serverInfo["extra_ver"].(string); ok {
	// 	sc.ExtraVer = extraVer
	// }

	sc.LoginInfo = loginInfo

	httpx.Logger(sc).Info("Logged in as ", loginInfo.UID)

	return NewAPI(sc), nil
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
