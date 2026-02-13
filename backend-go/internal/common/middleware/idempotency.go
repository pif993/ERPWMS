package middleware

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPayload(v string) string {
	s := sha256.Sum256([]byte(v))
	return hex.EncodeToString(s[:])
}
