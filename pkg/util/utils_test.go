package util

import "testing"

func TestGenSecretWithEmptySecret(t *testing.T) {
	if got := GenSecret("", 0, 0); got != nil {
		t.Fatalf("expected nil secret for empty input, got len=%d", len(got))
	}
}

func TestGenSecretWithSecret(t *testing.T) {
	got := GenSecret("abc", 0, 0)
	if len(got) == 0 {
		t.Fatal("expected derived key to be non-empty")
	}
}
