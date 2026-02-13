package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

type EncValue struct {
	Ciphertext string
	Nonce      string
	KeyID      string
}

type FieldEncryption struct {
	CurrentKey []byte
	CurrentID  string
	PrevKey    []byte
	PrevID     string
}

func (f FieldEncryption) EncryptString(plaintext, aad string) (EncValue, error) {
	gcm, err := newGCM(f.CurrentKey)
	if err != nil {
		return EncValue{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return EncValue{}, err
	}
	ct := gcm.Seal(nil, nonce, []byte(plaintext), []byte(aad))
	return EncValue{Ciphertext: base64.StdEncoding.EncodeToString(ct), Nonce: base64.StdEncoding.EncodeToString(nonce), KeyID: f.CurrentID}, nil
}

func (f FieldEncryption) DecryptString(v EncValue, aad string) (string, error) {
	key := f.CurrentKey
	if v.KeyID == f.PrevID && len(f.PrevKey) == 32 {
		key = f.PrevKey
	}
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(v.Nonce)
	if err != nil {
		return "", err
	}
	ct, err := base64.StdEncoding.DecodeString(v.Ciphertext)
	if err != nil {
		return "", err
	}
	pt, err := gcm.Open(nil, nonce, ct, []byte(aad))
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func (f FieldEncryption) RotateIfNeeded(v EncValue, aad string) (EncValue, error) {
	if v.KeyID == f.CurrentID {
		return v, nil
	}
	pt, err := f.DecryptString(v, aad)
	if err != nil {
		return EncValue{}, err
	}
	return f.EncryptString(pt, aad)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != 32 {
		return nil, errors.New("field key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
