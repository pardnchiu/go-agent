package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toolTypes "github.com/pardnchiu/go-agent-skills/internal/tools/types"
)

func read(e *toolTypes.Executor, path string) (string, error) {
	fullPath := getFullPath(e, path)

	if isExclude(e, fullPath) {
		return "", fmt.Errorf("path is excluded: %s", path)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file (%s): %w", path, err)
	}
	return string(data), nil
}

func getFullPath(e *toolTypes.Executor, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(e.WorkPath, path)
}

func isExclude(e *toolTypes.Executor, path string) bool {
	excluded := false
	for _, e := range e.Exclude {
		match, err := filepath.Match(e.File, filepath.Base(path))
		if err != nil {
			continue
		}

		if !match {
			match = strings.Contains(path, "/"+e.File+"/") ||
				strings.HasPrefix(path, e.File+"/")
		}
		if match {
			excluded = !e.Negate
		}
	}
	return excluded
}
