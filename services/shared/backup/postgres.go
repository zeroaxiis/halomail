package backup

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/jackc/pgx/v5"
)

type ArchiveEngine interface {
	Dump(context.Context, string) error
	Restore(context.Context, string, string) error
}

type Postgres struct{ Source, DumpBinary, RestoreBinary string }

func (p Postgres) CheckTools() error {
	for _, binary := range []string{p.DumpBinary, p.RestoreBinary} {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("PostgreSQL backup tools are missing; install pg_dump and pg_restore or use the deployment image")
		}
	}
	return nil
}

// pg_dump captures one consistent snapshot including every schema, table,
// sequence, function, constraint, index and large object in this database.
func (p Postgres) Dump(ctx context.Context, file string) error {
	if err := runPostgresTool(ctx, p.DumpBinary, p.Source, "--format=custom", "--no-owner", "--no-privileges", "--no-publications", "--no-subscriptions", "--file="+file); err != nil {
		return err
	}
	return p.inspect(ctx, file)
}

func (p Postgres) inspect(ctx context.Context, file string) error {
	if err := exec.CommandContext(ctx, p.RestoreBinary, "--list", file).Run(); err != nil {
		return fmt.Errorf("pg_restore could not read the backup archive")
	}
	return nil
}

func (p Postgres) Restore(ctx context.Context, file, target string) error {
	if target == "" {
		return fmt.Errorf("RESTORE_DATABASE_URL is required; use a new, empty PostgreSQL database")
	}
	if err := p.inspect(ctx, file); err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, target)
	if err != nil {
		return fmt.Errorf("cannot connect to restore target")
	}
	defer conn.Close(context.Background())
	var objects int
	err = conn.QueryRow(ctx, `SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname NOT IN ('pg_catalog','information_schema') AND n.nspname NOT LIKE 'pg_toast%' AND n.nspname NOT LIKE 'pg_temp_%' AND c.relkind IN ('r','p','v','m','S','f')`).Scan(&objects)
	if err != nil {
		return fmt.Errorf("cannot check restore target")
	}
	if objects != 0 {
		return fmt.Errorf("restore refused: target database is not empty; nothing was overwritten")
	}
	// No --clean: concurrent object creation causes failure, never destructive replacement.
	// A single transaction rolls back the restore if any statement fails.
	return runPostgresTool(ctx, p.RestoreBinary, target, "--exit-on-error", "--single-transaction", "--no-owner", "--no-privileges", "--no-publications", "--no-subscriptions", file)
}

func toolConnection(connection string) (string, []string, error) {
	u, err := url.Parse(connection)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" || u.User == nil {
		return "", nil, fmt.Errorf("backup database connection must be a postgres:// or postgresql:// URL")
	}
	password, _ := u.User.Password()
	u.User = url.User(u.User.Username())
	q := u.Query()
	if q.Has("password") {
		password = q.Get("password")
		q.Del("password")
	}
	u.RawQuery = q.Encode()
	// Keep credentials out of the process command line and prevent inherited
	// libpq settings from redirecting the connection to a different database.
	var environment []string
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(strings.ToUpper(name), "PG") {
			environment = append(environment, entry)
		}
	}
	environment = append(environment, "PGPASSWORD="+password, "PGCONNECT_TIMEOUT=15")
	return u.String(), environment, nil
}

func runPostgresTool(ctx context.Context, binary, database string, arguments ...string) error {
	connection, environment, err := toolConnection(database)
	if err != nil {
		return err
	}
	arguments = append([]string{"--no-password", "--dbname=" + connection}, arguments...)
	cmd := exec.CommandContext(ctx, binary, arguments...)
	cmd.Env = environment
	// Do not forward stderr: SQL errors can contain private row data or DSNs.
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("PostgreSQL backup/restore timed out or was cancelled")
		}
		return fmt.Errorf("PostgreSQL backup/restore failed; check connectivity, permissions, disk space and client/server version compatibility")
	}
	return nil
}
