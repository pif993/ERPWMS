package crypto

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	s := FieldEncryption{CurrentKey: []byte("12345678901234567890123456789012"), CurrentID: "v1"}
	e, err := s.EncryptString("hello", "users:1:email")
	if err != nil {
		t.Fatal(err)
	}
	pt, err := s.DecryptString(e, "users:1:email")
	if err != nil || pt != "hello" {
		t.Fatal("decrypt")
	}
}
