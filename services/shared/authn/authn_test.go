package authn

import (
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"testing"
	"time"
)

func TestVerifyRequiresHS256ExpiryAndPrincipal(test *testing.T) {
	secret := strings.Repeat("s", 32)
	for _, name := range []string{"valid", "wrong algorithm", "expired", "no expiry", "no subject", "no organization"} {
		test.Run(name, func(test *testing.T) {
			claims := Claims{OrgID: "org", RegisteredClaims: jwt.RegisteredClaims{Subject: "owner", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
			method := jwt.SigningMethodHS256
			switch name {
			case "wrong algorithm":
				method = jwt.SigningMethodHS512
			case "expired":
				claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
			case "no expiry":
				claims.ExpiresAt = nil
			case "no subject":
				claims.Subject = ""
			case "no organization":
				claims.OrgID = ""
			}
			token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
			if err != nil {
				test.Fatal(err)
			}
			owner, _, err := NewVerifier(secret).Verify(token)
			if name == "valid" {
				if err != nil || owner != "owner" {
					test.Fatalf("valid token rejected: %v", err)
				}
			} else if err == nil {
				test.Fatal("invalid token accepted")
			}
		})
	}
}
