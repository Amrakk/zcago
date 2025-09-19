package cryptox

import (
	"bytes"
	"errors"
)

var (
	ErrInvalidBlockSize    = errors.New("invalid block size")
	ErrInvalidPKCS7Data    = errors.New("invalid PKCS#7 data")
	ErrInvalidPKCS7Padding = errors.New("invalid PKCS#7 padding")
)

func Pkcs7Padding(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	padLen := blockSize - (len(data) % blockSize)
	if padLen == 0 {
		padLen = blockSize
	}
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...), nil
}

func Pkcs7Unpadding(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrInvalidPKCS7Data
	}

	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize || padLen > len(data) {
		return nil, ErrInvalidPKCS7Padding
	}

	if !bytes.Equal(bytes.Repeat([]byte{byte(padLen)}, padLen), data[len(data)-padLen:]) {
		return nil, ErrInvalidPKCS7Padding
	}

	return data[:len(data)-padLen], nil
}
