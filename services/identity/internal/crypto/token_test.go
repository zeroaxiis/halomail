package crypto

import (
	"strings"
	"testing"
	"time"
)

func TestTokenIssueAndParse(t *testing.T) {
	issuer := NewTokenIssuer(strings.Repeat("s", 32))
	token, exp, err := issuer.Issue("usr_1", "org_1", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if left := time.Until(exp); left < 14*time.Minute || left > 15*time.Minute {
		t.Fatalf("expiry %v is not ~15m away", exp)
	}
	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "usr_1" || claims.OrgID != "org_1" {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestTokenParseRejectsOtherSecret(t *testing.T) {
	token, _, err := NewTokenIssuer(strings.Repeat("a", 32)).Issue("usr_1", "org_1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTokenIssuer(strings.Repeat("b", 32)).Parse(token); err == nil {
		t.Fatal("token signed with a different secret accepted")
	}
}

func TestTokenParseRejectsExpiredAndIncomplete(t *testing.T) {
	issuer := NewTokenIssuer(strings.Repeat("s", 32))

	expired, _, err := issuer.Issue("usr_1", "org_1", -time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	noOrg, _, err := issuer.Issue("usr_1", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	noUser, _, err := issuer.Issue("", "org_1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for name, token := range map[string]string{"expired": expired, "no org": noOrg, "no user": noUser, "garbage": "abc"} {
		if _, err := issuer.Parse(token); err == nil {
			t.Errorf("%s token accepted", name)
		}
	}
}
