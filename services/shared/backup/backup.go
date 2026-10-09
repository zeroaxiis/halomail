package backup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Manifest struct {
	Version     int       `json:"version"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Archive     string    `json:"archive"`
	SHA256      string    `json:"sha256"`
	Size        int64     `json:"size"`
}

type Service struct {
	config          Config
	store           ObjectStore
	engine          ArchiveEngine
	mu              sync.RWMutex
	lastSnapshot    time.Time
	operationMu     sync.Mutex
	retentionFailed bool
	now             func() time.Time
}

func New(c Config, store ObjectStore, engine ArchiveEngine) *Service {
	return &Service{config: c, store: store, engine: engine, now: time.Now}
}

// Snapshot only publishes a discoverable manifest after downloading and hashing
// the uploaded archive. A failed/cancelled upload cannot replace a good backup.
// Unique keys identify recovery points independently of upload completion order.
func (s *Service) Snapshot(ctx context.Context) (string, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.snapshot(ctx)
}

func (s *Service) snapshot(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	started := s.now().UTC()
	id := started.Format("20060102T150405.000000000Z") + "-" + uuid.NewString()
	file, err := os.CreateTemp("", "halomail-backup-*.dump")
	if err != nil {
		return "", err
	}
	name := file.Name()
	file.Close()
	defer os.Remove(name)
	if err := s.engine.Dump(ctx, name); err != nil {
		return "", err
	}
	file, err = os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", fmt.Errorf("refusing to upload an empty database archive")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	m := Manifest{Version: 1, StartedAt: started, Archive: s.config.Prefix + "/archives/" + id + ".dump", SHA256: hex.EncodeToString(hash.Sum(nil)), Size: size}
	publishing := false
	defer func() {
		// Before publication an unsuccessful upload is not a recovery point.
		// After a manifest PUT starts its outcome may be ambiguous; retain the
		// archive rather than deleting a potentially published backup.
		if !publishing {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			defer cleanupCancel()
			_ = s.store.Delete(cleanupCtx, m.Archive)
		}
	}()
	if err := s.store.Put(ctx, m.Archive, file, "application/octet-stream"); err != nil {
		return "", fmt.Errorf("R2 archive upload failed: %w", err)
	}
	if err := s.download(ctx, m, io.Discard); err != nil {
		return "", fmt.Errorf("R2 archive verification failed: %w", err)
	}
	m.CompletedAt = s.now().UTC()
	body, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	key := s.config.Prefix + "/manifests/" + id + ".json"
	publishing = true
	if err := s.store.Put(ctx, key, bytes.NewReader(body), "application/json"); err != nil {
		return "", fmt.Errorf("R2 manifest upload failed: %w", err)
	}
	s.markVerified(started)
	if err := s.prune(ctx); err != nil {
		return key, fmt.Errorf("backup verified, but removing older batches failed: %w", err)
	}
	return key, nil
}

const retainedBatches = 2

// prune verifies BOTH retained batches before deleting anything. Archive-first
// deletion leaves a manifest to retry if the second deletion fails. A failed
// cleanup may temporarily leave more than two batches; it never deletes a keeper.
func (s *Service) prune(ctx context.Context) (err error) {
	defer func() {
		s.mu.Lock()
		s.retentionFailed = err != nil
		s.mu.Unlock()
	}()
	keys, err := s.List(ctx)
	if err != nil {
		return err
	}
	if len(keys) < retainedBatches {
		return nil
	}
	for _, key := range keys[:retainedBatches] {
		if err := s.Verify(ctx, key); err != nil {
			return fmt.Errorf("retained batch could not be verified: %w", err)
		}
	}
	// Validate every deletion target before performing any deletion.
	victims := make([]Manifest, len(keys)-retainedBatches)
	for i, key := range keys[retainedBatches:] {
		victims[i], err = s.ReadManifest(ctx, key)
		if err != nil {
			return err
		}
	}
	for i, m := range victims {
		if err := s.store.Delete(ctx, m.Archive); err != nil {
			return err
		}
		if err := s.store.Delete(ctx, keys[retainedBatches+i]); err != nil {
			return err
		}
	}
	// Abandoned uploads can be removed once older than BOTH retained batches
	// and twice the operation timeout. Leave unrelated/recent objects alone.
	oldestKept, err := s.ReadManifest(ctx, keys[retainedBatches-1])
	if err != nil {
		return err
	}
	cutoff := oldestKept.StartedAt
	if grace := s.now().Add(-2 * s.config.Timeout); grace.Before(cutoff) {
		cutoff = grace
	}
	archives, err := s.store.List(ctx, s.config.Prefix+"/archives/")
	if err != nil {
		return err
	}
	referenced := make(map[string]bool, len(keys))
	for _, key := range keys {
		referenced[strings.TrimSuffix(path.Base(key), ".json")+".dump"] = true
	}
	for _, archive := range archives {
		base := path.Base(archive)
		if path.Dir(archive) != s.config.Prefix+"/archives" || !strings.HasSuffix(base, ".dump") || referenced[base] {
			continue
		}
		manifestKey := s.config.Prefix + "/manifests/" + strings.TrimSuffix(base, ".dump") + ".json"
		if !s.manifestKey(manifestKey) {
			continue
		}
		started, _ := time.Parse("20060102T150405.000000000Z", base[:len("20060102T150405.000000000Z")])
		if started.Before(cutoff) {
			if err := s.store.Delete(ctx, archive); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) markVerified(started time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if started.After(s.lastSnapshot) {
		s.lastSnapshot = started
	}
}

// Latest selects a published recovery point without needing a live source DB.
func (s *Service) Latest(ctx context.Context) (string, error) {
	keys, err := s.List(ctx)
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "", fmt.Errorf("no completed R2 backups found")
	}
	return keys[0], nil
}

func (s *Service) List(ctx context.Context) ([]string, error) {
	keys, err := s.store.List(ctx, s.config.Prefix+"/manifests/")
	if err != nil {
		return nil, err
	}
	valid := keys[:0]
	for _, key := range keys {
		if s.manifestKey(key) {
			valid = append(valid, key)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(valid)))
	return valid, nil
}

func (s *Service) manifestKey(key string) bool {
	if path.Dir(key) != s.config.Prefix+"/manifests" || !strings.HasSuffix(key, ".json") || path.Clean(key) != key {
		return false
	}
	id := strings.TrimSuffix(path.Base(key), ".json")
	const stampLength = len("20060102T150405.000000000Z")
	if len(id) != stampLength+1+36 || id[stampLength] != '-' {
		return false
	}
	_, timeErr := time.Parse("20060102T150405.000000000Z", id[:stampLength])
	_, uuidErr := uuid.Parse(id[stampLength+1:])
	return timeErr == nil && uuidErr == nil
}

func (s *Service) ReadManifest(ctx context.Context, key string) (Manifest, error) {
	var m Manifest
	if !s.manifestKey(key) {
		return m, fmt.Errorf("manifest key is outside the configured backup prefix")
	}
	body, err := s.store.Get(ctx, key)
	if err != nil {
		return m, err
	}
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, (64<<10)+1))
	if err != nil || len(data) > 64<<10 {
		return m, fmt.Errorf("invalid backup manifest size")
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("invalid backup manifest")
	}
	expectedArchive := s.config.Prefix + "/archives/" + strings.TrimSuffix(path.Base(key), ".json") + ".dump"
	hash, err := hex.DecodeString(m.SHA256)
	if m.Version != 1 || m.Size <= 0 || err != nil || len(hash) != sha256.Size || m.Archive != expectedArchive || m.StartedAt.IsZero() || m.CompletedAt.Before(m.StartedAt) {
		return m, fmt.Errorf("invalid backup manifest contents")
	}
	if !strings.HasPrefix(path.Base(key), m.StartedAt.UTC().Format("20060102T150405.000000000Z")+"-") {
		return m, fmt.Errorf("manifest timestamp does not match its key")
	}
	return m, nil
}

func (s *Service) download(ctx context.Context, m Manifest, destination io.Writer) error {
	body, err := s.store.Get(ctx, m.Archive)
	if err != nil {
		return err
	}
	defer body.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(destination, hash), io.LimitReader(body, m.Size+1))
	if err != nil {
		return err
	}
	if size != m.Size || hex.EncodeToString(hash.Sum(nil)) != m.SHA256 {
		return fmt.Errorf("archive size or SHA-256 checksum does not match")
	}
	return nil
}

func (s *Service) Verify(ctx context.Context, key string) error {
	m, err := s.ReadManifest(ctx, key)
	if err != nil {
		return err
	}
	return s.download(ctx, m, io.Discard)
}

func (s *Service) Restore(ctx context.Context, key, target string) error {
	m, err := s.ReadManifest(ctx, key)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp("", "halomail-restore-*.dump")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = s.download(ctx, m, file)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return s.engine.Restore(ctx, file.Name(), target)
}

// Healthy exposes backup staleness independently of database availability.
func (s *Service) Healthy(context.Context) error {
	s.mu.RLock()
	last := s.lastSnapshot
	retentionFailed := s.retentionFailed
	s.mu.RUnlock()
	if last.IsZero() {
		return fmt.Errorf("no verified R2 backup available yet")
	}
	if s.now().Sub(last) > s.config.MaxAge {
		return fmt.Errorf("verified R2 backup is stale")
	}
	if retentionFailed {
		return fmt.Errorf("R2 backup retention cleanup needs retry")
	}
	return nil
}

func (s *Service) Run(ctx context.Context, logger *slog.Logger) {
	for {
		key, delay, err := s.cycle(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error("R2 backup cycle failed; retrying in five minutes", "error", err.Error())
			delay = 5 * time.Minute
		} else if key != "" {
			logger.Info("R2 database backup verified", "manifest", key)
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

// cycle consults R2 on startup/retry, so restarts reuse the existing hourly
// schedule. Anchor to snapshot START time instead of adding an hour after upload.
// Run only one worker per prefix (gateway OR standalone), including across hosts.
func (s *Service) cycle(ctx context.Context) (string, time.Duration, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	keys, err := s.List(ctx)
	if err != nil {
		return "", 0, err
	}
	if len(keys) > 0 {
		m, err := s.ReadManifest(ctx, keys[0])
		if err != nil {
			return "", 0, err
		}
		next := m.StartedAt.Add(s.config.Interval)
		if next.After(s.now()) {
			if err := s.Verify(ctx, keys[0]); err != nil {
				return "", 0, err
			}
			s.markVerified(m.StartedAt)
			if err := s.prune(ctx); err != nil {
				return "", 0, err
			}
			return "", max(time.Second, next.Sub(s.now())), nil
		}
	}
	key, err := s.snapshot(ctx)
	if err != nil {
		return key, 0, err
	}
	s.mu.RLock()
	next := s.lastSnapshot.Add(s.config.Interval)
	s.mu.RUnlock()
	return key, max(time.Second, next.Sub(s.now())), nil
}

// Open creates a configured backup service. A nil result means disabled or no
// R2 credentials. Invalid partial configuration is never silently ignored.
func Open(databaseURL string) (*Service, error) {
	c, err := Load(databaseURL)
	if err != nil {
		return nil, err
	}
	if !c.Active() {
		return nil, nil
	}
	p := Postgres{Source: c.DatabaseURL, DumpBinary: c.DumpBinary, RestoreBinary: c.RestoreBinary}
	if err := p.CheckTools(); err != nil {
		return nil, err
	}
	if _, _, err := toolConnection(c.DatabaseURL); err != nil {
		return nil, err
	}
	return New(c, NewR2(c), p), nil
}
