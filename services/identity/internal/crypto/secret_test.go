package crypto

import (
	"encoding/hex"
	"testing"
)

func TestRandomToken(t *testing.T) {
	first, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	if raw, err := hex.DecodeString(first); err != nil || len(raw) != 32 {
		t.Fatalf("token is not 32 hex-encoded bytes: %q", first)
	}
	second, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two random tokens are identical")
	}
}
