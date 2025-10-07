package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type EncryptType string

const (
	EncryptTypeHex    EncryptType = "hex"
	EncryptTypeBase64 EncryptType = "base64"
)

func EncodeAESCBC(key []byte, data string, encType EncryptType) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cryptox: new cipher: %w", err)
	}

	plain, err := Pkcs7Padding([]byte(data), aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("cryptox: pkcs7 padding: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	ct := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ct, plain)

	switch encType {
	case EncryptTypeHex:
		return hex.EncodeToString(ct), nil
	default:
		return base64.StdEncoding.EncodeToString(ct), nil
	}
}

func DecodeAESCBC(key []byte, data string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("cryptox: base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cryptox: new cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)

	plain, err = Pkcs7Unpadding(plain, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("cryptox: pkcs7 unpadding: %w", err)
	}

	return plain, nil
}

func DecodeAESGCM(key, iv, aad, ct []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cryptox: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("cryptox: new gcm: %w", err)
	}

	plain, err := gcm.Open(nil, iv, ct, aad)
	if err != nil {
		return nil, fmt.Errorf("cryptox: gcm open: %w", err)
	}

	return plain, nil
}
