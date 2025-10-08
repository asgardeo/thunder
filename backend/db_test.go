package main_test

import (
	"testing"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func TestDBDrivers(t *testing.T) {
    // This test is to ensure the db drivers are imported.
}