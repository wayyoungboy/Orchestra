package security

import (
	"testing"
)

func TestKeyEncryptor_EncryptDecrypt(t *testing.T) {
	key := "test-key-32-bytes-long-1234567890"
	encryptor, err := NewKeyEncryptor(key)
	if err != nil {
		t.Fatalf("NewKeyEncryptor() error = %v", err)
	}

	plaintext := "my-secret-api-key"
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if ciphertext == plaintext {
		t.Error("ciphertext should not equal plaintext")
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestKeyEncryptor_InvalidKey(t *testing.T) {
	_, err := NewKeyEncryptor("short")
	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestKeyEncryptor_InvalidCiphertext(t *testing.T) {
	encryptor, _ := NewKeyEncryptor("test-key-32-bytes-long-1234567890")

	_, err := encryptor.Decrypt("invalid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}