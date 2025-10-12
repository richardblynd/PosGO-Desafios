package services

import "weather-api/internal/models"

type TemperatureService struct{}

func NewTemperatureService() *TemperatureService {
	return &TemperatureService{}
}

func (ts *TemperatureService) ConvertTemperatures(celsius float64) models.WeatherResponse {
	fahrenheit := celsius*1.8 + 32
	kelvin := celsius + 273

	return models.WeatherResponse{
		TempC: celsius,
		TempF: fahrenheit,
		TempK: kelvin,
	}
}
