package weatherReport

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pardnchiu/go-agent-skills/internal/utils"
)

const (
	apiPath = "https://wttr.in"
)

type responseData struct {
	CurrentCondition []struct {
		TempC          string `json:"temp_C"`
		FeelsLikeC     string `json:"FeelsLikeC"`
		Humidity       string `json:"humidity"`
		CloudCover     string `json:"cloudcover"`
		WindspeedKmph  string `json:"windspeedKmph"`
		WindDir16Point string `json:"winddir16Point"`
		Visibility     string `json:"visibility"`
		WeatherDesc    []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"current_condition"`
	NearestArea []struct {
		AreaName []struct {
			Value string `json:"value"`
		} `json:"areaName"`
		Country []struct {
			Value string `json:"value"`
		} `json:"country"`
	} `json:"nearest_area"`
	Weather []struct {
		Date     string `json:"date"`
		MaxTempC string `json:"maxtempC"`
		MinTempC string `json:"mintempC"`
		Hourly   []struct {
			TimeValue   string `json:"time"`
			TempC       string `json:"tempC"`
			WeatherDesc []struct {
				Value string `json:"value"`
			} `json:"weatherDesc"`
			ChanceOfRain string `json:"chanceofrain"`
		} `json:"hourly"`
	} `json:"weather"`
}

type WeatherResult struct {
	Location string        `json:"location"`
	Current  CurrentData   `json:"current"`
	Forecast []ForecastDay `json:"forecast"`
}

type CurrentData struct {
	TempC        int    `json:"temp_c"`
	FeelsLikeC   int    `json:"feels_like_c"`
	Humidity     int    `json:"humidity"`
	CloudCover   int    `json:"cloud_cover"`
	WindKmph     int    `json:"wind_kmph"`
	WindDir      string `json:"wind_dir"`
	VisibilityKm int    `json:"visibility_km"`
	Description  string `json:"description"`
}

type ForecastDay struct {
	Date     string       `json:"date"`
	MaxTempC int          `json:"max_temp_c"`
	MinTempC int          `json:"min_temp_c"`
	Hourly   []HourlyData `json:"hourly"`
}

type HourlyData struct {
	Hour        int    `json:"hour"`
	TempC       int    `json:"temp_c"`
	Description string `json:"description"`
	RainChance  int    `json:"rain_chance"`
}

func Fetch(city string, days, hourInterval int) (string, error) {
	city = strings.TrimSpace(city)

	if days < -1 || days > 3 || days == 0 {
		days = 3
	}

	if hourInterval <= 0 {
		hourInterval = 3
	}

	var requsetPath string
	if city != "" {
		requsetPath = fmt.Sprintf("%s/%s?format=j1", apiPath, city)
	} else {
		requsetPath = fmt.Sprintf("%s?format=j1", apiPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, _, err := utils.GET[responseData](ctx, nil, requsetPath, map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"Accept":     "application/json",
	})
	if err != nil {
		return "", fmt.Errorf("utils.GET: %w", err)
	}
	return parse(data, days, hourInterval)
}

func parse(data responseData, days, hourlyInterval int) (string, error) {
	if len(data.CurrentCondition) == 0 {
		return "", fmt.Errorf("no weather data")
	}

	cur := data.CurrentCondition[0]

	result := WeatherResult{}

	if len(data.NearestArea) > 0 {
		area := data.NearestArea[0]
		parts := []string{}
		if len(area.AreaName) > 0 {
			parts = append(parts, area.AreaName[0].Value)
		}
		if len(area.Country) > 0 {
			parts = append(parts, area.Country[0].Value)
		}
		result.Location = strings.Join(parts, ", ")
	}

	result.Current = CurrentData{
		TempC:        atoi(cur.TempC),
		FeelsLikeC:   atoi(cur.FeelsLikeC),
		Humidity:     atoi(cur.Humidity),
		CloudCover:   atoi(cur.CloudCover),
		WindKmph:     atoi(cur.WindspeedKmph),
		WindDir:      cur.WindDir16Point,
		VisibilityKm: atoi(cur.Visibility),
	}
	if len(cur.WeatherDesc) > 0 {
		result.Current.Description = cur.WeatherDesc[0].Value
	}

	if days != -1 {
		for i, day := range data.Weather {
			if i >= days {
				break
			}
			fd := ForecastDay{
				Date:     day.Date,
				MaxTempC: atoi(day.MaxTempC),
				MinTempC: atoi(day.MinTempC),
			}
			for _, h := range day.Hourly {
				t := atoi(h.TimeValue) / 100
				if t%hourlyInterval != 0 {
					continue
				}
				hd := HourlyData{
					Hour:       t,
					TempC:      atoi(h.TempC),
					RainChance: atoi(h.ChanceOfRain),
				}
				if len(h.WeatherDesc) > 0 {
					hd.Description = h.WeatherDesc[0].Value
				}
				fd.Hourly = append(fd.Hourly, hd)
			}
			result.Forecast = append(result.Forecast, fd)
		}
	}

	out, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	return string(out), nil
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
