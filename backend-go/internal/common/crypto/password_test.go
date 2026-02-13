package crypto

import "testing"

func TestPasswordHashVerify(t *testing.T) {
	h, err := HashPassword("Strong!Pass123", DefaultArgon2Params())
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("Strong!Pass123", h) {
		t.Fatal("verify failed")
	}
}
