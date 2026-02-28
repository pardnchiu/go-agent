package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	agentTypes "github.com/pardnchiu/agenvoy/internal/agents/types"
)

func runEvents(_ context.Context, cancel context.CancelFunc, fn func(chan<- agentTypes.Event) error) error {
	start := time.Now()
	ch := make(chan agentTypes.Event, 16)
	var execErr error

	go func() {
		defer close(ch)
		execErr = fn(ch)
	}()

	skillNone := false
	for ev := range ch {
		switch ev.Type {
		case agentTypes.EventSkillSelect:
			fmt.Printf("[~] Selecting skill...")

		case agentTypes.EventSkillResult:
			if ev.Text == "none" {
				skillNone = true
				fmt.Printf("\033[2K\r[*] Skill: none")
			} else {
				fmt.Printf("\033[2K\r[*] Skill: %s\n", ev.Text)
			}

		case agentTypes.EventAgentSelect:
			if skillNone {
				fmt.Printf("\033[2K\r[~] Selecting agent...")
			} else {
				fmt.Printf("[~] Selecting agent...")
			}

		case agentTypes.EventAgentResult:
			fmt.Printf("\033[2K\r[*] Agent: %s\n", ev.Text)

		case agentTypes.EventText:
			fmt.Printf("[*] %s\n", ev.Text)

		case agentTypes.EventToolCall:
			printTool(ev)

		case agentTypes.EventToolCallStart, agentTypes.EventToolCallEnd:
			printHint("──────────────────────────────────────────────────")

		case agentTypes.EventToolCallText:
			printHint(strings.TrimSpace(ev.Text))

		case agentTypes.EventToolConfirm:
			prompt := promptui.Select{
				Label:        fmt.Sprintf("Run %s?", ev.ToolName),
				Items:        []string{"Yes", "Skip", "Stop"},
				Size:         3,
				HideSelected: true,
			}
			idx, _, err := prompt.Run()
			if err != nil || idx == 2 {
				fmt.Printf("[x] User stopped\n")
				cancel()
				ev.ReplyCh <- false
			} else if idx == 1 {
				fmt.Printf("[x] User skipped: %s\n", ev.ToolName)
				ev.ReplyCh <- false
			} else {
				ev.ReplyCh <- true
			}

		case agentTypes.EventToolSkipped:
			fmt.Printf("[x] Skipped: %s\n", ev.ToolName)

		case agentTypes.EventToolResult:
			fmt.Printf("[*] Result: %s\n", strings.TrimSpace(ev.Result))

		case agentTypes.EventError:
			if ev.Err != nil {
				fmt.Fprintf(os.Stderr, "[!] Error: %v\n", ev.Err)
			}

		case agentTypes.EventDone:
			fmt.Printf(" (%s)", time.Since(start).Round(time.Millisecond))
			fmt.Println()
		}
	}

	return execErr
}
