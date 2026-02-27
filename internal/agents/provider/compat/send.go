package compat

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-agent-skills/internal/agents/exec"
	atypes "github.com/pardnchiu/go-agent-skills/internal/agents/types"
	"github.com/pardnchiu/go-agent-skills/internal/skill"
	ttypes "github.com/pardnchiu/go-agent-skills/internal/tools/types"
	"github.com/pardnchiu/go-agent-skills/internal/utils"
)

func (a *Agent) Execute(ctx context.Context, skill *skill.Skill, userInput string, events chan<- atypes.Event, allowAll bool) error {
	return exec.Execute(ctx, a, a.workDir, skill, userInput, events, allowAll)
}

func (a *Agent) Send(ctx context.Context, messages []atypes.Message, tools []ttypes.Tool) (*atypes.Output, error) {
	chatAPI := a.baseURL + "/v1/chat/completions"

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if a.apiKey != "" {
		headers["Authorization"] = "Bearer " + a.apiKey
	}

	result, _, err := utils.POST[atypes.Output](ctx, a.httpClient, chatAPI, headers, map[string]any{
		"model":    a.model,
		"messages": messages,
		"tools":    tools,
	}, "json")
	if err != nil {
		return nil, fmt.Errorf("utils.POST: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("utils.POST: %s", result.Error.Message)
	}

	return &result, nil
}
