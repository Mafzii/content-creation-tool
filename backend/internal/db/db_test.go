package db

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitSQL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"single statement", "CREATE TABLE foo (id INT)", 1},
		{"two statements", "CREATE TABLE foo (id INT); CREATE TABLE bar (id INT)", 2},
		{"trailing semicolon", "CREATE TABLE foo (id INT);", 1},
		{"whitespace-only segment skipped", "CREATE TABLE foo (id INT);  ;  ", 1},
		{"empty input", "", 0},
		{"multiline", "CREATE TABLE foo (\n  id INT\n);\nCREATE TABLE bar (\n  id INT\n);", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitSQL(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitSQL(%q) returned %d statements, want %d: %v", tt.input, len(got), tt.want, got)
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"  hello  ", "hello"},
		{"\n\thello\r\n", "hello"},
		{"", ""},
		{"   ", ""},
		{"no-trim", "no-trim"},
	}
	for _, tt := range tests {
		got := trimSpace(tt.input)
		if got != tt.want {
			t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsDuplicateColumnErr(t *testing.T) {
	if isDuplicateColumnErr(nil) {
		t.Error("nil error should return false")
	}
	if !isDuplicateColumnErr(errors.New("duplicate column name: foo")) {
		t.Error("should detect 'duplicate column name'")
	}
	if !isDuplicateColumnErr(errors.New("column already exists")) {
		t.Error("should detect 'already exists'")
	}
	if isDuplicateColumnErr(errors.New("syntax error")) {
		t.Error("unrelated error should return false")
	}
}

func TestOpen(t *testing.T) {
	origDir, _ := os.Getwd()
	backendDir := filepath.Join(origDir, "..", "..")
	os.Chdir(backendDir)
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "sub", "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='topics'").Scan(&name)
	if err != nil {
		t.Fatalf("topics table not created: %v", err)
	}

	var journalMode string
	db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var fk int
	db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	db2.Close()
}
