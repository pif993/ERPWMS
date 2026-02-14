package crypto

import "testing"

func TestEncryptDecryptRotate(t *testing.T) {
	oldFE := FieldEncryption{
		CurrentKey: []byte("12345678901234567890123456789012"),
		CurrentID:  "v1",
	}
	encOld, err := oldFE.EncryptString("secret", "users:1:email")
	if err != nil {
		t.Fatal(err)
	}

	fe := FieldEncryption{
		CurrentKey:  []byte("abcdefghijklmnopqrstuvwxzy123456"),
		PreviousKey: []byte("12345678901234567890123456789012"),
		CurrentID:   "v2",
		PreviousID:  "v1",
	}
	pt, err := fe.DecryptString(encOld, "users:1:email")
	if err != nil || pt != "secret" {
		t.Fatalf("decrypt failed")
	}
	enc2, err := fe.RotateIfNeeded(encOld, "users:1:email")
	if err != nil || enc2.KeyID != fe.CurrentID {
		t.Fatalf("rotate failed")
	}
}
