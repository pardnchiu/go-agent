package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"

	c "github.com/pardnchiu/go-agent-skills/internal/client"
	"github.com/pardnchiu/go-agent-skills/internal/skill"
)

func main() {
	client, err := c.NewCopilot()
	if err != nil {
		slog.Error("failed to load Copilot token",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	if os.Args[1] == "list" {
		scanner := skill.NewScanner()
		skillList, err := scanner.Scan()
		if err != nil {
			slog.Error("failed to scan skills", slog.String("error", err.Error()))
			os.Exit(1)
		}

		if len(skillList.ByName) == 0 {
			fmt.Println("No skills found")
			fmt.Println("\nScanned paths:")
			for _, path := range skillList.Paths {
				fmt.Printf("  - %s\n", path)
			}
			return
		}

		names := skillList.List()
		sort.Strings(names)

		fmt.Printf("Found %d skill(s):\n\n", len(names))
		for _, name := range names {
			s := skillList.ByName[name]
			fmt.Printf("• %s\n", name)
			if s.Description != "" {
				fmt.Printf("  %s\n", s.Description)
			}
			fmt.Printf("  Path: %s\n\n", s.Path)
		}
		return
	}

	slog.Info("successfully loaded Copilot token",
		slog.String("access_token", client.Token.AccessToken),
		slog.String("token_type", client.Token.TokenType),
		slog.String("scope", client.Token.Scope),
		slog.Time("expires_at", client.Token.ExpiresAt))

	if len(os.Args) < 3 || os.Args[1] != "input" {
		slog.Error("usage: go run cmd/cli/main.go input \"your message\"")
		os.Exit(1)
	}

	userInput := os.Args[2]
	ctx := context.Background()
	messages := []c.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: userInput,
		},
	}

	resp, err := client.SendChat(ctx, messages, nil)
	if err != nil {
		slog.Error("failed to send chat",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if content, ok := choice.Message.Content.(string); ok {
			fmt.Println("Response:", content)
		}
	}
}
