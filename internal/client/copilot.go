package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type CopilotToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CopilotAgent struct {
	httpClient *http.Client
	Token      *CopilotToken
	Refresh    *RefreshToken
	workPath   string
}

func NewCopilot() (*CopilotAgent, error) {
	workDir, _ := os.Getwd()

	agent := &CopilotAgent{
		httpClient: &http.Client{},
		workPath:   workDir,
	}

	var token *CopilotToken

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".config", "go-agent-skills", "copilot_token.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// * if is not exist, then login
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			token, err = CopilotLogin(ctx, path)
			if err != nil {
				return nil, err
			}
			agent.Token = token
			return agent, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	agent.Token = token

	agent.checkAndRefresnToken(context.Background())

	return agent, nil
}
