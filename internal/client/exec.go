package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pardnchiu/go-agent-skills/internal/skill"
	"github.com/pardnchiu/go-agent-skills/internal/tools"
	"github.com/pardnchiu/go-agent-skills/internal/utils"
)

var (
	CopilotChatAPI = "https://api.githubcopilot.com/chat/completions"
)

type Message struct {
	Role      string `json:"role"`
	Content   any    `json:"content,omitempty"`
	ToolCalls []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type OpenAIOutput struct {
	Choices []struct {
		Message      Message `json:"message"`
		Delta        Message `json:"delta"`
		FinishReason string  `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (c *CopilotAgent) Execute(ctx context.Context, skill *skill.Skill, userInput string, output io.Writer) error {
	if err := c.checkExpires(ctx); err != nil {
		return err
	}

	if skill.Content == "" {
		return fmt.Errorf("SKILL.md is empty: %s", skill.Path)
	}

	exec, err := tools.NewExecutor(c.workDir)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	systemPrompt := systemPrompt(c.workDir, skill)
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userInput,
		},
	}

	for i := 0; i < MaxToolIterations; i++ {
		resp, err := c.sendChat(ctx, messages, exec.Tools)
		if err != nil {
			return err
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("no choices in response")
		}

		choice := resp.Choices[0]

		if len(choice.Message.ToolCalls) > 0 {
			messages = append(messages, choice.Message)

			for _, e := range choice.Message.ToolCalls {
				fmt.Printf("[*] Tool: %s\n", e.Function.Name)

				result, err := exec.Execute(e.Function.Name, json.RawMessage(e.Function.Arguments))
				if err != nil {
					result = "Error: " + err.Error()
				}

				fmt.Printf("\033[90m──────────────────────────────────────────────────\n")
				fmt.Printf("%s\n", strings.TrimSpace(result))
				fmt.Printf("──────────────────────────────────────────────────\033[0m\n")

				messages = append(messages, Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: e.ID,
				})
			}
			continue
		}

		switch v := choice.Message.Content.(type) {
		case string:
			if v != "" {
				output.Write([]byte(v))
				output.Write([]byte("\n"))
			}
		case nil:
		default:
			return fmt.Errorf("unexpected content type: %T", choice.Message.Content)
		}
		return nil
	}

	return fmt.Errorf("exceeded max iterations (%d)", MaxToolIterations)
}

func (c *CopilotAgent) sendChat(ctx context.Context, messages []Message, toolDefs []tools.Tool) (*OpenAIOutput, error) {
	result, _, error := utils.POSTJson[OpenAIOutput](ctx, c.httpClient, CopilotChatAPI, map[string]string{
		"Authorization":         "Bearer " + c.Refresh.Token,
		"Editor-Version":        "vscode/1.95.0",
		"Editor-Plugin-Version": "copilot/1.245.0",
		"Openai-Organization":   "github-copilot",
	}, map[string]any{
		"model":    CopilotDefaultModel,
		"messages": messages,
		"tools":    toolDefs,
	})
	if error != nil {
		return nil, fmt.Errorf("API request: %w", error)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s", result.Error.Message)
	}

	return &result, nil
}

func systemPrompt(workPath string, skill *skill.Skill) string {
	content := skill.Content
	for _, prefix := range []string{"scripts/", "templates/", "assets/"} {
		resolved := filepath.Join(skill.Path, prefix) + string(filepath.Separator)
		resolved = filepath.Clean(resolved) + string(filepath.Separator)
		content = strings.ReplaceAll(skill.Content, prefix, resolved)
	}
	return `你可以使用以下工具來與檔案系統互動：
- read_file(path): 讀取檔案內容
- list_files(path, recursive): 列出目錄內容
- glob_files(pattern): 依模式尋找檔案
- write_file(path, content): 寫入/建立檔案

工作目錄：` + workPath + `
技能目錄：` + skill.Path + `

關鍵：以下技能指令中的任何相對路徑都必須相對於技能目錄來解析。
重要：當被要求產生檔案時，你必須使用 write_file 工具將它們儲存到磁碟。

---

` + content
}
