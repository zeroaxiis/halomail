// Package authn verifies access tokens. Every service shares the same JWT
// secret, so any service (or the gateway) can validate a bearer token locally
// without a round-trip to identity.
package authn

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// Verifier validates HS256 access tokens issued by the identity service.
type Verifier struct {
	secret []byte
}

func NewVerifier(secret string) Verifier { return Verifier{secret: []byte(secret)} }

// Claims mirrors the identity access-token payload (Subject = user id).
type Claims struct {
	OrgID string `json:"org_id"`
	jwt.RegisteredClaims
}

// Verify parses and validates token, returning the principal.
func (v Verifier) Verify(token string) (userID, orgID string, err error) {
	c := &Claims{}
	if _, err = jwt.ParseWithClaims(token, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired()); err != nil {
		return "", "", err
	}
	if c.Subject == "" || c.OrgID == "" {
		return "", "", errors.New("missing principal")
	}
	return c.Subject, c.OrgID, nil
}
