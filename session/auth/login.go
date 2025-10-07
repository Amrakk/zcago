package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/session"
)

// ----------------------------------------
// Login
// ----------------------------------------

func Login(ctx context.Context, sc session.MutableContext, encryptParams bool) (*session.LoginInfo, error) {
	encryptedParams, err := generateLoginParams(sc, encryptParams)
	if err != nil {
		return nil, err
	}

	response, err := makeLoginRequest(ctx, sc, encryptedParams)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	rawData, err := processLoginResponse(sc, response)
	if err != nil {
		return nil, err
	}

	return decryptAndParseLoginData(encryptedParams, rawData)
}

func generateLoginParams(sc session.MutableContext, encryptParams bool) (*httpx.EncryptParamResult, error) {
	encryptedParams, err := httpx.GetEncryptParam(sc, encryptParams, "getlogininfo")
	if err != nil {
		logger.Log(sc).Error("Failed to generate encrypted parameters:", err)
		return nil, errs.NewZCA("failed to generate encrypted parameters", "auth.generateLoginParams")
	}
	return encryptedParams, nil
}

func makeLoginRequest(ctx context.Context, sc session.MutableContext, encryptedParams *httpx.EncryptParamResult) (*http.Response, error) {
	params := map[string]any{"nretry": 0}

	for k, v := range encryptedParams.Params {
		params[k] = v
	}

	u := httpx.MakeURL(sc, "https://wpa.chat.zalo.me/api/login/getLoginInfo", params, true)
	response, err := httpx.Request(ctx, sc, u, nil)
	if err != nil {
		status := "no response"
		if response != nil {
			status = response.Status
		}
		return nil, errs.WrapZCA("Failed to fetch login info: "+status, "auth.makeLoginRequest", err)
	}

	return response, nil
}

func processLoginResponse(sc session.MutableContext, response *http.Response) (*httpx.BaseResponse, error) {
	data, err := httpx.ParseBaseResponse(response)
	if err != nil {
		logger.Log(sc).Error("Failed to parse login response JSON:", err)
		return nil, errs.WrapZCA("Failed to decode getLoginInfo response: "+response.Status, "auth.processLoginResponse", err)
	}

	return data, nil
}

func decryptAndParseLoginData(encryptedParams *httpx.EncryptParamResult, rawData *httpx.BaseResponse) (*session.LoginInfo, error) {
	if encryptedParams == nil || encryptedParams.Enk == nil {
		return nil, nil
	}

	dataStr := rawData.Data
	if dataStr == nil {
		return nil, errs.NewZCA("Invalid data format in response", "auth.decryptAndParseLoginData")
	}

	decryptedData, err := decryptLoginResponse(encryptedParams.Enk, dataStr)
	if err != nil {
		return nil, errs.WrapZCA("Failed to decrypt response data", "auth.decryptAndParseLoginData", err)
	}

	return decryptedData.Data, nil
}

func decryptLoginResponse(key, data *string) (*httpx.Response[*session.LoginInfo], error) {
	if key == nil || data == nil {
		return nil, errs.NewZCA("key or data is nil", "auth.decryptLoginResponse")
	}

	u, err := url.PathUnescape(*data)
	if err != nil {
		return nil, err
	}

	plain, err := cryptox.DecodeAESCBC([]byte(*key), u)
	if err != nil {
		return nil, err
	}

	var obj httpx.Response[*session.LoginInfo]
	if err := json.Unmarshal(plain, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

// ----------------------------------------
// Get Server Info
// ----------------------------------------

func GetServerInfo(ctx context.Context, sc session.MutableContext, encryptParams bool) (*session.ServerInfo, error) {
	params, err := generateServerInfoParams(sc, encryptParams)
	if err != nil {
		return nil, err
	}

	response, err := makeServerInfoRequest(ctx, sc, params)
	if err != nil {
		return nil, err
	}

	return parseServerInfoResponse(response)
}

func generateServerInfoParams(sc session.MutableContext, encryptParams bool) (map[string]any, error) {
	encryptedParams, err := httpx.GetEncryptParam(sc, encryptParams, "getserverinfo")
	if err != nil {
		logger.Log(sc).Error("Failed to generate encrypted parameters for server info:", err)
		return nil, errs.WrapZCA("failed to generate encrypted parameters", "auth.generateServerInfoParams", err)
	}

	signkey, ok := encryptedParams.Params["signkey"].(string)
	if !ok || signkey == "" {
		logger.Log(sc).Error("Missing signkey in encrypted parameters")
		return nil, errs.NewZCA("missing signkey", "auth.generateServerInfoParams")
	}

	params := map[string]any{
		"signkey":        signkey,
		"imei":           sc.IMEI(),
		"type":           sc.APIType(),
		"client_version": sc.APIVersion(),
		"computer_name":  "Web",
	}

	return params, nil
}

func makeServerInfoRequest(ctx context.Context, sc session.MutableContext, params map[string]any) (*http.Response, error) {
	u := httpx.MakeURL(sc, "https://wpa.chat.zalo.me/api/login/getServerInfo", params, false)
	response, err := httpx.Request(ctx, sc, u, nil)
	if err != nil {
		return nil, errs.WrapZCA("Failed to fetch server info: "+response.Status, "auth.makeServerInfoRequest", err)
	}

	return response, nil
}

func parseServerInfoResponse(response *http.Response) (*session.ServerInfo, error) {
	data, err := httpx.ParseZaloResponse[*session.ServerInfo](response)
	if err != nil {
		return nil, errs.WrapZCA("Failed to decode server info response: "+response.Status, "auth.parseServerInfoResponse", err)
	}

	return data.Data, nil
}
