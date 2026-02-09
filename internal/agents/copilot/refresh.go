package copilot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pardnchiu/go-agent-skills/internal/utils"
)

var (
	copilotTokenAPI = "https://api.github.com/copilot_internal/v2/token"
)

type RefreshToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func (c *Agent) checkExpires(ctx context.Context) error {
	if c.Token == nil || time.Now().After(c.Token.ExpiresAt.Add(-60*time.Second)) {
		return c.refresh(ctx)
	}
	return nil
}

func (c *Agent) refresh(ctx context.Context) error {
	token, code, err := utils.GET[RefreshToken](ctx, nil, copilotTokenAPI, map[string]string{
		"Authorization":  "token " + c.Token.AccessToken,
		"Accept":         "application/json",
		"Editor-Version": "vscode/1.95.0",
	})
	if code == http.StatusUnauthorized {
		return fmt.Errorf("token expired, please login again")
	}
	if code == http.StatusForbidden || code == http.StatusNotFound {
		return fmt.Errorf("token refresh failed, please login again")
	}
	if code != http.StatusOK {
		return fmt.Errorf("failed to refresh token, status code: %d", code)
	}
	if err != nil {
		return err
	}

	c.Refresh = &token

	return nil
}
