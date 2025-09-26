package httpx

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/session"
)

type EncryptedPayload struct {
	EncryptedData   string `json:"encrypted_data"`
	EncryptedParams Params `json:"encrypted_params"`
	Enk             string `json:"enk"`
}

type EncryptParamResult struct {
	Params map[string]any
	Enk    *string
}

type ParamsEncryptor interface {
	GetEncryptKey() (string, error)
	GetParams() *Params
}

type Params struct {
	ZCID          string `json:"zcid"`
	EncryptVer    string `json:"enc_ver"`
	ZCIDExtension string `json:"zcid_ext"`
}

func (p *Params) ToMap() map[string]any {
	return map[string]any{
		"zcid":     p.ZCID,
		"enc_ver":  p.EncryptVer,
		"zcid_ext": p.ZCIDExtension,
	}
}

type paramsEncryptor struct {
	Params
	EncryptKey *string `json:"encryptKey"`
}

func NewParamsEncryptor(apiType uint, imei string, firstLaunchTime int64) (ParamsEncryptor, error) {
	p := Params{
		ZCID:          "",
		EncryptVer:    "v2",
		ZCIDExtension: randomString(nil, nil),
	}

	pe := &paramsEncryptor{
		Params:     p,
		EncryptKey: nil,
	}

	if err := pe.createZCID(apiType, imei, firstLaunchTime); err != nil {
		return nil, err
	}
	if err := pe.createEncryptKey(); err != nil {
		return nil, err
	}

	return pe, nil
}

func (pe *paramsEncryptor) GetEncryptKey() (string, error) {
	if pe.EncryptKey == nil {
		return "", errs.NewZCAError("didn't create encryptKey yet", "getEncryptKey", nil)
	}
	return *pe.EncryptKey, nil
}

func (pe *paramsEncryptor) GetParams() *Params {
	if pe.ZCID == "" {
		return nil
	}
	return &pe.Params
}

func (pe *paramsEncryptor) createZCID(apiType uint, imei string, firstLaunchTime int64) error {
	if apiType == 0 || imei == "" || firstLaunchTime == 0 {
		return errs.NewZCAError("invalid params", "createZCID", nil)
	}

	key := "3FC4F0D2AB50057BCE0D90D9187A22B1"
	data := fmt.Sprintf("%d,%s,%d", apiType, imei, firstLaunchTime)
	encType := cryptox.EncryptTypeHex

	s, err := cryptox.EncodeAES([]byte(key), data, encType)
	if err != nil {
		return err
	}

	pe.ZCID = strings.ToUpper(s)
	return nil
}

func (pe *paramsEncryptor) createEncryptKey() error {
	if pe.ZCID == "" || pe.ZCIDExtension == "" {
		return errs.NewZCAError("invalid params", "createEncryptKey", nil)
	}

	sum := md5.Sum([]byte(pe.ZCIDExtension))
	nUpper := strings.ToUpper(hex.EncodeToString(sum[:]))

	if err := pe.deriveKey(nUpper, pe.ZCID); err != nil {
		return err
	}

	return nil
}

func (pe *paramsEncryptor) deriveKey(ext, id string) error {
	evenE, _ := processStr(ext)
	evenI, oddI := processStr(id)
	if len(evenE) == 0 || len(evenI) == 0 || len(oddI) == 0 {
		return errs.NewZCAError("invalid params", "deriveKey", nil)
	}

	var b strings.Builder
	b.WriteString(joinFirst(evenE, 8))
	b.WriteString(joinFirst(evenI, 12))
	reversedOdd := reverseCopy(oddI)
	b.WriteString(joinFirst(reversedOdd, 12))

	key := b.String()
	pe.EncryptKey = &key

	return nil
}

func processStr(s string) (even []string, odd []string) {
	if s == "" {
		return nil, nil
	}

	runes := []rune(s)
	for i, r := range runes {
		if i%2 == 0 {
			even = append(even, string(r))
		} else {
			odd = append(odd, string(r))
		}
	}
	return even, odd
}

func joinFirst(parts []string, n int) string {
	if n > len(parts) {
		n = len(parts)
	}
	return strings.Join(parts[:n], "")
}

func reverseCopy[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func randomString(min, max *int) string {
	minLen := 6
	maxLen := 12

	if min != nil {
		minLen = *min
	}
	if max != nil && min != nil && *max > *min {
		maxLen = *max
	}

	length := minLen
	if maxLen > minLen {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(maxLen-minLen+1)))
		length = minLen + int(n.Int64())
	}

	byteLen := (length + 1) / 2
	buf := make([]byte, byteLen)
	_, _ = rand.Read(buf)

	return hex.EncodeToString(buf)[:length]
}

// GetEncryptParam generates encrypted parameters for Zalo API requests
func GetEncryptParam(sc session.Context, encryptParams bool, typeStr string) (*EncryptParamResult, error) {
	data := map[string]any{
		"computer_name": "Web",
		"imei":          sc.IMEI(),
		"language":      sc.Language(),
		"ts":            time.Now().UnixMilli(),
	}

	var enc *EncryptedPayload
	if encryptParams {
		if e, err := EncryptParam(sc, data); err != nil {
			return nil, errs.NewZCAError("Failed to encrypt params", "GetEncryptParam", &err)
		} else {
			enc = e
		}
	}

	params := make(map[string]any, 8)
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

	params["type"] = sc.APIType()
	params["client_version"] = sc.APIVersion()
	if typeStr == "getserverinfo" {
		params["signkey"] = GenerateZaloSignKey(typeStr, map[string]any{
			"imei":           sc.IMEI(),
			"type":           sc.APIType(),
			"client_version": sc.APIVersion(),
			"computer_name":  "Web",
		})
	} else {
		params["signkey"] = GenerateZaloSignKey(typeStr, params)
	}

	var enkPtr *string
	if enc != nil {
		enkPtr = &enc.Enk
	}
	return &EncryptParamResult{Params: params, Enk: enkPtr}, nil
}

// EncryptParam encrypts the provided data using session context
func EncryptParam(sc session.Context, data map[string]any) (*EncryptedPayload, error) {
	enc, err := NewParamsEncryptor(
		sc.APIType(),
		sc.IMEI(),
		time.Now().UnixMilli(),
	)
	if err != nil {
		return nil, ErrEncryptParams(err)
	}

	blob, err := json.Marshal(data)
	if err != nil {
		return nil, ErrEncryptParams(err)
	}

	key, err := enc.GetEncryptKey()
	if err != nil {
		return nil, ErrEncryptParams(err)
	}

	cipher, err := cryptox.EncodeAES([]byte(key), string(blob), "")
	if err != nil {
		return nil, ErrEncryptParams(err)
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

// ErrEncryptParams wraps encryption parameter errors
func ErrEncryptParams(err error) error {
	return errs.NewZCAError("failed to encrypt params", "EncryptParam", &err)
}
