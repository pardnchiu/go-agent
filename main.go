package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	CopilotClientID           = "Iv1.b507a08c87ecfe98"
	GitHubDeviceCodeAPI       = "https://github.com/login/device/code"
	GitHubOauthAccessTokenAPI = "https://github.com/login/oauth/access_token"
)

var ErrAuthorizationPending = fmt.Errorf("authorization pending") // pre declare error for ensuring padding wont cause login exit

func main() {
	_, token, err := NewCopilot()
	if err != nil {
		slog.Error("failed to load Copilot token",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("successfully loaded Copilot token",
		slog.String("access_token", token.AccessToken),
		slog.String("token_type", token.TokenType),
		slog.String("scope", token.Scope),
		slog.Time("expires_at", token.ExpiresAt))
}

type CopilotToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func NewCopilot() (string, *CopilotToken, error) {
	var token *CopilotToken

	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
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
				return "", nil, err
			}
			return path, token, nil
		}
		return "", nil, err
	}

	if err := json.Unmarshal(data, &token); err != nil {
		return "", nil, err
	}

	return path, token, nil
}

type GopilotDeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func CopilotLogin(ctx context.Context, tokenPath string) (*CopilotToken, error) {
	code, err := POSTForm[GopilotDeviceCode](ctx, nil, GitHubDeviceCodeAPI, url.Values{
		"client_id": {CopilotClientID},
	})

	expires := time.Now().Add(time.Duration(code.ExpiresIn) * time.Second)
	fmt.Printf("[*] url:      %-42s\n", code.VerificationURI)
	fmt.Printf("[*] code:     %-42s\n", code.UserCode)
	fmt.Printf("[*] expires:  %-36s\n", expires.Format("2006-01-02 15:04:05"))
	fmt.Print("[*] press Enter to open browser")

	go func() {
		var input string
		fmt.Scanln(&input)

		url := code.VerificationURI

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		default:
			fmt.Printf("[!] can not open browser, please open: %-48s\n", url)
		}

		if cmd != nil {
			cmd.Start()
		}
	}()

	interval := time.Duration(code.Interval) * time.Second
	deadline := time.Now().Add(time.Duration(code.ExpiresIn) * time.Second)

	var token *CopilotToken
	client := &http.Client{} // * use the same http client for reuse connection
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		token, err = GetAccessToken(ctx, client, code.DeviceCode)
		if err != nil {
			if errors.Is(err, ErrAuthorizationPending) {
				continue
			}
			return nil, err
		}

		path := filepath.Dir(tokenPath)
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, err
		}

		data, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			return nil, err
		}

		os.WriteFile(tokenPath, data, 0600)

		return token, nil
	}
	return nil, fmt.Errorf("device code expired")
}

type GopilotAccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

func GetAccessToken(ctx context.Context, client *http.Client, deviceCode string) (*CopilotToken, error) {
	accessToken, err := POSTForm[GopilotAccessToken](ctx, client, GitHubOauthAccessTokenAPI, url.Values{
		"client_id":   {CopilotClientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	})
	if err != nil {
		return nil, err
	}

	switch accessToken.Error {
	case "":
		return &CopilotToken{
			AccessToken: accessToken.AccessToken,
			TokenType:   accessToken.TokenType,
			Scope:       accessToken.Scope,
		}, nil
	case "authorization_pending":
		return nil, ErrAuthorizationPending
	default:
		return nil, fmt.Errorf("oauth error: %s", accessToken.Error)
	}
}

func POSTForm[T any](ctx context.Context, client *http.Client, api string, form url.Values) (T, error) {
	var result T

	if client == nil {
		client = &http.Client{}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", api, strings.NewReader(form.Encode()))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}
