package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const baseUrl = "https://api.open-meteo.com"

func getWeather(ctx context.Context, lat float32, long float32) (*WeatherDto, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%v/v1/forecast?latitude=%v&longitude=%v&current=temperature_2m,wind_speed_10m", baseUrl, lat, long), bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("expect 200 status code but got %v", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var model WeatherDto
	err = json.Unmarshal(body, &model)
	if err != nil {
		return nil, err
	}
	return &model, nil
}

type WeatherDto struct {
	Latitude             float32 `json:"latitude"`
	Longitude            float32 `json:"longitude"`
	GenerationtimeMs     float32 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float32 `json:"elevation"`
	CurrentUnits         struct {
		Time          string `json:"time"`
		Interval      string `json:"interval"`
		Temperature2M string `json:"temperature_2m"`
		WindSpeed10M  string `json:"wind_speed_10m"`
	} `json:"current_units"`
	Current struct {
		Time          string  `json:"time"`
		Interval      int     `json:"interval"`
		Temperature2M float32 `json:"temperature_2m"`
		WindSpeed10M  float32 `json:"wind_speed_10m"`
	} `json:"current"`
}

type WeatherModel struct {
	Temperature       float32
	TemperatureUnists string
	WindSpeed         float32
	WindSpeedUnits    string
}

func getCachedWeather(ctx context.Context, lat float32, long float32) (*WeatherDto, error) {
	cacheKey := fmt.Sprintf("weather:%v:%v", lat, long)
	if val, err := redisClient.Get(ctx, cacheKey).Result(); err != nil {
		tmp, err := getWeather(ctx, lat, long)
		if err != nil {
			return nil, err
		}
		tmpBytes, err := json.Marshal(tmp)
		if err != nil {
			fmt.Println(err.Error())
		}
		_, err = redisClient.Set(ctx, cacheKey, string(tmpBytes), time.Duration(30)*time.Second).Result()
		if err != nil {
			fmt.Println(err.Error())
		}

		return tmp, nil
	} else {
		tmp := &WeatherDto{}
		if err := json.Unmarshal([]byte(val), tmp); err != nil {
			return nil, err
		}
		return tmp, nil
	}
}

func HandleGetWeather(c *gin.Context) {
	lat, err1 := strconv.ParseFloat(c.Query("lat"), 32)
	long, err2 := strconv.ParseFloat(c.Query("long"), 32)
	if err1 != nil || err2 != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if dto, err := getCachedWeather(c.Request.Context(), float32(lat), float32(long)); err == nil {
		c.JSON(http.StatusOK, WeatherModel{dto.Current.Temperature2M, dto.CurrentUnits.Temperature2M, dto.Current.WindSpeed10M, dto.CurrentUnits.WindSpeed10M})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
	}
}
