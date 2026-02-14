package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
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
		prev, err = base64.StdEncoding.DecodeString(previousB64)
		if err != nil {
			return nil, errors.New("invalid previous key")
		}
	}
	return &Service{CurrentKey: cur, PreviousKey: prev, CurrentKeyID: "v2", PreviousKeyID: "v1"}, nil
}

func (s *Service) EncryptString(plaintext, aad string) (EncValue, error) {
	fe := FieldEncryption{CurrentKey: s.CurrentKey, PreviousKey: s.PreviousKey, CurrentID: s.CurrentKeyID, PreviousID: s.PreviousKeyID}
	return fe.EncryptString(plaintext, aad)
}

func (s *Service) DecryptString(enc EncValue, aad string) (string, error) {
	fe := FieldEncryption{CurrentKey: s.CurrentKey, PreviousKey: s.PreviousKey, CurrentID: s.CurrentKeyID, PreviousID: s.PreviousKeyID}
	return fe.DecryptString(enc, aad)
}

func (s *Service) RotateIfNeeded(enc EncValue, aad string) (EncValue, error) {
	fe := FieldEncryption{CurrentKey: s.CurrentKey, PreviousKey: s.PreviousKey, CurrentID: s.CurrentKeyID, PreviousID: s.PreviousKeyID}
	return fe.RotateIfNeeded(enc, aad)
}

type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

func DefaultArgon2Params() Argon2Params {
	return Argon2Params{Time: 2, Memory: 64 * 1024, Threads: 2, KeyLen: 32, SaltLen: 16}
}

func HashPassword(password string, p Argon2Params) (string, error) {
	if p.SaltLen == 0 {
		p.SaltLen = 16
	}
	salt := make([]byte, p.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	h := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return fmt.Sprintf(
		"argon2id$v=19$m=%d$t=%d$p=%d$%s$%s",
		p.Memory,
		p.Time,
		p.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(h),
	), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) == 7 && parts[0] == "argon2id" && parts[1] == "v=19" {
		memory, ok := parseParam(parts[2], "m")
		if !ok {
			return false
		}
		timeCost, ok := parseParam(parts[3], "t")
		if !ok {
			return false
		}
		threads, ok := parseParam(parts[4], "p")
		if !ok {
			return false
		}

		salt, err := base64.RawStdEncoding.DecodeString(parts[5])
		if err != nil {
			return false
		}
		expected, err := base64.RawStdEncoding.DecodeString(parts[6])
		if err != nil {
			return false
		}

		got := argon2.IDKey([]byte(password), salt, uint32(timeCost), uint32(memory), uint8(threads), uint32(len(expected)))
		return subtleCompare(got, expected)
	}

	legacySalt := []byte("static-salt-change")
	legacy := argon2.IDKey([]byte(password), legacySalt, 2, 64*1024, 2, 32)
	return base64.StdEncoding.EncodeToString(legacy) == encoded
}

func parseParam(s, key string) (int, bool) {
	if !strings.HasPrefix(s, key+"=") {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimPrefix(s, key+"="))
	if err != nil {
		return 0, false
	}
	return v, true
}

type FieldEncryption struct {
	CurrentKey  []byte
	PreviousKey []byte
	CurrentID   string
	PreviousID  string
}

func (fe FieldEncryption) EncryptString(plaintext, aad string) (EncValue, error) {
	gcm, err := newGCM(fe.CurrentKey)
	if err != nil {
		return EncValue{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return EncValue{}, err
	}
	ct := gcm.Seal(nil, nonce, []byte(plaintext), []byte(aad))
	return EncValue{
		Ciphertext: base64.StdEncoding.EncodeToString(ct),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		KeyID:      fe.CurrentID,
	}, nil
}

func (fe FieldEncryption) DecryptString(enc EncValue, aad string) (string, error) {
	key := fe.CurrentKey
	if enc.KeyID == fe.PreviousID && len(fe.PreviousKey) == 32 {
		key = fe.PreviousKey
	}
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce, err := base64.StdEncoding.DecodeString(enc.Nonce)
	if err != nil {
		return "", err
	}
	ct, err := base64.StdEncoding.DecodeString(enc.Ciphertext)
	if err != nil {
		return "", err
	}
	pt, err := gcm.Open(nil, nonce, ct, []byte(aad))
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func (fe FieldEncryption) RotateIfNeeded(enc EncValue, aad string) (EncValue, error) {
	if enc.KeyID == fe.CurrentID {
		return enc, nil
	}
	pt, err := fe.DecryptString(enc, aad)
	if err != nil {
		return EncValue{}, err
	}
	return fe.EncryptString(pt, aad)
}

func SearchHash(value, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(strings.ToLower(strings.TrimSpace(value))))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
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

func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	v := byte(0)
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
