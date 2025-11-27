package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"weather-api/internal/models"
	"weather-api/internal/services"

	"github.com/gorilla/mux"
)

type WeatherHandler struct {
	viaCEPService      *services.ViaCEPService
	weatherAPIService  *services.WeatherAPIService
	temperatureService *services.TemperatureService
	validationService  *services.ValidationService
}

func NewWeatherHandler() *WeatherHandler {
	return &WeatherHandler{
		viaCEPService:      services.NewViaCEPService(),
		weatherAPIService:  services.NewWeatherAPIService(),
		temperatureService: services.NewTemperatureService(),
		validationService:  services.NewValidationService(),
	}
}

func (h *WeatherHandler) GetWeatherByZipcode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	zipcode := vars["zipcode"]

	zipcode = strings.ReplaceAll(zipcode, "-", "")

	if !h.validationService.ValidateZipcode(zipcode) {
		h.sendErrorResponse(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	location, err := h.viaCEPService.GetLocationByZipcode(zipcode)
	if err != nil {
		if strings.Contains(err.Error(), "zipcode not found") {
			h.sendErrorResponse(w, http.StatusNotFound, "can not find zipcode")
			return
		}
		h.sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	cityQuery := location.Localidade + "," + location.UF + ",Brazil"

	weather, err := h.weatherAPIService.GetCurrentWeather(cityQuery)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "weather service unavailable")
		// log the error to console
		log.Println(err)
		return
	}

	response := h.temperatureService.ConvertTemperatures(weather.Current.TempC)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *WeatherHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ErrorResponse{Message: message})
}
