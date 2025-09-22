package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"time"

	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/session"
)

type EncryptedPayload struct {
	EncryptedData   string       `json:"encrypted_data"`
	EncryptedParams httpx.Params `json:"encrypted_params"`
	Enk             string       `json:"enk"`
}

type ContextBase struct {
	IMEI          string `json:"imei"`
	Type          uint   `json:"type"`
	ClientVersion uint   `json:"client_version"`
	ComputerName  string `json:"computer_name"`
}

type EncryptParamResult struct {
	Params map[string]any
	Enk    *string
}

type ServerInfo struct {
	Settings *session.Settings `json:"settings"`
	ExtraVer *session.ExtraVer `json:"extra_ver"`
}

func Login(ctx context.Context, sc session.MutableContext, encryptParams bool) (*session.LoginInfo, error) {
	encryptedParams, err := getEncryptParam(sc, encryptParams, "getlogininfo")
	if err != nil {
		logger.Log(sc).Error("Failed to generate encrypted parameters:", err)
		return nil, err
	}

	params := map[string]any{
		"nretry": 0,
	}
	for k, v := range encryptedParams.Params {
		params[k] = v
	}

	u, _ := httpx.MakeURL(sc, "https://wpa.chat.zalo.me/api/login/getLoginInfo", params, true)
	response, err := httpx.Request(ctx, sc, u, nil, false)
	if err != nil {
		status := "no response"
		if response != nil {
			status = response.Status
		}
		return nil, errs.NewZCAError("Failed to fetch server info: "+status, "GetServerInfo", &err)
	}

	defer response.Body.Close()

	body, err := httpx.DecodeBody(response)
	if err != nil {
		logger.Log(sc).Error("Failed to decode login response body:", err)
		return nil, errs.NewZCAError("Failed to decode getLoginInfo response", "Login", &err)
	}
	defer body.Close()

	raw, readErr := io.ReadAll(body)
	if readErr != nil {
		logger.Log(sc).Error("Failed to read login response:", readErr)
		return nil, errs.NewZCAError("Failed to read getLoginInfo response", "Login", &readErr)
	}

	var data httpx.EncryptedResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		logger.Log(sc).Error("Failed to parse login response JSON:", err).
			Debug("Raw response:", string(raw))
		return nil, errs.NewZCAError("Failed to decode getLoginInfo response: "+response.Status, "Login", &err)
	}

	if encryptedParams == nil || encryptedParams.Enk == nil {
		return nil, nil
	}

	dataStr := data.Data
	if dataStr == nil {
		return nil, errs.NewZCAError("Invalid data format in response", "Login", nil)
	}

	decryptedData, err := decryptResp(encryptedParams.Enk, dataStr)
	if err != nil {
		return nil, errs.NewZCAError("Failed to decrypt response data", "Login", &err)
	}

	return decryptedData.Data, nil
}

func GetServerInfo(ctx context.Context, sc session.MutableContext, encryptParams bool) (*ServerInfo, error) {
	encryptedParams, err := getEncryptParam(sc, encryptParams, "getserverinfo")
	if err != nil {
		logger.Log(sc).Error("Failed to generate encrypted parameters for server info:", err)
		return nil, err
	}

	signkey, ok := encryptedParams.Params["signkey"].(string)
	if !ok || signkey == "" {
		logger.Log(sc).Error("Missing signkey in encrypted parameters")
		return nil, errs.NewZCAError("missing signkey", "GetServerInfo", nil)
	}

	params := map[string]any{
		"signkey":        signkey,
		"imei":           sc.IMEI(),
		"type":           sc.APIType(),
		"client_version": sc.APIVersion(),
		"computer_name":  "Web",
	}

	u, _ := httpx.MakeURL(sc, "https://wpa.chat.zalo.me/api/login/getServerInfo", params, false)

	response, err := httpx.Request(ctx, sc, u, nil, false)
	if err != nil {
		return nil, errs.NewZCAError("Failed to fetch server info: "+response.Status, "GetServerInfo", &err)
	}

	body, err := httpx.DecodeBody(response)
	if err != nil {
		logger.Log(sc).Error("Failed to decode server info response body:", err)
		return nil, errs.NewZCAError("Failed to decode getLoginInfo response", "Login", &err)
	}
	defer body.Close()

	defer response.Body.Close()
	raw, readErr := io.ReadAll(body)
	if readErr != nil {
		logger.Log(sc).Error("Failed to read server info response:", readErr)
		return nil, errs.NewZCAError("Failed to read getLoginInfo response", "Login", &readErr)
	}

	var data httpx.BaseResponse[ServerInfo]
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, errs.NewZCAError("Failed to decode server info response: "+response.Status, "GetServerInfo", &err)
	}

	return data.Data, nil
}

func getEncryptParam(sc session.Context, encryptParams bool, typeStr string) (*EncryptParamResult, error) {
	params := make(map[string]any, 8)

	data := map[string]any{
		"computer_name": "Web",
		"imei":          sc.IMEI(),
		"language":      sc.Language(),
		"ts":            time.Now().UnixNano() / int64(time.Millisecond),
	}

	enc, err := encryptParam(sc, data, encryptParams)
	if err != nil {
		return nil, errs.NewZCAError("Failed to encrypt params", "getEncryptParam", &err)
	}

	if enc == nil {
		for k, v := range data {
			params[k] = v
		}
	} else {
		for k, v := range enc.EncryptedParams.ToMap() {
			params[k] = v
		}
		params["params"] = enc.EncryptedData
	}

	params["type"] = sc.APIType
	params["client_version"] = sc.APIVersion

	if typeStr == "getserverinfo" {
		params["signkey"] = httpx.GetSignKey(typeStr, map[string]any{
			"imei":           sc.IMEI(),
			"type":           sc.APIType(),
			"client_version": sc.APIVersion(),
			"computer_name":  "Web",
		})
	} else {
		params["signkey"] = httpx.GetSignKey(typeStr, params)
	}

	var enkPtr *string
	if enc != nil {
		enk := enc.Enk
		enkPtr = &enk
	}

	return &EncryptParamResult{
		Params: params,
		Enk:    enkPtr,
	}, nil
}

func encryptParam(sc session.Context, data map[string]any, encryptParams bool) (*EncryptedPayload, error) {
	if encryptParams {
		enc, err := httpx.NewParamsEncryptor(
			sc.APIType(),
			sc.IMEI(),
			uint(time.Now().UnixNano()/int64(time.Millisecond)),
		)
		if err != nil {
			return nil, errEncryptParams(err)
		}

		blob, err := json.Marshal(data)
		if err != nil {
			return nil, errEncryptParams(err)
		}

		key, err := enc.GetEncryptKey()
		if err != nil {
			return nil, errEncryptParams(err)
		}

		cipher, err := cryptox.EncodeAES([]byte(key), string(blob), "")
		if err != nil {
			return nil, errEncryptParams(err)
		}

		params := enc.GetParams()
		if params == nil {
			return nil, nil
		}

		return &EncryptedPayload{
			EncryptedData:   cipher,
			EncryptedParams: *params,
			Enk:             key,
		}, nil
	}

	return nil, nil
}

func decryptResp(key, data *string) (*httpx.BaseResponse[session.LoginInfo], error) {
	if key == nil || data == nil {
		return nil, errs.NewZCAError("key or data is nil", "decryptResp", nil)
	}

	u, err := url.PathUnescape(*data)
	if err != nil {
		return nil, err
	}

	plain, err := cryptox.DecodeAES([]byte(*key), u)
	if err != nil {
		return nil, err
	}

	var obj httpx.BaseResponse[session.LoginInfo]
	if err := json.Unmarshal([]byte(plain), &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func errEncryptParams(err error) error {
	return errs.NewZCAError("failed to encrypt params", "encryptParam", &err)
}
