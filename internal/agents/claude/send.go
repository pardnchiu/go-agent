package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pardnchiu/go-agent-skills/internal/agents"
	"github.com/pardnchiu/go-agent-skills/internal/skill"
	t "github.com/pardnchiu/go-agent-skills/internal/tools"
	"github.com/pardnchiu/go-agent-skills/internal/utils"
)

var (
	defaultModel = "claude-sonnet-4-5"
	messagesAPI  = "https://api.anthropic.com/v1/messages"
	maxTokens    = 8192
)

func (a *Agent) Execute(ctx context.Context, skill *skill.Skill, userInput string, output io.Writer, allowAll bool) error {
	return agents.Execute(ctx, a, a.workDir, skill, userInput, output, allowAll)
}

func (a *Agent) Send(ctx context.Context, messages []agents.Message, tools []t.Tool) (*agents.OpenAIOutput, error) {
	var systemPrompt string
	var newMessages []map[string]any

	for _, msg := range messages {
		if msg.Role == "system" {
			if content, ok := msg.Content.(string); ok {
				systemPrompt = content
			}
			continue
		}

		message := a.convertToMessage(msg)
		newMessages = append(newMessages, message)
	}

	newTools := make([]map[string]any, len(tools))
	for i, tool := range tools {
		newTools[i] = map[string]any{
			"name":         tool.Function.Name,
			"description":  tool.Function.Description,
			"input_schema": json.RawMessage(tool.Function.Parameters),
		}
	}

	result, _, err := utils.POSTJson[map[string]any](ctx, a.httpClient, messagesAPI, map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
		"Content-Type":      "application/json",
	}, map[string]any{
		"model":      defaultModel,
		"max_tokens": maxTokens,
		"system":     systemPrompt,
		"messages":   newMessages,
		"tools":      newTools,
	})
	if err != nil {
		return nil, fmt.Errorf("API request: %w", err)
	}

	if errObj, ok := result["error"].(map[string]any); ok {
		return nil, fmt.Errorf("API error: %s", errObj["message"])
	}

	return a.convertToOutput(result), nil
}

func (a *Agent) convertToMessage(message agents.Message) map[string]any {
	newMessage := map[string]any{}
	if message.ToolCallID != "" {
		newMessage["role"] = "user"
		newMessage["content"] = []map[string]any{
			{
				"type":        "tool_result",
				"tool_use_id": message.ToolCallID,
				"content":     message.Content,
			},
		}
		return newMessage
	}

	newMessage["role"] = message.Role

	if len(message.ToolCalls) > 0 {
		var content []map[string]any
		for _, tool := range message.ToolCalls {
			var input map[string]any
			json.Unmarshal([]byte(tool.Function.Arguments), &input)
			content = append(content, map[string]any{
				"type":  "tool_use",
				"id":    tool.ID,
				"name":  tool.Function.Name,
				"input": input,
			})
		}
		newMessage["content"] = content
		return newMessage
	}

	newMessage["content"] = message.Content
	return newMessage
}

func (a *Agent) convertToOutput(resp map[string]any) *agents.OpenAIOutput {
	output := &agents.OpenAIOutput{
		Choices: make([]struct {
			Message      agents.Message `json:"message"`
			Delta        agents.Message `json:"delta"`
			FinishReason string         `json:"finish_reason,omitempty"`
		}, 1),
	}

	content, ok := resp["content"].([]any)
	if !ok {
		return output
	}

	var toolCalls []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	}
	var textContent string

	for _, e := range content {
		item, ok := e.(map[string]any)
		if !ok {
			continue
		}

		itemType, _ := item["type"].(string)
		switch itemType {
		case "text":
			if text, ok := item["text"].(string); ok {
				textContent = text
			}

		case "tool_use":
			args, _ := json.Marshal(item["input"])
			id, _ := item["id"].(string)
			name, _ := item["name"].(string)

			toolCall := struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			}{
				ID:   id,
				Type: "function",
			}
			toolCall.Function.Name = name
			toolCall.Function.Arguments = string(args)
			toolCalls = append(toolCalls, toolCall)
		}
	}

	output.Choices[0].Message = agents.Message{
		Role:      "assistant",
		Content:   textContent,
		ToolCalls: toolCalls,
	}

	return output
}
