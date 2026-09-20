package crypto

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenIssuer signs and verifies short-lived access tokens (HS256 JWT).
type TokenIssuer struct {
	secret []byte
}

func NewTokenIssuer(secret string) TokenIssuer {
	return TokenIssuer{secret: []byte(secret)}
}

// Claims is the access-token payload: Subject is the user id.
type Claims struct {
	OrgID string `json:"org_id"`
	jwt.RegisteredClaims
}

// Issue mints an access token for the user valid for ttl. It returns the token
// and its expiry.
func (t TokenIssuer) Issue(userID, orgID string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)
	claims := Claims{
		OrgID: orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// Parse validates the token signature and expiry and returns its claims.
func (t TokenIssuer) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(tok *jwt.Token) (any, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return t.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if claims.Subject == "" || claims.OrgID == "" {
		return nil, errors.New("missing principal")
	}
	return claims, nil
}
