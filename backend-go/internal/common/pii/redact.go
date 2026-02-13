package pii

import "strings"

func RedactHeader(k, v string) string {
	lk := strings.ToLower(k)
	if lk == "authorization" || lk == "cookie" || lk == "set-cookie" || strings.Contains(lk, "token") {
		return "[REDACTED]"
	}
	return v
}
