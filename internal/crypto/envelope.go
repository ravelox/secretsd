package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type KEK interface {
	Wrap(DEK []byte) (wrapped []byte, kid string, err error)
	Unwrap(wrapped []byte, kid string) (DEK []byte, err error)
}

type Envelope struct{ kek KEK }

func NewEnvelope(k KEK) *Envelope { return &Envelope{kek: k} }

// Encrypt returns ciphertext with the nonce prefixed: [nonce||ciphertext]
func (e *Envelope) Encrypt(plaintext []byte) (ciphertext, wrappedDEK []byte, kekID string, err error) {
	dek := make([]byte, 32) // AES-256
	if _, err = io.ReadFull(rand.Reader, dek); err != nil { return nil, nil, "", fmt.Errorf("rand dek: %w", err) }

	block, err := aes.NewCipher(dek)
	if err != nil { return nil, nil, "", fmt.Errorf("cipher: %w", err) }
	aead, err := cipher.NewGCM(block)
	if err != nil { return nil, nil, "", fmt.Errorf("gcm: %w", err) }
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil { return nil, nil, "", fmt.Errorf("nonce: %w", err) }
	ct := aead.Seal(nil, nonce, plaintext, nil)
	ciphertext = append(nonce, ct...)

	wrapped, kid, err := e.kek.Wrap(dek)
	if err != nil { return nil, nil, "", fmt.Errorf("wrap: %w", err) }
	return ciphertext, wrapped, kid, nil
}

// Decrypt expects ciphertext as [nonce||ciphertext]
func (e *Envelope) Decrypt(ciphertext, wrappedDEK []byte, kekID string) ([]byte, error) {
	dek, err := e.kek.Unwrap(wrappedDEK, kekID)
	if err != nil { return nil, fmt.Errorf("unwrap: %w", err) }
	block, err := aes.NewCipher(dek)
	if err != nil { return nil, fmt.Errorf("cipher: %w", err) }
	aead, err := cipher.NewGCM(block)
	if err != nil { return nil, fmt.Errorf("gcm: %w", err) }
	ns := aead.NonceSize()
	if len(ciphertext) < ns { return nil, fmt.Errorf("ciphertext too short") }
	nonce, ct := ciphertext[:ns], ciphertext[ns:]
	pt, err := aead.Open(nil, nonce, ct, nil)
	if err != nil { return nil, fmt.Errorf("open: %w", err) }
	return pt, nil
}
