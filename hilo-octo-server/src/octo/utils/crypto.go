package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func EncryptAes256(data []byte, key [32]byte) ([]byte, error) {
	cipherBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	data = padByPkcs7(data) // add padding
	encrypted := make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(cipherBlock, iv)
	mode.CryptBlocks(encrypted[aes.BlockSize:], data)
	return encrypted, nil
}

func padByPkcs7(data []byte) []byte {
	padSize := aes.BlockSize
	if len(data) % aes.BlockSize != 0 {
		padSize = aes.BlockSize - (len(data)) % aes.BlockSize
	}

	pad := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(data, pad...)
}
