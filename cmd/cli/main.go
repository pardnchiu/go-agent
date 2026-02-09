package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sort"

	"github.com/manifoldco/promptui"
	"github.com/pardnchiu/go-agent-skills/internal/agents"
	"github.com/pardnchiu/go-agent-skills/internal/agents/claude"
	"github.com/pardnchiu/go-agent-skills/internal/agents/copilot"
	"github.com/pardnchiu/go-agent-skills/internal/agents/nvidia"
	"github.com/pardnchiu/go-agent-skills/internal/agents/openai"
	"github.com/pardnchiu/go-agent-skills/internal/skill"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/cli/main.go list")
		fmt.Println("  go run cmd/cli/main.go run <skill_name> <input> [--allow]")
		os.Exit(1)
	}

	if os.Args[1] == "list" {
		scanner := skill.NewScanner()

		if len(scanner.Skills.ByName) == 0 {
			fmt.Println("No skills found")
			fmt.Println("\nScanned paths:")
			for _, path := range scanner.Skills.Paths {
				fmt.Printf("  - %s\n", path)
			}
			return
		}

		names := scanner.List()
		sort.Strings(names)

		fmt.Printf("Found %d skill(s):\n\n", len(names))
		for _, name := range names {
			s := scanner.Skills.ByName[name]
			fmt.Printf("• %s\n", name)
			if s.Description != "" {
				fmt.Printf("  %s\n", s.Description)
			}
			fmt.Printf("  Path: %s\n\n", s.Path)
		}
		return
	}

	if os.Args[1] == "run" {
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run cmd/cli/main.go run <skill_name> <input> [--allow]")
			os.Exit(1)
		}

		skillName := os.Args[2]
		userInput := os.Args[3]
		allowAll := slices.Contains(os.Args[4:], "--allow")

		agent := selectAgent()
		scanner := skill.NewScanner()
		targetSkill, ok := scanner.Skills.ByName[skillName]
		if !ok {
			slog.Error("skill not found", slog.String("name", skillName))
			os.Exit(1)
		}

		ctx := context.Background()
		if err := agent.Execute(ctx, targetSkill, userInput, os.Stdout, allowAll); err != nil {
			slog.Error("failed to execute skill", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return
	}

}

func selectAgent() agents.Agent {
	prompt := promptui.Select{
		Label: "Select Agent",
		Items: []string{
			"GitHub Copilot",
			"OpenAI",
			"Claude",
			"Nvidia",
		},
		HideSelected: true,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		slog.Error("agent selection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	switch idx {
	case 0:
		agent, err := copilot.New()
		if err != nil {
			slog.Error("failed to initialize Copilot", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return agent

	case 1:
		agent, err := openai.New()
		if err != nil {
			slog.Error("failed to initialize OpenAI", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return agent

	case 2:
		agent, err := claude.New()
		if err != nil {
			slog.Error("failed to initialize Anthropic", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return agent

	case 3:
		agent, err := nvidia.New()
		if err != nil {
			slog.Error("failed to initialize Anthropic", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return agent

	default:
		os.Exit(1)
		return nil
	}
}
