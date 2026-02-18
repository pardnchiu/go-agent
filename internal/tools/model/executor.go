package model

import "encoding/json"

type Executor struct {
	WorkPath       string
	Allowed        []string // limit to these folders to use
	AllowedCommand map[string]bool
	Exclude        []Exclude
	Tools          []Tool
}

type Exclude struct {
	File   string
	Negate bool
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
