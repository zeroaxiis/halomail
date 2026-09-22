package idgen

import (
	"github.com/google/uuid"
	"testing"
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
