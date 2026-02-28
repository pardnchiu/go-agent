package exec

import (
	"context"
	"fmt"
	"os"
	"strings"

	agentTypes "github.com/pardnchiu/agenvoy/internal/agents/types"
	"github.com/pardnchiu/agenvoy/internal/skill"
)

func Run(ctx context.Context, bot agentTypes.Agent, registry agentTypes.AgentRegistry, scanner *skill.Scanner, userInput string, events chan<- agentTypes.Event, allowAll bool) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	trimInput := strings.TrimSpace(userInput)

	events <- agentTypes.Event{
		Type: agentTypes.EventSkillSelect,
	}
	matchedSkill := selectSkill(ctx, bot, scanner, trimInput)
	if matchedSkill != nil {
		events <- agentTypes.Event{
			Type: agentTypes.EventSkillResult,
			Text: strings.TrimSpace(matchedSkill.Name),
		}
	} else {
		events <- agentTypes.Event{
			Type: agentTypes.EventSkillResult,
			Text: "none",
		}
	}

	events <- agentTypes.Event{
		Type: agentTypes.EventAgentSelect,
	}
	// * default is fallback
	agent := registry.Fallback
	if chosen := selectAgent(ctx, bot, registry.Entries, trimInput); chosen != "" {
		if a, ok := registry.Registry[chosen]; ok {
			agent = a
		}
		events <- agentTypes.Event{
			Type: agentTypes.EventAgentResult,
			Text: strings.TrimSpace(chosen),
		}
	} else {
		events <- agentTypes.Event{
			Type: agentTypes.EventAgentResult,
			Text: "fallback",
		}
	}

	return Execute(ctx, agent, workDir, matchedSkill, trimInput, events, allowAll)
}
