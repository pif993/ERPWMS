package middleware

import "testing"

func TestHashPayloadDeterministic(t *testing.T) {
	if HashPayload("abc") != HashPayload("abc") {
		t.Fatal("not deterministic")
	}
}
