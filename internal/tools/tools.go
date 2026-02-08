package tools

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func (e *Executor) getFullPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(e.WorkPath, path)
}

func (e *Executor) isExclude(path string) bool {
	for _, pattern := range e.Exclude {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}

		if strings.Contains(path, "/"+pattern+"/") ||
			strings.HasPrefix(path, pattern+"/") {
			return true
		}
	}
	return false
}

func (e *Executor) readFile(path string) (string, error) {
	fullPath := e.getFullPath(path)

	if e.isExclude(fullPath) {
		return "", fmt.Errorf("path is excluded: %s", path)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file (%s): %w", path, err)
	}
	return string(data), nil
}

func (e *Executor) listFiles(path string, recursive bool) (string, error) {
	fullPath := e.getFullPath(path)

	var result strings.Builder
	if recursive {
		err := filepath.Walk(fullPath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				slog.Warn("failed to access path",
					slog.String("error", err.Error()))
				return nil
			}

			if e.isExclude(p) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			relPath, err := filepath.Rel(fullPath, p)
			if err != nil {
				slog.Warn("failed to get relative path",
					slog.String("error", err.Error()))
				return nil
			}
			if relPath == "." {
				return nil
			}
			if strings.HasPrefix(filepath.Base(p), ".") && info.IsDir() {
				return filepath.SkipDir
			}
			if info.IsDir() {
				result.WriteString(relPath + "/\n")
			} else {
				result.WriteString(relPath + "\n")
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("failed to walk directory (%s): %w", path, err)
		}
	} else {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read directory (%s): %w", path, err)
		}
		for _, entry := range entries {
			entryPath := filepath.Join(fullPath, entry.Name())
			if e.isExclude(entryPath) {
				continue
			}

			if entry.IsDir() {
				result.WriteString(entry.Name() + "/\n")
			} else {
				result.WriteString(entry.Name() + "\n")
			}
		}
	}

	return result.String(), nil
}

// * just fit one level of glob pattern
// TODO: need to do more work to support complex glob patterns
func (e *Executor) globFiles(pattern string) (string, error) {
	var result strings.Builder
	err := filepath.WalkDir(e.WorkPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			slog.Warn("failed to access path",
				slog.String("error", err.Error()))
			return nil
		}

		if e.isExclude(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != e.WorkPath {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(e.WorkPath, path)
		if err != nil {
			slog.Warn("failed to get relative path",
				slog.String("error", err.Error()))
			return nil
		}

		matched, err := filepath.Match(pattern, relPath)
		if err != nil {
			slog.Warn("failed to match pattern",
				slog.String("error", err.Error()))
			return nil
		}
		if matched {
			result.WriteString(relPath + "\n")
			return nil
		}

		if strings.Contains(pattern, "**") {
			parts := strings.SplitN(pattern, "**", 2)
			prefix := parts[0]
			suffix := strings.TrimPrefix(parts[1], "/")
			if strings.HasPrefix(relPath, prefix) {
				rest := relPath[len(prefix):]
				if suffix == "" {
					result.WriteString(relPath + "\n")
				} else if matched, _ := filepath.Match(suffix, filepath.Base(rest)); matched {
					result.WriteString(relPath + "\n")
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk directory (%s): %w", pattern, err)
	}

	if result.Len() == 0 {
		return fmt.Sprintf("No fils found: %s", pattern), nil
	}
	return result.String(), nil
}

func (e *Executor) writeFile(path, content string) (string, error) {
	fullPath := e.getFullPath(path)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory (%s): %w", path, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file (%s): %w", path, err)
	}

	return fmt.Sprintf("Successfully wrote file: %s", path), nil
}
