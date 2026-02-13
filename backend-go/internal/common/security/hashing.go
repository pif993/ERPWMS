package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func EmailHash(email, pepper string) string {
	return hmacHash(strings.ToLower(strings.TrimSpace(email)), pepper)
}

func IPHash(ip, pepper string) string   { return hmacHash(ip, pepper) }
func UAHash(ua, pepper string) string   { return hmacHash(ua, pepper) }
func TokenHash(v, pepper string) string { return hmacHash(v, pepper) }

func hmacHash(v, pepper string) string {
	h := hmac.New(sha256.New, []byte(pepper))
	h.Write([]byte(v))
	return hex.EncodeToString(h.Sum(nil))
}
