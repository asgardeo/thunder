package authz

import (
	"testing"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func TestMain(m *testing.M) {
	// This function is intentionally left empty.
	// Its purpose is to include the driver imports in the test binary for this package.
	m.Run()
}