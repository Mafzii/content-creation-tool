package fileutil

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

var allowedExtensions = map[string]bool{
	".txt": true,
	".md":  true,
}

// ReadTextFile reads from an io.Reader and validates that the filename has an allowed extension (.txt, .md).
func ReadTextFile(filename string, r io.Reader) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("unsupported file type %q; only .txt and .md are allowed", ext)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}
