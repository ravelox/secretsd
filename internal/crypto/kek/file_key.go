package kek

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileKEK struct {
	key []byte
	kid string // "file://<abs path>"
}

func NewFileKEK(path string) (*FileKEK, error) {
	ap, err := filepath.Abs(path)
	if err != nil { return nil, err }
	key, err := loadOrCreate32(ap)
	if err != nil { return nil, err }
	return &FileKEK{key: key, kid: "file://" + ap}, nil
}

func (f *FileKEK) Wrap(DEK []byte) ([]byte, string, error) {
	block, err := aes.NewCipher(f.key)
	if err != nil { return nil, "", err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return nil, "", err }
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil { return nil, "", err }
	ct := aead.Seal(nil, nonce, DEK, nil)
	wrapped := append(nonce, ct...)
	return wrapped, f.kid, nil
}

func (f *FileKEK) Unwrap(wrapped []byte, kid string) ([]byte, error) {
	if kid != "" && kid != f.kid {
		return nil, fmt.Errorf("kek mismatch: have %s want %s", f.kid, kid)
	}
	block, err := aes.NewCipher(f.key)
	if err != nil { return nil, err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return nil, err }
	ns := aead.NonceSize()
	if len(wrapped) < ns { return nil, fmt.Errorf("wrapped too short") }
	nonce, ct := wrapped[:ns], wrapped[ns:]
	return aead.Open(nil, nonce, ct, nil)
}

func loadOrCreate32(path string) ([]byte, error) {
	if b, err := os.ReadFile(path); err == nil && len(b) == 32 { return b, nil }
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil { return nil, err }
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { return nil, err }
	if err := os.WriteFile(path, key, 0o600); err != nil { return nil, err }
	fmt.Fprintf(os.Stderr, "Generated new KEK at %s (hex %s...)\n", path, hex.EncodeToString(key)[:8])
	return key, nil
}
