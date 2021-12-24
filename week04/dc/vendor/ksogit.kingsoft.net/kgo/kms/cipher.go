package kms

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

type SymmetricEncrypt interface {
	Encrypt(plaintext, key []byte) ([]byte, error)
	Decrypt(ciphertext, key []byte) ([]byte, error)
}

type AES256GCM struct {
	RandomNonce bool
}

func (a AES256GCM) DecodeDecrypt(eData, key string) (string, error) {
	hKey, err := hashKey(key)
	if err != nil {
		return "", err
	}
	edata, err := base64.StdEncoding.DecodeString(eData)
	if err != nil {
		return "", err
	}
	data, err := a.Decrypt(edata, hKey)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a AES256GCM) Encrypt(plaintext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(c, 12)
	if err != nil {
		return nil, err
	}
	var nonce []byte
	if a.RandomNonce {
		nonce, err = a.randomNonce()
	} else {
		nonce, err = a.hashNonce(plaintext)
	}
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (AES256GCM) Decrypt(ciphertext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(c, 12)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertextTooShort")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (AES256GCM) randomNonce() ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func (AES256GCM) hashNonce(plaintext []byte) ([]byte, error) {
	input := make([]byte, len(plaintext))
	copy(input, plaintext)
	seed := []byte("kms-svr")
	input = append(input, seed...)
	s := sha1.New()
	_, err := s.Write(plaintext)
	if err != nil {
		return nil, err
	}
	return s.Sum(nil)[:12], nil
}

func hashKey(key string) ([]byte, error) {
	s := sha256.New()
	_, err := s.Write([]byte(key))
	if err != nil {
		return nil, err
	}
	return s.Sum(nil), nil
}
