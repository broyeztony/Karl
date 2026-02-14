package interpreter

import (
	"os"
	"path/filepath"
	"strings"
)

func (e *Evaluator) resolveImportPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Relative imports ("./" and "../") resolve relative to the importing file,
	// not the current working directory. This makes example programs runnable
	// from the repo root without `cd` into the module directory.
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		if e.filename != "" && e.filename != "<stdin>" {
			return filepath.Join(filepath.Dir(e.filename), path), nil
		}
	}
	root := e.projectRoot
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		root = cwd
	}
	return filepath.Join(root, path), nil
}
