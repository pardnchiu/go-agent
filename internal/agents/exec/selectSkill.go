package exec

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	agentTypes "github.com/pardnchiu/go-agent-skills/internal/agents/types"
	"github.com/pardnchiu/go-agent-skills/internal/skill"
)

//go:embed prompt/skillSelector.md
var skillSelectorPrompt string

func selectSkill(ctx context.Context, bot agentTypes.Agent, scanner *skill.Scanner, userInput string) *skill.Skill {
	skills := scanner.List()
	if len(skills) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, skill := range skills {
		s := scanner.Skills.ByName[skill]
		sb.WriteString(fmt.Sprintf("- %s: %s\n", skill, s.Description))
	}

	messages := []agentTypes.Message{
		{Role: "system", Content: skillSelectorPrompt},
		{
			Role:    "user",
			Content: fmt.Sprintf("Available skills:\n%s\nUser request: %s", sb.String(), userInput),
		},
	}

	resp, err := bot.Send(ctx, messages, nil)
	if err != nil || len(resp.Choices) == 0 {
		return nil
	}

	answer := ""
	if content, ok := resp.Choices[0].Message.Content.(string); ok {
		answer = strings.Trim(strings.TrimSpace(content), "\"'` \n")
	}

	if answer == "NONE" || answer == "" {
		return nil
	} else if s, ok := scanner.Skills.ByName[answer]; ok {
		return s
	}

	return nil
}
