package accesskey

import (
	"context"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"strings"
	"testing"
)

// Secrets outside the accepted length are rejected before any query runs, so
// a nil pool is safe here.
func TestVerifyRejectsSecretsOfInvalidLength(t *testing.T) {
	for _, secret := range []string{"", "short", strings.Repeat("k", 31), strings.Repeat("k", 129)} {
		_, err := Verify(context.Background(), nil, secret, "forms")
		if errs.KindOf(err) != errs.KindUnauthorized {
			t.Errorf("secret of length %d: err = %v, want unauthorized", len(secret), err)
		}
	}
}
