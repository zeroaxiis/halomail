// Command backup creates and restores complete PostgreSQL backups in R2.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aashishrajdev/halomail/services/shared/backup"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()
	action := flag.String("action", "snapshot", "snapshot, run, list, verify or restore")
	manifest := flag.String("manifest", "latest", "latest (default), or an exact manifest key from list, for verify/restore")
	flag.Parse()
	c, err := backup.Load(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	if !c.Active() {
		return fmt.Errorf("R2 backup is disabled or credentials are missing")
	}
	engine := backup.Postgres{Source: c.DatabaseURL, DumpBinary: c.DumpBinary, RestoreBinary: c.RestoreBinary}
	service := backup.New(c, backup.NewR2(c), engine)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *action == "run" {
		if err := engine.CheckTools(); err != nil {
			return err
		}
		service.Run(ctx, slog.Default())
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	if (*action == "verify" || *action == "restore") && *manifest == "latest" {
		key, err := service.Latest(ctx)
		if err != nil {
			return err
		}
		*manifest = key
		fmt.Println("Selected backup:", key)
	}
	switch *action {
	case "snapshot":
		key, err := service.Snapshot(ctx)
		if err != nil {
			return err
		}
		fmt.Println(key)
	case "list":
		keys, err := service.List(ctx)
		if err != nil {
			return err
		}
		for _, key := range keys {
			fmt.Println(key)
		}
	case "verify":
		if err := service.Verify(ctx, *manifest); err != nil {
			return err
		}
		fmt.Println("Archive checksum verified.")
	case "restore":
		if os.Getenv("RESTORE_DATABASE_URL") == "" {
			return fmt.Errorf("RESTORE_DATABASE_URL must point to a new, empty PostgreSQL database")
		}
		if err := service.Restore(ctx, *manifest, os.Getenv("RESTORE_DATABASE_URL")); err != nil {
			return err
		}
		fmt.Println("Backup restored. Validate application data before routing traffic to this database.")
	default:
		return fmt.Errorf("unknown backup action")
	}
	return nil
}
