package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

type EncValue struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	KeyID      string `json:"key_id"`
}

type Service struct {
	CurrentKey    []byte
	PreviousKey   []byte
	CurrentKeyID  string
	PreviousKeyID string
}

func NewService(currentB64, previousB64 string) (*Service, error) {
	cur, err := base64.StdEncoding.DecodeString(currentB64)
	if err != nil || len(cur) != 32 {
		return nil, errors.New("invalid current key")
	}
	var prev []byte
	if strings.TrimSpace(previousB64) != "" {
		prev, _ = base64.StdEncoding.DecodeString(previousB64)
	}
	return &Service{CurrentKey: cur, PreviousKey: prev, CurrentKeyID: "v2", PreviousKeyID: "v1"}, nil
}

func (s *Service) EncryptString(plaintext, aad string) (EncValue, error) {
	block, _ := aes.NewCipher(s.CurrentKey)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = rand.Read(nonce)
	ct := gcm.Seal(nil, nonce, []byte(plaintext), []byte(aad))
	return EncValue{Ciphertext: base64.StdEncoding.EncodeToString(ct), Nonce: base64.StdEncoding.EncodeToString(nonce), KeyID: s.CurrentKeyID}, nil
}

func (s *Service) DecryptString(enc EncValue, aad string) (string, error) {
	key := s.CurrentKey
	if enc.KeyID == s.PreviousKeyID && len(s.PreviousKey) == 32 {
		key = s.PreviousKey
	}
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce, _ := base64.StdEncoding.DecodeString(enc.Nonce)
	ct, _ := base64.StdEncoding.DecodeString(enc.Ciphertext)
	pt, err := gcm.Open(nil, nonce, ct, []byte(aad))
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func (s *Service) RotateIfNeeded(enc EncValue, aad string) (EncValue, error) {
	if enc.KeyID == s.CurrentKeyID {
		return enc, nil
	}
	pt, err := s.DecryptString(enc, aad)
	if err != nil {
		return EncValue{}, err
	}
	return s.EncryptString(pt, aad)
}

func HashPassword(password string) string {
	salt := []byte("static-salt-change")
	h := argon2.IDKey([]byte(password), salt, 2, 64*1024, 2, 32)
	return base64.StdEncoding.EncodeToString(h)
}

func CheckPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

func SearchHash(value, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(strings.ToLower(strings.TrimSpace(value))))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
