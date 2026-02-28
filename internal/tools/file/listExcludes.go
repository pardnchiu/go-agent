package file

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	toolTypes "github.com/pardnchiu/agenvoy/internal/tools/types"
)

//go:embed embed/exclude.json
var excludeFiles []byte

func ListExcludes(root string) []toolTypes.Exclude {
	var defaults []string
	if err := json.Unmarshal(excludeFiles, &defaults); err != nil {
		slog.Warn("failed to unmarshal exclude files, using empty list",
			slog.String("error", err.Error()))
	}

	newFiles := make([]toolTypes.Exclude, 0, len(defaults))
	for _, line := range defaults {
		if ef, ok := checkLine(line); ok {
			newFiles = append(newFiles, ef)
		}
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return newFiles
	}

	for _, entry := range entries {
		// * to fit file name like .*ignore
		name := entry.Name()
		if entry.IsDir() ||
			!strings.HasSuffix(name, "ignore") ||
			!strings.HasPrefix(name, ".") {
			continue
		}

		newFiles = append(newFiles, parseIgnore(filepath.Join(root, name))...)
	}

	return newFiles
}

func parseIgnore(path string) []toolTypes.Exclude {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var files []toolTypes.Exclude
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if ef, ok := checkLine(scanner.Text()); ok {
			files = append(files, ef)
		}
	}

	return files
}

func checkLine(raw string) (toolTypes.Exclude, bool) {
	line := strings.TrimSpace(raw)
	if line == "" || strings.HasPrefix(line, "#") {
		return toolTypes.Exclude{}, false
	}

	negate := false
	if strings.HasPrefix(line, "!") {
		negate = true
		line = strings.TrimPrefix(line, "!")
	}

	line = strings.TrimPrefix(line, "/")
	line = strings.TrimSuffix(line, "/")
	if line == "" {
		return toolTypes.Exclude{}, false
	}

	return toolTypes.Exclude{
		File:   line,
		Negate: negate,
	}, true
}
