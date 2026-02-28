package apis

import (
	"encoding/json"
	"fmt"

	"github.com/pardnchiu/go-agent-skills/internal/tools/apiAdapter"
	"github.com/pardnchiu/go-agent-skills/internal/tools/apis/googleRSS"
	"github.com/pardnchiu/go-agent-skills/internal/tools/apis/weatherReport"
	"github.com/pardnchiu/go-agent-skills/internal/tools/apis/yahooFinance"
	toolTypes "github.com/pardnchiu/go-agent-skills/internal/tools/types"
)

func Routes(e *toolTypes.Executor, name string, args json.RawMessage) (string, error) {
	switch name {
	case "send_http_request":
		var params struct {
			URL         string            `json:"url"`
			Method      string            `json:"method"`
			Headers     map[string]string `json:"headers"`
			Body        map[string]any    `json:"body"`
			ContentType string            `json:"content_type"`
			Timeout     int               `json:"timeout"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("json.Unmarshal: %w", err)
		}
		return apiAdapter.Send(params.URL, params.Method, params.Headers, params.Body, params.ContentType, params.Timeout)

	case "fetch_yahoo_finance":
		var params struct {
			Symbol   string `json:"symbol"`
			Interval string `json:"interval"`
			Range    string `json:"range"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("json.Unmarshal: %w", err)
		}
		return yahooFinance.Fetch(params.Symbol, params.Interval, params.Range)

	case "fetch_google_rss":
		var params struct {
			Keyword string `json:"keyword"`
			Time    string `json:"time"`
			Lang    string `json:"lang"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("json.Unmarshal: %w", err)
		}
		return googleRSS.Fetch(params.Keyword, params.Time, params.Lang)

	case "fetch_weather":
		var params struct {
			City           string      `json:"city"`
			Days           int         `json:"days"`
			HourlyInterval json.Number `json:"hourly_interval"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal json (%s): %w", name, err)
		}
		hourlyInterval, _ := params.HourlyInterval.Int64()
		return weatherReport.Fetch(params.City, params.Days, int(hourlyInterval))

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
