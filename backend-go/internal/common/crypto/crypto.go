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

// Argon2Params are configurable parameters for Argon2id.
type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Time:    2,
		Memory:  64 * 1024,
		Threads: 2,
		KeyLen:  32,
		SaltLen: 16,
	}
}

// HashPassword returns an encoded Argon2id hash string:
// "argon2id$v=19$m=<mem>$t=<time>$p=<threads>$<b64salt>$<b64hash>"
func HashPassword(password string, p Argon2Params) (string, error) {
	if p.SaltLen < 8 {
		p.SaltLen = 16
	}
	salt := make([]byte, p.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("argon2id$v=19$m=%d$t=%d$p=%d$%s$%s", p.Memory, p.Time, p.Threads, b64Salt, b64Hash), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	// expected: argon2id, v=19, m=..., t=..., p=..., salt, hash
	if len(parts) == 7 && parts[0] == "argon2id" && parts[1] == "v=19" {
		mem, ok := parseParam(parts[2], "m")
		if !ok {
			return false
		}
		tm, ok := parseParam(parts[3], "t")
		if !ok {
			return false
		}
		th, ok := parseParam(parts[4], "p")
		if !ok {
			return false
		}
		salt, err := base64.RawStdEncoding.DecodeString(parts[5])
		if err != nil {
			return false
		}
		want, err := base64.RawStdEncoding.DecodeString(parts[6])
		if err != nil {
			return false
		}
		got := argon2.IDKey([]byte(password), salt, uint32(tm), uint32(mem), uint8(th), uint32(len(want)))
		return subtleEqual(got, want)
	}

	// Legacy fallback (dev only): old repo used base64(hash) with static salt.
	legacySalt := []byte("static-salt-change")
	got := argon2.IDKey([]byte(password), legacySalt, 2, 64*1024, 2, 32)
	return base64.StdEncoding.EncodeToString(got) == encoded
}

func parseParam(s, key string) (int, bool) {
	if !strings.HasPrefix(s, key+"=") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(s, key+"="))
	if err != nil {
		return 0, false
	}
	return n, true
}

func subtleEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// FieldEncryption provides AES-256-GCM field-level encryption with key rotation.
type FieldEncryption struct {
	CurrentKey  []byte
	PreviousKey []byte
	CurrentID   string
	PreviousID  string
}

func (fe FieldEncryption) EncryptString(plaintext, aad string) (EncValue, error) {
	if len(fe.CurrentKey) != 32 {
		return EncValue{}, errors.New("invalid current field key")
	}
	block, err := aes.NewCipher(fe.CurrentKey)
	if err != nil {
		return EncValue{}, err
	}
	gcm, err := cipher.NewGCM(block)
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
	var key []byte
	switch enc.KeyID {
	case fe.CurrentID:
		key = fe.CurrentKey
	case fe.PreviousID:
		key = fe.PreviousKey
	default:
		key = fe.CurrentKey
	}
	if len(key) != 32 {
		return "", errors.New("invalid field key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
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

// SearchHash is used for deterministic lookup (e.g., email_hash-like).
// NOTE: do not use for passwords; it's for lookups only.
func SearchHash(value, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(strings.ToLower(strings.TrimSpace(value))))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
