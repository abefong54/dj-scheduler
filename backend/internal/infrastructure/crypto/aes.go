// Package crypto provides symmetric encryption for secrets stored at rest,
// such as the per-event LINE Notify token (US-006).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// KeySize is the required AES key length in bytes (256-bit AES).
const KeySize = 32

// ErrKeySize is returned when a key is not exactly KeySize bytes.
var ErrKeySize = errors.New("crypto: key must be 32 bytes (AES-256)")

// Encrypt seals plaintext with AES-256-GCM under key and returns a base64
// (std) string. A fresh random nonce is generated per call and prepended to
// the ciphertext, so encrypting the same plaintext twice yields different
// outputs. key must be exactly KeySize bytes.
func Encrypt(plaintext string, key []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	// Seal appends the ciphertext+tag to nonce, giving nonce||ciphertext.
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. It returns an error if the input is not valid
// base64, is too short to contain a nonce, or fails GCM authentication (wrong
// key or tampered ciphertext).
func Decrypt(ciphertext string, key []byte) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, sealed := raw[:ns], raw[ns:]
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, ErrKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
