package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	agentTypes "github.com/pardnchiu/go-agent-skills/internal/agents/types"
	"github.com/pardnchiu/go-agent-skills/internal/tools"
	toolTypes "github.com/pardnchiu/go-agent-skills/internal/tools/types"
)

func toolCall(ctx context.Context, exec *toolTypes.Executor, choice agentTypes.OutputChoices, sessionData *agentTypes.AgentSession, events chan<- agentTypes.Event, allowAll bool, alreadyCall map[string]string) (*agentTypes.AgentSession, map[string]string, error) {
	sessionData.Messages = append(sessionData.Messages, choice.Message)

	for _, tool := range choice.Message.ToolCalls {
		toolName := tool.Function.Name
		toolArg := tool.Function.Arguments

		hash := fmt.Sprintf("%v|%v", toolName, toolArg)
		if cached, ok := alreadyCall[hash]; ok && cached != "" {
			sessionData.Messages = append(sessionData.Messages, agentTypes.Message{
				Role:       "tool",
				Content:    cached,
				ToolCallID: tool.ID,
			})
			continue
		}

		if idx := strings.Index(toolName, "<|"); idx != -1 {
			toolName = toolName[:idx]
		}

		events <- agentTypes.Event{
			Type:     agentTypes.EventToolCall,
			ToolName: toolName,
			ToolArgs: tool.Function.Arguments,
			ToolID:   tool.ID,
		}

		if !allowAll {
			replyCh := make(chan bool, 1)
			events <- agentTypes.Event{
				Type:     agentTypes.EventToolConfirm,
				ToolName: toolName,
				ToolArgs: tool.Function.Arguments,
				ToolID:   tool.ID,
				ReplyCh:  replyCh,
			}
			proceed := <-replyCh
			if !proceed {
				events <- agentTypes.Event{
					Type:     agentTypes.EventToolSkipped,
					ToolName: toolName,
					ToolID:   tool.ID,
				}
				sessionData.Tools = append(sessionData.Tools, agentTypes.Message{
					Role:       "tool",
					Content:    "Skipped by user",
					ToolCallID: tool.ID,
				})
				sessionData.Messages = append(sessionData.Messages, agentTypes.Message{
					Role:       "tool",
					Content:    "Skipped by user",
					ToolCallID: tool.ID,
				})
				continue
			}
		}

		result, err := tools.Execute(ctx, exec, toolName, json.RawMessage(tool.Function.Arguments))
		if err != nil {
			result = "no data"
		}

		content := fmt.Sprintf("[%s] %s", toolName, result)
		alreadyCall[hash] = content

		events <- agentTypes.Event{
			Type:     agentTypes.EventToolResult,
			ToolName: toolName,
			ToolID:   tool.ID,
			Result:   result,
		}
		sessionData.Tools = append(sessionData.Tools, agentTypes.Message{
			Role:       "tool",
			Content:    content,
			ToolCallID: tool.ID,
		})
		sessionData.Messages = append(sessionData.Messages, agentTypes.Message{
			Role:       "tool",
			Content:    content,
			ToolCallID: tool.ID,
		})
	}
	return sessionData, alreadyCall, nil
}
