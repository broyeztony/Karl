package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoSpawnDirectForInSnippets(t *testing.T) {
	roots := []string{
		filepath.Join("..", "examples"),
		filepath.Join("..", ".playground"),
	}

	var files []string
	for _, root := range roots {
		info, err := os.Stat(root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("stat %s: %v", root, err)
		}
		if !info.IsDir() {
			continue
		}
		files = append(files, listKarlFiles(t, root)...)
	}

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lines := strings.Split(string(data), "\n")
		for i, raw := range lines {
			line := raw
			if idx := strings.Index(line, "//"); idx >= 0 {
				line = line[:idx]
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.Contains(trimmed, "& for ") || strings.Contains(trimmed, "spawn for ") {
				t.Fatalf("%s:%d uses invalid spawn target syntax; wrap for-expr in a call: & (() -> for ... then ... )()", path, i+1)
			}
		}
	}
}
