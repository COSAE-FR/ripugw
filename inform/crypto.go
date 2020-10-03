/*
 * Copyright (c) 2017 ZAP Québec
 * Copyright (c) 2020 Gaetan Crahay
 *
 * Use of this source code is governed by an MIT-style
 * license that can be found in the LICENSE file or at
 * https://opensource.org/licenses/MIT.
 */

package inform

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

type IV Key

const (
	CBC = iota
	GCM
)

var (
	blockSize int = len(DefaultKey) // bytes
)

func GenerateIV() (IV, error) {
	iv := make([]byte, blockSize)
	if n, err := rand.Read(iv); err != nil {
		return nil, err
	} else if n != blockSize {
		return nil, errors.New("Could not get enough randomness to generate IV")
	}
	return iv, nil
}

func Encrypt(mode int, iv IV, key Key, data []byte, aad []byte) ([]byte, error) {
	switch mode {
	case GCM:
		return encryptGCM(iv, key, data, aad)
	default:
		return encryptCBC(iv, key, data)
	}
}

func encryptCBC(iv IV, key Key, data []byte) (result []byte, err error) {
	if err = assertCryptoParams(iv, key); err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	srcLen := len(data)
	padLen := blockSize - (srcLen % blockSize)
	srcBuf := make([]byte, len(data)+padLen)
	dstBuf := make([]byte, len(data)+padLen)

	copy(srcBuf, data)
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	copy(srcBuf[srcLen:], padding)

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(dstBuf, srcBuf)

	return dstBuf, nil
}

func Decrypt(mode int, iv IV, key Key, data []byte, aad []byte) (result []byte, err error) {
	switch mode {
	case GCM:
		return decryptGCM(iv, key, data, aad)
	default:
		return decryptCBC(iv, key, data)
	}
}

func decryptCBC(iv IV, key Key, data []byte) (result []byte, err error) {
	if err = assertCryptoParams(iv, key); err != nil {
		return
	}

	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("encrypted data must be a multiple of %d bytes", blockSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	dataLen := len(data)
	result = make([]byte, dataLen)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(result, data)

	padLen := int(result[dataLen-1])
	if padLen > blockSize {
		return nil, fmt.Errorf("Invalid padding: %d > %d (blockSize)", padLen, blockSize)
	}

	return result[:dataLen-padLen], nil
}

func assertCryptoParams(iv IV, key Key) error {
	if !Key(iv).IsValid() {
		return fmt.Errorf("iv length must be %d bytes [len(iv)==%d]", blockSize, len(iv))
	}
	if !key.IsValid() {
		return fmt.Errorf("key length must be %d bytes [len(key)==%d]", blockSize, len(key))
	}
	return nil
}

func encryptGCM(iv IV, key Key, data []byte, aad []byte) (result []byte, err error) {
	if err = assertCryptoParams(iv, key); err != nil {
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return
	}
	result = aesgcm.Seal(nil, iv, data, aad)
	return result, nil
}

func decryptGCM(iv IV, key Key, data []byte, aad []byte) (result []byte, err error) {
	if err = assertCryptoParams(iv, key); err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return
	}

	return aesgcm.Open(nil, iv, data, aad)
}
