package crypto

import (
	"testing"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := VerifyPassword(hash, "correct horse battery staple"); err != nil || !ok {
		t.Fatalf("correct password rejected: ok=%v err=%v", ok, err)
	}
	if ok, err := VerifyPassword(hash, "Correct horse battery staple"); err != nil || ok {
		t.Fatalf("wrong password accepted: ok=%v err=%v", ok, err)
	}
}

func TestVerifyPasswordRejectsMalformedHashes(t *testing.T) {
	for _, encoded := range []string{
		"",
		"plaintext",
		"$bcrypt$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA",
		"$argon2id$v=18$m=65536,t=1,p=4$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=lots$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=65536,t=1,p=4$!!!$aGFzaA",
		"$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$!!!",
	} {
		if ok, err := VerifyPassword(encoded, "password"); ok || err == nil {
			t.Errorf("malformed hash %q: ok=%v err=%v", encoded, ok, err)
		}
	}
}
