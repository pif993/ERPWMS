package crypto

import (
	"encoding/base64"
	"testing"
)

func TestEncryptDecryptRotate(t *testing.T) {
	k1 := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	k2 := base64.StdEncoding.EncodeToString([]byte("abcdefghijklmnopqrstuvwxzy123456"))
	oldSvc, _ := NewService(k1, "")
	oldSvc.CurrentKeyID = "v1"
	encOld, err := oldSvc.EncryptString("secret", "users:1:email")
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewService(k2, k1)
	if err != nil {
		t.Fatal(err)
	}
	pt, err := s.DecryptString(encOld, "users:1:email")
	if err != nil || pt != "secret" {
		t.Fatalf("decrypt failed")
	}
	enc2, err := s.RotateIfNeeded(encOld, "users:1:email")
	if err != nil || enc2.KeyID != s.CurrentKeyID {
		t.Fatalf("rotate failed")
	}
}
