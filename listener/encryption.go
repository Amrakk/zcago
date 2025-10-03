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

	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/errs"
)

const (
	EncryptionTypeNone      uint = 0
	EncryptionTypeBase64    uint = 1
	EncryptionTypeAESGCM    uint = 2
	EncryptionTypeAESGCMRaw uint = 3
)

const minEncryptedDataLength = 48

func encodeFrame(p WSPayload) ([]byte, error) {
	if p.Data == nil {
		p.Data = map[string]any{}
	}

	body, err := json.Marshal(p.Data)
	if err != nil {
		return nil, errs.NewZaloAPIError("failed to marshal data", nil)
	}

	buf := make([]byte, 4+len(body))
	buf[0] = uint8(p.Version)
	binary.LittleEndian.PutUint16(buf[1:3], uint16(p.CMD))
	buf[3] = uint8(p.SubCMD)
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
		return nil, errs.NewZCAError("payload is not valid UTF-8", "decode_event_data", nil)
	}

	return parseJSON[T](payload)
}

func extractEncryptionType(parsed BaseWSMessage) (uint, error) {
	encType := parsed.Encrypt

	if encType > EncryptionTypeAESGCMRaw {
		errMsg := fmt.Sprintf("Invalid encryption type, expected 0-%d but got %d", EncryptionTypeAESGCMRaw, encType)
		return 0, errs.NewZaloAPIError(errMsg, nil)
	}

	return encType, nil
}

func decodeAndDecrypt(data string, encType uint, cipherKey string) ([]byte, error) {
	b64Data := data
	if encType != EncryptionTypeBase64 {
		unescaped, err := url.PathUnescape(data)
		if err != nil {
			return nil, errs.NewZCAError("failed to URL unescape data", "decode_and_decrypt", &err)
		}
		b64Data = unescaped
	}

	decoded, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, errs.NewZCAError("failed to base64-decode data", "decode_and_decrypt", &err)
	}

	switch encType {
	case EncryptionTypeBase64:
		return decoded, nil
	case EncryptionTypeAESGCM, EncryptionTypeAESGCMRaw:
		return decryptAESGCM(decoded, cipherKey)
	default:
		return nil, errs.NewZCAError("unsupported encryption type", "decode_and_decrypt", nil)
	}
}

func decryptAESGCM(encryptedData []byte, cipherKey string) ([]byte, error) {
	if cipherKey == "" {
		return nil, errs.NewZCAError("cipher key is required for encrypted data", "decrypt_aes_gcm", nil)
	}

	if len(encryptedData) < minEncryptedDataLength {
		return nil, errs.NewZCAError("encrypted data too short", "decrypt_aes_gcm", nil)
	}

	key, err := base64.StdEncoding.DecodeString(cipherKey)
	if err != nil {
		return nil, errs.NewZCAError("failed to decode cipher key", "decrypt_aes_gcm", &err)
	}

	iv := encryptedData[0:16]
	aad := encryptedData[16:32]
	ciphertext := encryptedData[32:]

	plaintext, err := cryptox.DecodeAESGCM(key, iv, aad, ciphertext)
	if err != nil {
		return nil, errs.NewZCAError("AES-GCM decryption failed", "decrypt_aes_gcm", &err)
	}

	return plaintext, nil
}

func decompressGzip(compressed []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, errs.NewZCAError("failed to create gzip reader", "decompress_gzip", &err)
	}
	defer reader.Close()

	var output bytes.Buffer
	if _, err := io.Copy(&output, reader); err != nil {
		return nil, errs.NewZCAError("failed to decompress gzip data", "decompress_gzip", &err)
	}

	return output.Bytes(), nil
}

func parseJSON[T any](data []byte) (*WSMessage[T], error) {
	var result WSMessage[T]
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errs.NewZCAError("failed to parse JSON data", "parse_json", &err)
	}
	return &result, nil
}
