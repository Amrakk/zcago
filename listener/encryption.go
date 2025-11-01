package listener

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"unicode/utf8"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/cryptox"
)

const (
	EncryptionTypeNone      uint = 0
	EncryptionTypeBase64    uint = 1
	EncryptionTypeAESGCM    uint = 2
	EncryptionTypeAESGCMRaw uint = 3

	minEncryptedDataLength = 48
)

func encodeFrame(p WSPayload) ([]byte, error) {
	if p.Data == nil {
		p.Data = map[string]any{}
	}

	body, err := json.Marshal(p.Data)
	if err != nil {
		return nil, errs.WrapZCA("failed to marshal data", "listener.encodeFrame", err)
	}

	buf := make([]byte, 4+len(body))
	buf[0] = p.Version
	binary.LittleEndian.PutUint16(buf[1:3], p.CMD)
	buf[3] = p.SubCMD
	copy(buf[4:], body)

	return buf, nil
}

func decodeEventData[T any](parsed BaseWSMessage, cipherKey string) (*WSMessage[T], error) {
	data := parsed.Data

	encType, err := extractEncryptionType(parsed)
	if err != nil {
		return nil, err
	}

	if encType == EncryptionTypeNone {
		return parseJSON[T]([]byte(data))
	}

	payload, err := decodeAndDecrypt(data, encType, cipherKey)
	if err != nil {
		return nil, err
	}

	if encType != EncryptionTypeAESGCMRaw {
		payload, err = decompressGzip(payload)
		if err != nil {
			return nil, err
		}
	}

	if !utf8.Valid(payload) {
		return nil, errs.NewZCA("payload is not valid UTF-8", "listener.decodeEventData")
	}

	return parseJSON[T](payload)
}

func extractEncryptionType(parsed BaseWSMessage) (uint, error) {
	encType := parsed.Encrypt

	if encType > EncryptionTypeAESGCMRaw {
		errMsg := fmt.Sprintf("invalid encryption type, expected 0-%d but got %d", EncryptionTypeAESGCMRaw, encType)
		return 0, errs.NewZCA(errMsg, "listener.extractEncryptionType")
	}

	return encType, nil
}

func decodeAndDecrypt(data string, encType uint, cipherKey string) ([]byte, error) {
	b64Data := data
	if encType != EncryptionTypeBase64 {
		unescaped, err := url.PathUnescape(data)
		if err != nil {
			return nil, errs.WrapZCA("failed to URL unescape data", "listener.decodeAndDecrypt", err)
		}
		b64Data = unescaped
	}

	decoded, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, errs.WrapZCA("failed to base64-decode data", "listener.decodeAndDecrypt", err)
	}

	switch encType {
	case EncryptionTypeBase64:
		return decoded, nil
	case EncryptionTypeAESGCM, EncryptionTypeAESGCMRaw:
		return decryptAESGCM(decoded, cipherKey)
	default:
		return nil, errs.NewZCA("unsupported encryption type", "listener.decodeAndDecrypt")
	}
}

func decryptAESGCM(encryptedData []byte, cipherKey string) ([]byte, error) {
	if cipherKey == "" {
		return nil, errs.NewZCA("cipher key is required for encrypted data", "listener.decryptAESGCM")
	}

	if len(encryptedData) < minEncryptedDataLength {
		return nil, errs.NewZCA("encrypted data too short", "listener.decryptAESGCM")
	}

	key, err := base64.StdEncoding.DecodeString(cipherKey)
	if err != nil {
		return nil, errs.WrapZCA("failed to decode cipher key", "listener.decryptAESGCM", err)
	}

	iv := encryptedData[0:16]
	aad := encryptedData[16:32]
	ciphertext := encryptedData[32:]

	plaintext, err := cryptox.DecodeAESGCM(key, iv, aad, ciphertext)
	if err != nil {
		return nil, errs.WrapZCA("AES-GCM decryption failed", "listener.decryptAESGCM", err)
	}

	return plaintext, nil
}

func decompressGzip(compressed []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, errs.WrapZCA("failed to create gzip reader", "listener.decompressGzip", err)
	}
	defer reader.Close()

	var output bytes.Buffer

	// #nosec G110 — enhance later if needed
	if _, err := io.Copy(&output, reader); err != nil {
		return nil, errs.WrapZCA("failed to decompress gzip data", "listener.decompressGzip", err)
	}

	return output.Bytes(), nil
}

func parseJSON[T any](data []byte) (*WSMessage[T], error) {
	var result WSMessage[T]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errs.WrapZCA("failed to parse JSON data", "listener.parseJSON", err)
	}
	return &result, nil
}
