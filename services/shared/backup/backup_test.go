package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

type memoryStore struct {
	objects                     map[string][]byte
	failManifest, corruptUpload bool
	failDelete                  string
	deleted                     []string
}

func (s *memoryStore) Delete(_ context.Context, key string) error {
	if key == s.failDelete {
		return errors.New("delete denied")
	}
	s.deleted = append(s.deleted, key)
	delete(s.objects, key)
	return nil
}

func (s *memoryStore) Put(_ context.Context, key string, r io.Reader, _ string) error {
	if s.failManifest && strings.HasSuffix(key, ".json") {
		return errors.New("unavailable")
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if s.corruptUpload && strings.HasSuffix(key, ".dump") {
		body = []byte("corrupted upload")
	}
	s.objects[key] = body
	return nil
}
func (s *memoryStore) Get(_ context.Context, key string) (io.ReadCloser, error) {
	b, ok := s.objects[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (s *memoryStore) List(_ context.Context, prefix string) ([]string, error) {
	var keys []string
	for key := range s.objects {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

type fakeEngine struct {
	data, restored []byte
	err            error
	restoreCalls   int
	afterDump      func()
}

func (e *fakeEngine) Dump(_ context.Context, file string) error {
	if e.err != nil {
		return e.err
	}
	if e.afterDump != nil {
		e.afterDump()
	}
	return os.WriteFile(file, e.data, 0600)
}
func (e *fakeEngine) Restore(_ context.Context, file, _ string) error {
	e.restoreCalls++
	var err error
	e.restored, err = os.ReadFile(file)
	return err
}

func fixture() (*Service, *memoryStore, *fakeEngine) {
	store := &memoryStore{objects: make(map[string][]byte)}
	engine := &fakeEngine{data: []byte("whole database, including schemas and private rows")}
	return New(Config{Prefix: "test/site", Interval: time.Hour, Timeout: time.Minute, MaxAge: 2 * time.Hour}, store, engine), store, engine
}

func TestSnapshotRoundTripAndHistory(t *testing.T) {
	s, _, engine := fixture()
	ctx := context.Background()
	if s.Healthy(ctx) == nil {
		t.Fatal("reported protected before first snapshot")
	}
	first, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	original := append([]byte{}, engine.data...)
	engine.data = []byte("database after update and deletion")
	second, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := s.List(ctx)
	if err != nil || len(keys) != 2 || keys[0] != second {
		t.Fatalf("history: %v %v", keys, err)
	}
	if err := s.Restore(ctx, second, "target"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(engine.restored, engine.data) {
		t.Fatal("latest state not restored")
	}
	if err := s.Restore(ctx, first, "target"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(engine.restored, original) {
		t.Fatal("historical recovery point overwritten")
	}
	if err := s.Healthy(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestFailedBackupNeverPublishesRecoveryPoint(t *testing.T) {
	for _, reason := range []string{"dump", "corrupt-upload", "manifest", "empty"} {
		t.Run(reason, func(t *testing.T) {
			s, store, engine := fixture()
			switch reason {
			case "dump":
				engine.err = errors.New("dump failed")
			case "corrupt-upload":
				store.corruptUpload = true
			case "manifest":
				store.failManifest = true
			case "empty":
				engine.data = nil
			}
			if _, err := s.Snapshot(context.Background()); err == nil {
				t.Fatal("expected failure")
			}
			keys, _ := s.List(context.Background())
			if len(keys) != 0 {
				t.Fatal("published incomplete backup")
			}
			if s.Healthy(context.Background()) == nil {
				t.Fatal("failure reported healthy")
			}
		})
	}
}

func TestCorruptRestoreNeverTouchesDatabase(t *testing.T) {
	s, store, engine := fixture()
	ctx := context.Background()
	key, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	m, err := s.ReadManifest(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	store.objects[m.Archive][0] ^= 0xff
	if err := s.Restore(ctx, key, "target"); err == nil {
		t.Fatal("accepted corrupted backup")
	}
	if engine.restoreCalls != 0 {
		t.Fatal("touched database before checksum validation")
	}
}

func TestRejectForeignManifestAndArchive(t *testing.T) {
	s, store, _ := fixture()
	ctx := context.Background()
	if _, err := s.ReadManifest(ctx, "other/production/manifests/test.json"); err == nil {
		t.Fatal("accepted other environment")
	}
	key, err := s.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := s.ReadManifest(ctx, key)
	m.Archive = "other/production/secrets"
	store.objects[key], _ = json.Marshal(m)
	if err := s.Verify(ctx, key); err == nil {
		t.Fatal("accepted foreign archive")
	}
}

func TestBackupStaleness(t *testing.T) {
	s, _, _ := fixture()
	s.lastSnapshot = time.Now().Add(-3 * time.Hour)
	if err := s.Healthy(context.Background()); err == nil {
		t.Fatal("accepted stale backup")
	}
}

func TestToolConnectionDoesNotExposePassword(t *testing.T) {
	t.Setenv("PGSERVICE", "wrong-database")
	t.Setenv("PGHOST", "wrong-host")
	connection, env, err := toolConnection("postgres://alice:secret%40value@db.example/app?sslmode=require")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(connection, "secret") {
		t.Fatal("password in process argument")
	}
	if connection != "postgres://alice@db.example/app?sslmode=require" {
		t.Fatal(connection)
	}
	if !strings.Contains(strings.Join(env, "\n"), "PGPASSWORD=secret@value") {
		t.Fatal("password missing from environment")
	}
	for _, entry := range env {
		if strings.HasPrefix(entry, "PGSERVICE=") || strings.HasPrefix(entry, "PGHOST=") {
			t.Fatal("inherited routing settings")
		}
	}
}

func TestConfiguration(t *testing.T) {
	valid := Config{Enabled: true, Endpoint: "https://account.r2.cloudflarestorage.com", AccessKey: "id", SecretKey: "secret", Bucket: "private", Prefix: "app/production", Interval: 5 * time.Minute, Timeout: 15 * time.Minute, MaxAge: 30 * time.Minute}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"http://host", "https://user:password@host", "https://host/bucket", "https://host?token=secret"} {
		bad := valid
		bad.Endpoint = endpoint
		if bad.Validate() == nil {
			t.Fatal("accepted invalid endpoint")
		}
	}
	bad := valid
	bad.SecretKey = ""
	if bad.Validate() == nil {
		t.Fatal("accepted incomplete credentials")
	}
	for _, prefix := range []string{"", "../site", "/site", "site/../other", "."} {
		bad = valid
		bad.Prefix = prefix
		if bad.Validate() == nil {
			t.Fatal("accepted unsafe prefix")
		}
	}
	if (Config{Enabled: true}).Active() {
		t.Fatal("activated without credentials")
	}
}

func TestSingleDatabaseBackupDefaults(t *testing.T) {
	t.Setenv("R2_BACKUP_ENABLED", "false")
	t.Setenv("BACKUP_DATABASE_URL", "")
	t.Setenv("R2_BACKUP_INTERVAL", "")
	t.Setenv("R2_BACKUP_MAX_AGE", "")
	c, err := Load("primary")
	if err != nil || c.DatabaseURL != "primary" {
		t.Fatalf("backing up wrong database: %v", err)
	}
	if c.Interval != time.Hour || c.MaxAge != 2*time.Hour {
		t.Fatalf("incorrect hourly defaults: %v / %v", c.Interval, c.MaxAge)
	}
	t.Setenv("BACKUP_DATABASE_URL", "primary-direct")
	c, err = Load("primary")
	if err != nil || c.DatabaseURL != "primary-direct" {
		t.Fatalf("direct primary selection failed: %v", err)
	}
}

func TestRotationRetainsLatestTwoCompleteBatches(t *testing.T) {
	s, store, engine := fixture()
	ctx := context.Background()
	now := time.Now().UTC().Add(-3 * time.Hour)
	s.now = func() time.Time { return now }
	var keys []string
	var archives []string
	for i := 0; i < 4; i++ {
		engine.data = []byte{byte(i + 1)}
		key, err := s.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
		m, err := s.ReadManifest(ctx, key)
		if err != nil {
			t.Fatal(err)
		}
		archives = append(archives, m.Archive)
		now = now.Add(time.Hour)
	}
	got, _ := s.List(ctx)
	if len(got) != 2 || got[0] != keys[3] || got[1] != keys[2] {
		t.Fatalf("wrong retained batches: %v", got)
	}
	if len(store.objects) != 4 {
		t.Fatalf("want two archive/manifest pairs, got %d objects", len(store.objects))
	}
	for _, key := range append(keys[:2:2], archives[:2]...) {
		if _, ok := store.objects[key]; ok {
			t.Fatal("old batch object still exists")
		}
	}
	latest, err := s.Latest(ctx)
	if err != nil || latest != keys[3] {
		t.Fatal("did not choose newest batch")
	}
	if err := s.Restore(ctx, latest, "replacement-neon"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(engine.restored, []byte{4}) {
		t.Fatal("restored wrong batch")
	}
}

func TestFailedThirdBackupPreservesPreviousTwo(t *testing.T) {
	for _, failure := range []string{"dump", "upload", "manifest"} {
		t.Run(failure, func(t *testing.T) {
			s, store, engine := fixture()
			ctx := context.Background()
			for i := 0; i < 2; i++ {
				if _, err := s.Snapshot(ctx); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := s.List(ctx)
			switch failure {
			case "dump":
				engine.err = errors.New("neon unavailable")
			case "upload":
				store.corruptUpload = true
			case "manifest":
				store.failManifest = true
			}
			if _, err := s.Snapshot(ctx); err == nil {
				t.Fatal("expected failure")
			}
			after, _ := s.List(ctx)
			if strings.Join(before, ",") != strings.Join(after, ",") {
				t.Fatal("replaced a recovery point on failure")
			}
			for _, key := range before {
				if err := s.Verify(ctx, key); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRetentionDeleteFailureRetriesWithoutNewBackup(t *testing.T) {
	for _, object := range []string{"archive", "manifest"} {
		t.Run(object, func(t *testing.T) {
			s, store, _ := fixture()
			ctx := context.Background()
			first, err := s.Snapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			m, _ := s.ReadManifest(ctx, first)
			if object == "archive" {
				store.failDelete = m.Archive
			} else {
				store.failDelete = first
			}
			if _, err := s.Snapshot(ctx); err != nil {
				t.Fatal(err)
			}
			third, err := s.Snapshot(ctx)
			if err == nil || third == "" {
				t.Fatal("cleanup error must report the completed backup")
			}
			if s.Healthy(ctx) == nil {
				t.Fatal("did not report cleanup failure")
			}
			store.failDelete = ""
			key, delay, err := s.cycle(ctx)
			if err != nil || key != "" || delay <= 0 {
				t.Fatalf("retry created new batch or failed: %v", err)
			}
			keys, _ := s.List(ctx)
			if len(keys) != 2 || keys[0] != third || len(store.objects) != 4 {
				t.Fatal("cleanup retry failed")
			}
			if s.Healthy(ctx) != nil {
				t.Fatal("cleanup health did not recover")
			}
		})
	}
}

func TestRetentionNeverDeletesWhenKeeperIsCorrupt(t *testing.T) {
	s, store, _ := fixture()
	ctx := context.Background()
	first, _ := s.Snapshot(ctx)
	second, _ := s.Snapshot(ctx)
	m, _ := s.ReadManifest(ctx, second)
	store.objects[m.Archive][0] ^= 0xff
	if _, err := s.Snapshot(ctx); err == nil {
		t.Fatal("did not detect corrupt retained batch")
	}
	if err := s.Verify(ctx, first); err != nil {
		t.Fatal("removed older healthy recovery point")
	}
	keys, _ := s.List(ctx)
	if len(keys) != 3 {
		t.Fatal("pruned before confirming two healthy batches")
	}
}

func TestHourlyScheduleSurvivesRestartAndAccountsForUploadTime(t *testing.T) {
	s, store, engine := fixture()
	ctx := context.Background()
	now := time.Now().UTC()
	s.now = func() time.Time { return now }
	engine.afterDump = func() { now = now.Add(10 * time.Minute) }
	first, firstDelay, err := s.cycle(ctx)
	if err != nil || first == "" {
		t.Fatalf("initial backup failed: %v", err)
	}
	if firstDelay != 50*time.Minute {
		t.Fatalf("upload duration shifted schedule: %s", firstDelay)
	}
	now = now.Add(10 * time.Minute)
	restarted := New(s.config, store, engine)
	restarted.now = func() time.Time { return now }
	key, delay, err := restarted.cycle(ctx)
	if err != nil || key != "" || delay != 40*time.Minute {
		t.Fatalf("restart changed schedule: %s %s %v", key, delay, err)
	}
	if err := restarted.Healthy(ctx); err != nil {
		t.Fatal(err)
	}
	now = now.Add(40 * time.Minute)
	key, delay, err = restarted.cycle(ctx)
	if err != nil || key == "" || key == first || delay != 50*time.Minute {
		t.Fatalf("hourly backup failed: %v", err)
	}
}

func TestAbandonedUploadsAreCleanedButRecentUploadsArePreserved(t *testing.T) {
	s, store, _ := fixture()
	ctx := context.Background()
	now := time.Now().UTC().Add(-3 * time.Hour)
	s.now = func() time.Time { return now }
	// Simulate an ambiguous manifest request with an unlisted archive left behind.
	store.failManifest = true
	if _, err := s.Snapshot(ctx); err == nil {
		t.Fatal("expected manifest failure")
	}
	orphans, _ := store.List(ctx, s.config.Prefix+"/archives/")
	if len(orphans) != 1 {
		t.Fatal("missing abandoned upload")
	}
	store.failManifest = false
	now = now.Add(time.Hour)
	if _, err := s.Snapshot(ctx); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	// A recent incomplete upload must not be removed by cleanup.
	recent := s.config.Prefix + "/archives/" + now.Format("20060102T150405.000000000Z") + "-00000000-0000-4000-8000-000000000000.dump"
	store.objects[recent] = []byte("in flight")
	if _, err := s.Snapshot(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.objects[orphans[0]]; ok {
		t.Fatal("abandoned upload not cleaned")
	}
	if _, ok := store.objects[recent]; !ok {
		t.Fatal("removed recent upload")
	}
}

func TestRetentionIgnoresOtherObjects(t *testing.T) {
	s, store, _ := fixture()
	store.objects["other-site/archives/keep.dump"] = []byte("keep")
	store.objects[s.config.Prefix+"/manifests/notes.json"] = []byte("not a backup")
	for i := 0; i < 3; i++ {
		if _, err := s.Snapshot(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if string(store.objects["other-site/archives/keep.dump"]) != "keep" || store.objects[s.config.Prefix+"/manifests/notes.json"] == nil {
		t.Fatal("deleted unrelated object")
	}
}
