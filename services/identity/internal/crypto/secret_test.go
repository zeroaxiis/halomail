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

func TestSHA256Hex(t *testing.T) {
	cases := map[string]string{
		"":    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"abc": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
	}
	for in, want := range cases {
		if got := SHA256Hex(in); got != want {
			t.Errorf("SHA256Hex(%q) = %s, want %s", in, got, want)
		}
	}
}
