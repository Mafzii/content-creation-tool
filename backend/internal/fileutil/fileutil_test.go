package fileutil

import (
	"strings"
	"testing"
	"testing/iotest"
)

func TestReadTextFile(t *testing.T) {
	t.Run("reads .txt file", func(t *testing.T) {
		content, err := ReadTextFile("notes.txt", strings.NewReader("hello world"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "hello world" {
			t.Errorf("got %q, want %q", content, "hello world")
		}
	})

	t.Run("reads .md file", func(t *testing.T) {
		content, err := ReadTextFile("README.md", strings.NewReader("# Title"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "# Title" {
			t.Errorf("got %q, want %q", content, "# Title")
		}
	})

	t.Run("case insensitive extension", func(t *testing.T) {
		content, err := ReadTextFile("file.TXT", strings.NewReader("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "data" {
			t.Errorf("got %q, want %q", content, "data")
		}
	})

	t.Run("rejects unsupported extension", func(t *testing.T) {
		_, err := ReadTextFile("image.png", strings.NewReader("data"))
		if err == nil {
			t.Fatal("expected error for .png file")
		}
		if !strings.Contains(err.Error(), "unsupported file type") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("rejects no extension", func(t *testing.T) {
		_, err := ReadTextFile("Makefile", strings.NewReader("data"))
		if err == nil {
			t.Fatal("expected error for file without extension")
		}
	})

	t.Run("reader error", func(t *testing.T) {
		_, err := ReadTextFile("file.txt", iotest.ErrReader(iotest.ErrTimeout))
		if err == nil {
			t.Fatal("expected error on reader failure")
		}
	})
}
