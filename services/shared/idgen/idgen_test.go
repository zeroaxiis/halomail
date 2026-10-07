package idgen

import (
	"github.com/google/uuid"
	"strings"
	"testing"
	"time"
)

func TestNewReturnsUniqueUUIDv7(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := New()
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
		parsed, err := uuid.Parse(id)
		if err != nil {
			t.Fatalf("not a uuid: %q", id)
		}
		if parsed.Version() != 7 {
			t.Fatalf("version = %d, want 7", parsed.Version())
		}
	}
}

func TestPrefixed(t *testing.T) {
	id := Prefixed("evt_")
	if !strings.HasPrefix(id, "evt_") {
		t.Fatalf("missing prefix: %q", id)
	}
	if _, err := uuid.Parse(strings.TrimPrefix(id, "evt_")); err != nil {
		t.Fatalf("suffix is not a uuid: %q", id)
	}
}

func TestNewSortsByCreationTime(t *testing.T) {
	first := New()
	time.Sleep(2 * time.Millisecond)
	if second := New(); first >= second {
		t.Fatalf("ids out of order: %s then %s", first, second)
	}
}
