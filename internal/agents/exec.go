package agents

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pardnchiu/go-agent-skills/internal/skill"
	"github.com/pardnchiu/go-agent-skills/internal/tools"
)

//go:embed sysprompt.md
var sysPrompt string

var (
	MaxToolIterations = 128
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

type Agent interface {
	Send(ctx context.Context, messages []Message, toolDefs []tools.Tool) (*OpenAIOutput, error)
	Execute(ctx context.Context, skill *skill.Skill, userInput string, output io.Writer, allowAll bool) error
}

func Execute(ctx context.Context, agent Agent, workDir string, skill *skill.Skill, userInput string, output io.Writer, allowAll bool) error {
	if skill.Content == "" {
		return fmt.Errorf("SKILL.md is empty: %s", skill.Path)
	}

	exec, err := tools.NewExecutor(workDir)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	systemPrompt := systemPrompt(workDir, skill)
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
		resp, err := agent.Send(ctx, messages, exec.Tools)
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

				if !allowAll {
					var args map[string]any
					if err := json.Unmarshal([]byte(e.Function.Arguments), &args); err == nil {
						fmt.Printf("\033[90m──────────────────────────────────────────────────\n")
						for k, v := range args {
							fmt.Printf("- %s: %v\n", k, v)
						}
					} else {
						fmt.Printf("\033[90m──────────────────────────────────────────────────\n")
						fmt.Printf("- %s\n", e.Function.Arguments)
					}
					prompt := promptui.Select{
						Label: "Continue?",
						Items: []string{
							"Yes",
							"Cancel",
						},
						Size:         2,
						HideSelected: true,
					}

					idx, _, err := prompt.Run()
					if err != nil {
						fmt.Printf("[x] Prompt error: %v\n", err)
						continue
					}

					if idx == 1 {
						fmt.Printf("[x] User cancelled\n")
						messages = append(messages, Message{
							Role:       "tool",
							Content:    "User cancelled",
							ToolCallID: e.ID,
						})
						continue
					}
				}

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

func systemPrompt(workPath string, skill *skill.Skill) string {
	content := skill.Content

	for _, prefix := range []string{"scripts/", "templates/", "assets/"} {
		resolved := filepath.Join(skill.Path, prefix)

		if _, err := os.Stat(resolved); err == nil {
			content = strings.ReplaceAll(content, prefix, resolved+string(filepath.Separator))
		}
	}

	return strings.NewReplacer(
		"{{.WorkPath}}", workPath,
		"{{.SkillPath}}", skill.Path,
		"{{.Content}}", content,
	).Replace(sysPrompt)
}
