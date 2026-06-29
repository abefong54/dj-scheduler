package crypto_test

import (
	"strings"
	"testing"

	"eventlineup/internal/infrastructure/crypto"
)

// key32 is a throwaway 32-byte (256-bit) AES key for tests.
var key32 = []byte("0123456789abcdef0123456789abcdef")

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plaintext := "line-notify-personal-access-token-abc123"

	ct, err := crypto.Encrypt(plaintext, key32)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ct == "" {
		t.Fatal("ciphertext is empty")
	}
	if strings.Contains(ct, plaintext) {
		t.Fatalf("ciphertext leaks plaintext: %q", ct)
	}

	got, err := crypto.Decrypt(ct, key32)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plaintext {
		t.Fatalf("roundtrip mismatch: got %q want %q", got, plaintext)
	}
}

// Two encryptions of the same plaintext must differ (random nonce), so the
// ciphertext doesn't reveal that two events share a token.
func TestEncryptUsesRandomNonce(t *testing.T) {
	a, err := crypto.Encrypt("same", key32)
	if err != nil {
		t.Fatalf("Encrypt a: %v", err)
	}
	b, err := crypto.Encrypt("same", key32)
	if err != nil {
		t.Fatalf("Encrypt b: %v", err)
	}
	if a == b {
		t.Fatal("two encryptions of the same plaintext are identical — nonce not random")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	ct, err := crypto.Encrypt("secret", key32)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	wrong := []byte("ffffffffffffffffffffffffffffffff")
	if _, err := crypto.Decrypt(ct, wrong); err == nil {
		t.Fatal("expected Decrypt to fail with the wrong key, got nil error")
	}
}

func TestEncryptRejectsNon32ByteKey(t *testing.T) {
	if _, err := crypto.Encrypt("x", []byte("too-short")); err == nil {
		t.Fatal("expected error for a key that is not 32 bytes")
	}
}

func TestDecryptRejectsGarbage(t *testing.T) {
	if _, err := crypto.Decrypt("not-base64-$$$", key32); err == nil {
		t.Fatal("expected error decrypting non-base64 input")
	}
}
