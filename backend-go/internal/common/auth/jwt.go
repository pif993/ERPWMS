package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	Issuer   string
	Audience string
	Current  []byte
	Previous []byte
}

func (j JWTManager) Issue(userID string, ttl time.Duration) (string, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iss": j.Issuer,
		"aud": j.Audience,
		"exp": time.Now().Add(ttl).Unix(),
	})
	return tok.SignedString(j.Current)
}

func (j JWTManager) Parse(tokenStr string) (string, error) {
	claims := jwt.MapClaims{}
	keyFn := func(token *jwt.Token) (interface{}, error) { return j.Current, nil }
	t, err := jwt.ParseWithClaims(tokenStr, claims, keyFn, jwt.WithAudience(j.Audience), jwt.WithIssuer(j.Issuer))
	if err != nil && len(j.Previous) > 0 {
		t, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return j.Previous, nil }, jwt.WithAudience(j.Audience), jwt.WithIssuer(j.Issuer))
	}
	if err != nil || !t.Valid {
		return "", err
	}
	sub, _ := claims["sub"].(string)
	return sub, nil
}
