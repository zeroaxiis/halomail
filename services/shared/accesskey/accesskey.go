package accesskey

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Owner struct {
	ID     string
	Email  string
	Name   string
	Handle string
}

func Verify(ctx context.Context, pool *pgxpool.Pool, secret, feature string) (Owner, error) {
	owner := Owner{}
	if len(secret) < 32 || len(secret) > 128 {
		return owner, errs.Unauthorized("invalid access key")
	}
	scope := "forms:submit"
	if feature == "meetings" {
		scope = "meetings:book"
	}
	hash := sha256.Sum256([]byte(secret))
	err := pool.QueryRow(ctx, `SELECT u.id,u.email,u.name,u.handle FROM api_keys k JOIN users u ON u.id=k.user_id
	WHERE k.secret_hash=$1 AND NOT k.revoked AND k.scopes=ARRAY[$2]::text[]`, hex.EncodeToString(hash[:]), scope).Scan(&owner.ID, &owner.Email, &owner.Name, &owner.Handle)
	if err != nil {
		return Owner{}, errs.Unauthorized("invalid or deleted access key")
	}
	return owner, nil
}
