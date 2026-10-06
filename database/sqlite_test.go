package database

import (
	"path/filepath"
	"testing"
)

func TestNewSqliteEnablesWalAndBusyTimeout(t *testing.T) {
	db, err := NewSqlite(
		filepath.Join(t.TempDir(), "test.sqlite"),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	var journalMode string

	if err := db.Get(&journalMode, "PRAGMA journal_mode;"); err != nil {
		t.Fatal(err)
	}

	if journalMode != "wal" {
		t.Errorf("expected journal_mode wal, got %s", journalMode)
	}

	var busyTimeout int

	if err := db.Get(&busyTimeout, "PRAGMA busy_timeout;"); err != nil {
		t.Fatal(err)
	}

	if busyTimeout != 5000 {
		t.Errorf("expected busy_timeout 5000, got %d", busyTimeout)
	}
}
