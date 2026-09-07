package cryptox

import (
	"errors"
	"math"
)

var (
	ErrInvalidBlockSize    = errors.New("cryptox: invalid block size")
	ErrInvalidPKCS7Data    = errors.New("cryptox: invalid PKCS#7 data")
	ErrInvalidPKCS7Padding = errors.New("cryptox: invalid PKCS#7 padding")
)

const maxPKCS7BlockSize = math.MaxUint8

func Pkcs7Padding(data []byte, blockSize int) ([]byte, error) {
	if err := validatePKCS7BlockSize(blockSize); err != nil {
		return nil, err
	}

	padLen := blockSize - len(data)%blockSize
	if padLen < 1 || padLen > maxPKCS7BlockSize {
		return nil, ErrInvalidPKCS7Padding
	}

	paddingByte := byte(padLen)

	result := make([]byte, len(data)+padLen)
	copy(result, data)

	for i := len(data); i < len(result); i++ {
		result[i] = paddingByte
	}

	return result, nil
}

func Pkcs7Unpadding(data []byte, blockSize int) ([]byte, error) {
	if err := validatePKCS7BlockSize(blockSize); err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrInvalidPKCS7Data
	}

	paddingByte := data[len(data)-1]
	padLen := int(paddingByte)

	if padLen == 0 || padLen > blockSize || padLen > len(data) {
		return nil, ErrInvalidPKCS7Padding
	}

	for _, value := range data[len(data)-padLen:] {
		if value != paddingByte {
			return nil, ErrInvalidPKCS7Padding
		}
	}

	return data[:len(data)-padLen], nil
}

func validatePKCS7BlockSize(blockSize int) error {
	if blockSize <= 0 || blockSize > maxPKCS7BlockSize {
		return ErrInvalidBlockSize
	}

	return nil
}
