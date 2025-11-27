package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"weather-api/internal/models"
)

type WeatherAPIService struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewWeatherAPIService() *WeatherAPIService {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		apiKey = "demo_key"
	}

	log.Println("Using Weather API Key: ", apiKey)

	return &WeatherAPIService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "http://api.weatherapi.com/v1",
		apiKey:  apiKey,
	}
}

func (w *WeatherAPIService) GetCurrentWeather(city string) (*models.WeatherAPIResponse, error) {
	endpoint := fmt.Sprintf("%s/current.json", w.baseURL)

	params := url.Values{}
	params.Add("key", w.apiKey)
	params.Add("q", city)
	params.Add("aqi", "no")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := w.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WeatherAPI returned status %d", resp.StatusCode)
	}

	var weatherResp models.WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, err
	}

	return &weatherResp, nil
}
