package tools

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type Executor struct {
	WorkPath string
	Allowed  []string // limit to these folders to use
	Exclude  []string
	Tools    []Tool
}

//go:embed tools.json
var toolsJSON []byte

func NewExecutor(workPath string) (*Executor, error) {
	var tools []Tool
	if err := json.Unmarshal(toolsJSON, &tools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools: %w", err)
	}

	return &Executor{
		WorkPath: workPath,
		Exclude: []string{
			".DS_Store", ".git", "node_modules", "vendor", ".vscode", ".idea", "dist", "build",
		},
		Tools: tools,
	}, nil
}

func (e *Executor) Execute(name string, args json.RawMessage) (string, error) {
	switch name {
	case "read_file":
		var params struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal json (%s): %w", name, err)
		}
		return e.readFile(params.Path)

	case "list_files":
		var params struct {
			Path      string `json:"path"`
			Recursive bool   `json:"recursive"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal json (%s): %w", name, err)
		}
		return e.listFiles(params.Path, params.Recursive)

	case "glob_files":
		var params struct {
			Pattern string `json:"pattern"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal json (%s): %w", name, err)
		}
		return e.globFiles(params.Pattern)

	case "write_file":
		var params struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal json (%s): %w", name, err)
		}
		return e.writeFile(params.Path, params.Content)

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
