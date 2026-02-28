package exec

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	agentTypes "github.com/pardnchiu/go-agent-skills/internal/agents/types"
	"github.com/pardnchiu/go-agent-skills/internal/skill"
)

//go:embed prompt/systemPrompt.md
var systemPrompt string

//go:embed prompt/skillExtension.md
var skillExtensionPrompt string

//go:embed prompt/summaryPrompt.md
var summaryPrompt string

var (
	MaxToolIterations  = 16
	MaxSkillIterations = 128
)

func Run(ctx context.Context, bot agentTypes.Agent, registry agentTypes.AgentRegistry, scanner *skill.Scanner, userInput string, events chan<- agentTypes.Event, allowAll bool) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	matchedSkill := selectSkill(ctx, bot, scanner, userInput)
	if matchedSkill != nil {
		events <- agentTypes.Event{Type: agentTypes.EventText, Text: fmt.Sprintf("Skill: %s", matchedSkill.Name)}
	}

	agent := registry.Fallback
	if chosen := selectAgent(ctx, bot, registry.Entries, userInput); chosen != "" {
		if a, ok := registry.Registry[chosen]; ok {
			agent = a
			events <- agentTypes.Event{Type: agentTypes.EventText, Text: chosen}
		}
	}

	return Execute(ctx, agent, workDir, matchedSkill, userInput, events, allowAll)
}
