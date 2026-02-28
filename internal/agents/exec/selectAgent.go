package exec

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentTypes "github.com/pardnchiu/agenvoy/internal/agents/types"
	"github.com/pardnchiu/agenvoy/internal/utils"
)

//go:embed prompt/agentSelector.md
var agentSelectorPrompt string

func GetAgentEntries() []agentTypes.AgentEntry {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		return []agentTypes.AgentEntry{}
	}

	for _, dir := range configDir.Dirs {
		data, err := os.ReadFile(filepath.Join(dir, "config.json"))
		if err != nil {
			continue
		}
		var cfg struct {
			Models       []agentTypes.AgentEntry `json:"models"`
			DefaultModel string                  `json:"default_model"`
		}
		if json.Unmarshal(data, &cfg) != nil || len(cfg.Models) == 0 {
			continue
		}
		if cfg.DefaultModel != "" {
			for i, m := range cfg.Models {
				// * move default model to first be fallback
				if m.Name == cfg.DefaultModel {
					cfg.Models[0], cfg.Models[i] = cfg.Models[i], cfg.Models[0]
					break
				}
			}
		}
		return cfg.Models
	}
	return []agentTypes.AgentEntry{}
}

func selectAgent(ctx context.Context, bot agentTypes.Agent, agentEntries []agentTypes.AgentEntry, userInput string) string {
	trimInput := strings.TrimSpace(userInput)

	if len(agentEntries) == 0 {
		return ""
	}

	agentMap := make(map[string]struct{}, len(agentEntries))
	for _, a := range agentEntries {
		agentMap[a.Name] = struct{}{}
	}

	agentJson, err := json.Marshal(agentEntries)
	if err != nil {
		return ""
	}

	messages := []agentTypes.Message{
		{
			Role:    "system",
			Content: strings.TrimSpace(agentSelectorPrompt),
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				"Available agents:\n%s\nUser request: %s",
				string(agentJson),
				strings.TrimSpace(trimInput),
			),
		},
	}

	resp, err := bot.Send(ctx, messages, nil)
	if err != nil || len(resp.Choices) == 0 {
		return ""
	}

	answer := ""
	if content, ok := resp.Choices[0].Message.Content.(string); ok {
		answer = strings.Trim(strings.TrimSpace(content), "\"'` \n")
	}

	if answer == "NONE" || answer == "" {
		return ""
	}

	if _, ok := agentMap[answer]; ok {
		return answer
	}

	return ""
}
