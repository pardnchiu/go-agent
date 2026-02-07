package main

import (
	"log/slog"
	"os"

	"github.com/pardnchiu/go-agent-skills/internal/client"
)

func main() {
	client, err := client.NewCopilot()
	if err != nil {
		slog.Error("failed to load Copilot token",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("successfully loaded Copilot token",
		slog.String("access_token", client.Token.AccessToken),
		slog.String("token_type", client.Token.TokenType),
		slog.String("scope", client.Token.Scope),
		slog.Time("expires_at", client.Token.ExpiresAt))
}
