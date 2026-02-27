package exec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pardnchiu/go-agent-skills/internal/utils"
)

var (
	// matches trailing markdown JSON blocks: optional separator + optional label + ```json ... ```
	trailingJSONBlock = regexp.MustCompile(`(?s)\n*(?:---\s*\n)?(?:\*{0,2}[^\n*]*[Ss]ummary[^\n*]*\*{0,2}\s*\n)?` + "```" + `(?:json)?\s*(\{.*?\})\s*` + "```" + `\s*$`)
)

// isSummaryJSON checks if a parsed map contains summary-specific top-level keys.
func isSummaryJSON(m map[string]any) bool {
	summaryKeys := []string{"core_discussion", "discussion_log", "confirmed_needs", "current_conclusion"}
	matched := 0
	for _, k := range summaryKeys {
		if _, ok := m[k]; ok {
			matched++
		}
	}
	return matched >= 2
}

func extractSummary(configDir *utils.ConfigDirData, sessionID, value string) string {
	const summaryStart = "<!--SUMMARY_START-->"
	const summaryEnd = "<!--SUMMARY_END-->"

	var jsonData any
	var cleaned string

	// Primary: delimiter-wrapped summary
	start := strings.Index(value, summaryStart)
	end := strings.Index(value, summaryEnd)
	if start != -1 && end != -1 && end > start {
		jsonPart := strings.TrimSpace(value[start+len(summaryStart) : end])
		json.Unmarshal([]byte(jsonPart), &jsonData)
		cleaned = strings.TrimRight(value[:start], " \t\n\r")
	} else {
		// Fallback: strip any trailing markdown JSON block that looks like a summary
		if loc := trailingJSONBlock.FindStringSubmatchIndex(value); loc != nil {
			jsonPart := value[loc[2]:loc[3]]
			var m map[string]any
			if json.Unmarshal([]byte(jsonPart), &m) == nil && isSummaryJSON(m) {
				jsonData = m
				cleaned = strings.TrimRight(value[:loc[0]], " \t\n\r")
			}
		}
		if cleaned == "" {
			cleaned = value
		}
	}

	if jsonData != nil {
		path := filepath.Join(configDir.Home, sessionID, "summary.json")
		data, err := json.Marshal(jsonData)
		if err == nil {
			os.WriteFile(path, data, 0644)
		}
	}
	return cleaned
}
