package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/internal/timex"
	"github.com/Amrakk/zcago/session"
)

var (
	ErrLoginQRAborted  = &errs.ZCAError{Message: "login QR aborted by user", Op: "auth.LoginQR"}
	ErrLoginQRDeclined = &errs.ZCAError{Message: "login QR declined by user", Op: "auth.LoginQR"}
)

type LoginQRResult struct {
	UserInfo UserInfo
}

type LoginQRCallback func(event LoginQREvent)

const defaultQRPath = "qr.png"

func LoginQR(ctx context.Context, sc session.MutableContext, qrPath string, cb LoginQRCallback) (*LoginQRResult, error) {
	for {
		setup := setupQRAttempt(ctx, qrPath, cb)

		stopTimeout := func() {}

		go func() {
			res, stopTimeoutFn, err := runQRAttempt(setup.attemptCtx, sc, setup.config)
			stopTimeout = stopTimeoutFn
			if err != nil {
				setup.errCh <- err
				return
			}
			setup.resultCh <- res
		}()

		result := handleAttemptResult(ctx, sc, setup)
		cleanupAttempt(stopTimeout, setup.cancelAttempt)

		if result.shouldRetry {
			ctx = result.newCtx
			continue
		}

		if result.err != nil {
			return nil, result.err
		}

		return result.result, nil
	}
}

type qrAttemptSetup struct {
	attemptCtx    context.Context
	cancelAttempt context.CancelFunc
	retryCh       chan context.Context
	abortCh       chan struct{}
	resultCh      chan *LoginQRResult
	errCh         chan error
	config        qrAttemptConfig
}

func setupQRAttempt(ctx context.Context, qrPath string, cb LoginQRCallback) *qrAttemptSetup {
	attemptCtx, cancelAttempt := context.WithCancel(ctx)

	retryCh := make(chan context.Context, 1)
	abortCh := make(chan struct{}, 1)
	resultCh := make(chan *LoginQRResult, 1)
	errCh := make(chan error, 1)

	ctrl := qrController{
		retryFn: func(ctx context.Context) error {
			select {
			case retryCh <- ctx:
			default:
				{
				}
			}
			return nil
		},
		abortFn: func(ctx context.Context) error {
			select {
			case abortCh <- struct{}{}:
			default:
				{
				}
			}
			return nil
		},
	}

	config := qrAttemptConfig{
		qrPath:  qrPath,
		cb:      cb,
		ctrl:    ctrl,
		retryCh: retryCh,
	}

	return &qrAttemptSetup{
		attemptCtx:    attemptCtx,
		cancelAttempt: cancelAttempt,
		retryCh:       retryCh,
		abortCh:       abortCh,
		resultCh:      resultCh,
		errCh:         errCh,
		config:        config,
	}
}

func cleanupAttempt(stopTimeout func(), cancelAttempt context.CancelFunc) {
	stopTimeout()
	cancelAttempt()
}

type attemptResult struct {
	result      *LoginQRResult
	err         error
	newCtx      context.Context
	shouldRetry bool
}

func handleAttemptResult(ctx context.Context, sc session.MutableContext, setup *qrAttemptSetup) attemptResult {
	select {
	case <-ctx.Done():
		return attemptResult{err: ctx.Err()}

	case newCtx := <-setup.retryCh:
		return attemptResult{newCtx: newCtx, shouldRetry: true}

	case <-setup.abortCh:
		return attemptResult{err: ErrLoginQRAborted}

	case err := <-setup.errCh:
		if setup.attemptCtx.Err() == nil {
			logger.Log(sc).Error(err)
		}
		return attemptResult{err: err}

	case res := <-setup.resultCh:
		return attemptResult{result: res}
	}
}

type qrAttemptConfig struct {
	qrPath  string
	cb      LoginQRCallback
	ctrl    qrController
	retryCh chan context.Context
}

func runQRAttempt(ctx context.Context, sc session.MutableContext, config qrAttemptConfig) (*LoginQRResult, func(), error) {
	qrPath := config.qrPath
	cb := config.cb
	ctrl := config.ctrl
	retryCh := config.retryCh

	actions := commonActions{ctrl}

	ver, err := loadLoginPage(ctx, sc)
	if err != nil {
		return nil, nil, errs.NewZCA("Cannot get API login version", "auth.LoginQR")
	}

	logger.Log(sc).Info("Login version: ", ver)

	getLoginInfo(ctx, sc, ver)
	verifyClient(ctx, sc, ver)

	jar := sc.CookieJar()
	fmt.Println(jar.Cookies(&url.URL{Host: "id.zalo.me", Scheme: "https"}))

	qrData, imgBytes, err := generateAndProcessQR(ctx, sc, ver)
	if err != nil {
		return nil, nil, err
	}

	if err := handleQRCallback(cb, ctrl, qrData, imgBytes, qrPath, sc); err != nil {
		return nil, nil, err
	}

	stopTimeout := setupQRTimeout(ctx, sc, cb, ctrl, retryCh)

	scanResult, err := waitingScan(ctx, sc, ver, qrData.Code)
	if err != nil {
		return nil, stopTimeout, errs.NewZCA("Cannot get scan result", "auth.LoginQR")
	}

	if cb != nil {
		cb(EventQRCodeScanned{
			Data:    scanResult.Data,
			Actions: actions,
		})
	}

	if err := processConfirmation(ctx, sc, ver, qrData.Code, cb, actions); err != nil {
		return nil, stopTimeout, err
	}

	userInfo, err := finalizeLogin(ctx, sc, scanResult.Data.DisplayName)
	if err != nil {
		return nil, stopTimeout, err
	}

	return &LoginQRResult{
		UserInfo: userInfo.Data.Info,
	}, stopTimeout, nil
}

func generateAndProcessQR(ctx context.Context, sc session.MutableContext, ver string) (QRGeneratedData, []byte, error) {
	qrGenResult, err := generateQRCode(ctx, sc, ver)
	if err != nil {
		return QRGeneratedData{}, nil, errs.WrapZCA("Unable to generate QRCode", "auth.LoginQR", err)
	}

	qrData := qrGenResult.Data
	b64 := strings.TrimPrefix(qrData.Image, "data:image/png;base64,")
	imgBytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return QRGeneratedData{}, nil, err
	}

	return qrData, imgBytes, nil
}

func handleQRCallback(cb LoginQRCallback, ctrl qrController, qrData QRGeneratedData, imgBytes []byte, qrPath string, sc session.MutableContext) error {
	if cb != nil {
		qrData.Image = string(imgBytes)
		cb(EventQRCodeGenerated{
			Data: qrData,
			Actions: qrController{
				retryFn: ctrl.retryFn,
				abortFn: ctrl.abortFn,
				saveToFileFn: func(c context.Context, p string) error {
					return saveQRCodeToFile(sc, p, imgBytes)
				},
			},
		})
	} else {
		if err := saveQRCodeToFile(sc, qrPath, imgBytes); err != nil {
			return err
		}
	}
	return nil
}

func setupQRTimeout(ctx context.Context, sc session.MutableContext, cb LoginQRCallback, ctrl qrController, retryCh chan context.Context) func() {
	return timex.SetTimeout(ctx, 100*time.Second, func() {
		logger.Log(sc).Info("QR expired!")

		if cb != nil {
			cb(EventQRCodeExpired{Actions: commonActions{ctrl}})
		} else {
			select {
			case retryCh <- ctx:
			default:
				{
				}
			}
		}
	})
}

func processConfirmation(ctx context.Context, sc session.MutableContext, ver, code string, cb LoginQRCallback, actions commonActions) error {
	confirmResult, err := waitingConfirm(ctx, sc, ver, code)
	if err != nil {
		return errs.NewZCA("Cannot get confirm result", "auth.LoginQR")
	}

	if confirmResult.ErrorCode == -13 {
		if cb != nil {
			cb(EventQRCodeDeclined{
				Data:    QRDeclinedData{Code: code},
				Actions: actions,
			})
		} else {
			logger.Log(sc).Error("QRCode login declined")
			return ErrLoginQRDeclined
		}
	} else if confirmResult.ErrorCode != 0 {
		msg := fmt.Sprintf("An error has occurred\nResponse: Code: %d, Message: %s", confirmResult.ErrorCode, confirmResult.ErrorMessage)
		return errs.NewZCA(msg, "auth.LoginQR")
	}

	return nil
}

func finalizeLogin(ctx context.Context, sc session.MutableContext, displayName string) (*httpx.Response[QRUserInfo], error) {
	if err := checkSession(ctx, sc); err != nil {
		return nil, errs.NewZCA("Cannot get session, login failed", "auth.LoginQR")
	}

	logger.Log(sc).Info("Successfully logged into the account ", displayName)

	userInfo, err := getUserInfo(ctx, sc)
	if err != nil {
		return nil, errs.NewZCA("Can't get account info", "auth.LoginQR")
	} else if !userInfo.Data.Logged {
		return nil, errs.NewZCA("Can't login", "auth.LoginQR")
	}

	return userInfo, nil
}

func loadLoginPage(ctx context.Context, sc session.MutableContext) (string, error) {
	h := http.Header{}
	h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Cache-Control", "max-age=0")
	h.Set("Priority", "u=0, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "document")
	h.Set("Sec-Fetch-Mode", "navigate")
	h.Set("Sec-fetch-Site", "same-site")
	h.Set("Sec-fetch-User", "?1")
	h.Set("Upgrade-Insecure-Requests", "1")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	opts := httpx.RequestOptions{
		Headers: h,
		Method:  "GET",
	}

	resp, err := httpx.Request(ctx, sc, "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F", &opts)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := httpx.ReadBytes(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	re := regexp.MustCompile(`https:\/\/stc-zlogin\.zdn\.vn\/main-([\d.]+)\.js`)
	match := re.FindSubmatch(body)
	if len(match) < 2 {
		return "", fmt.Errorf("version not found in HTML")
	}

	return string(match[1]), nil
}

func getLoginInfo(ctx context.Context, sc session.MutableContext, ver string) {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-fetch-Site", "same-origin")
	h.Set("Referer", "https://id.zalo.me/account?continue=https%3A%2F%2Fzalo.me%2Fpc")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	body := httpx.BuildFormBody(map[string]string{
		"v":        ver,
		"continue": "https://zalo.me/pc",
	})

	opts := &httpx.RequestOptions{
		Method:  "POST",
		Body:    body,
		Headers: h,
	}
	resp, err := httpx.Request(ctx, sc, "https://id.zalo.me/account/logininfo", opts)
	if err != nil {
		logger.Log(sc).Error(err)
	}
	defer resp.Body.Close()

	// logger.Log(sc).Info(resp.Body)
}

func verifyClient(ctx context.Context, sc session.MutableContext, ver string) {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-fetch-Site", "same-origin")
	h.Set("Referer", "https://id.zalo.me/account?continue=https%3A%2F%2Fzalo.me%2Fpc")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	body := httpx.BuildFormBody(map[string]string{
		"v":        ver,
		"type":     "device",
		"continue": "https://zalo.me/pc",
	})

	opts := &httpx.RequestOptions{
		Method:  "POST",
		Body:    body,
		Headers: h,
	}
	resp, err := httpx.Request(ctx, sc, "https://id.zalo.me/account/verify-client", opts)
	if err != nil {
		logger.Log(sc).Error(err)
	}
	defer resp.Body.Close()

	// logger.Log(sc).Info(resp.Body)
}

func generateQRCode(ctx context.Context, sc session.MutableContext, ver string) (*httpx.Response[QRGeneratedData], error) {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-fetch-Site", "same-origin")
	h.Set("Referer", "https://id.zalo.me/account?continue=https%3A%2F%2Fzalo.me%2Fpc")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	form := httpx.BuildFormBody(map[string]string{
		"v":        ver,
		"continue": "https://zalo.me/pc",
	})

	opts := &httpx.RequestOptions{
		Method:  "POST",
		Body:    form,
		Headers: h,
	}
	resp, err := httpx.Request(ctx, sc, "https://id.zalo.me/account/authen/qr/generate", opts)
	if err != nil {
		logger.Log(sc).Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	var body httpx.Response[QRGeneratedData]
	if err := httpx.ReadJSON(resp, &body); err != nil {
		return nil, fmt.Errorf("failed to read body %w", err)
	}

	return &body, nil
}

func saveQRCodeToFile(sc session.MutableContext, filepath string, imageData []byte) error {
	if filepath == "" {
		filepath = defaultQRPath
	}
	if err := os.WriteFile(filepath, imageData, 0o600); err != nil {
		return err
	}

	logger.Log(sc).Infof("Scan the QR code at '%s' to proceed with login", filepath)
	return nil
}

func waitingScan(
	ctx context.Context,
	sc session.MutableContext,
	ver string,
	code string,
) (*httpx.Response[QRScannedData], error) {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-fetch-Site", "same-origin")
	h.Set("Referer", "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	form := httpx.BuildFormBody(map[string]string{
		"v":        ver,
		"code":     code,
		"continue": "https://zalo.me/pc",
	})

	opts := &httpx.RequestOptions{
		Method:  "POST",
		Body:    form,
		Headers: h,
	}

	resp, err := httpx.Request(ctx, sc, "https://id.zalo.me/account/authen/qr/waiting-scan", opts)
	if err != nil {
		if ctx.Err() == nil {
			logger.Log(sc).Error(err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	var body httpx.Response[QRScannedData]
	if err := httpx.ReadJSON(resp, &body); err != nil {
		return nil, fmt.Errorf("failed to read body %w", err)
	}

	if body.ErrorCode == 8 {
		return waitingScan(ctx, sc, ver, code)
	}

	return &body, nil
}

func waitingConfirm(
	ctx context.Context,
	sc session.MutableContext,
	ver string,
	code string,
) (*httpx.Response[struct{}], error) {
	logger.Log(sc).Info("Please confirm on your phone")

	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-fetch-Site", "same-origin")
	h.Set("Referer", "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	form := httpx.BuildFormBody(map[string]string{
		"v":        ver,
		"code":     code,
		"gToken":   "",
		"gAction":  "CONFIRM_QR",
		"continue": "https://zalo.me/pc",
	})

	opts := &httpx.RequestOptions{
		Method:  "POST",
		Body:    form,
		Headers: h,
	}

	resp, err := httpx.Request(ctx, sc, "https://id.zalo.me/account/authen/qr/waiting-confirm", opts)
	if err != nil {
		if ctx.Err() == nil {
			logger.Log(sc).Error(err)
		}
		return nil, err
	}
	defer resp.Body.Close()

	var body httpx.Response[struct{}]
	if err := httpx.ReadJSON(resp, &body); err != nil {
		return nil, fmt.Errorf("failed to read body %w", err)
	}

	if body.ErrorCode == 8 {
		return waitingConfirm(ctx, sc, ver, code)
	}

	return &body, nil
}

func checkSession(ctx context.Context, sc session.MutableContext) error {
	h := http.Header{}
	h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Priority", "u=0, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "document")
	h.Set("Sec-Fetch-Mode", "navigate")
	h.Set("Sec-fetch-Site", "same-origin")
	h.Set("Upgrade-Insecure-Requests", "1")
	h.Set("Referer", "https://id.zalo.me/account?continue=https%3A%2F%2Fchat.zalo.me%2F")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	opts := httpx.RequestOptions{
		Headers: h,
		Method:  "GET",
	}

	_, err := httpx.Request(ctx, sc, "https://id.zalo.me/account/checksession?continue=https%3A%2F%2Fchat.zalo.me%2Findex.html", &opts)
	if err != nil {
		logger.Log(sc).Error(err)
		return err
	}

	return nil
}

func getUserInfo(ctx context.Context, sc session.MutableContext) (*httpx.Response[QRUserInfo], error) {
	h := http.Header{}
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", "vi-VN,vi;q=0.9,fr-FR;q=0.8,fr;q=0.7,en-US;q=0.6,en;q=0.5")
	h.Set("Cache-Control", "max-age=0")
	h.Set("Priority", "u=1, i")
	h.Set("Sec-CH-UA", `"Chromium";v="130", "Google Chrome";v="130", "Not?A_Brand";v="99"`)
	h.Set("Sec-CH-UA-Mobile", "?0")
	h.Set("Sec-CH-UA-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-fetch-Site", "same-site")
	h.Set("Referer", "https://chat.zalo.me/")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	opts := httpx.RequestOptions{
		Headers: h,
		Method:  "GET",
	}

	resp, err := httpx.Request(ctx, sc, "https://jr.chat.zalo.me/jr/userinfo", &opts)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var body httpx.Response[QRUserInfo]
	if err := httpx.ReadJSON(resp, &body); err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return &body, nil
}
