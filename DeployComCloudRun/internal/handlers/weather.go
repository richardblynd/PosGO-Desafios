package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	// Necessário para o timestamp do log
	"weather-api/internal/models"
	"weather-api/internal/services"

	"github.com/gorilla/mux"
)

// --- Definição e Construtor do Handler ---

type WeatherHandler struct {
	viaCEPService      *services.ViaCEPService
	weatherAPIService  *services.WeatherAPIService
	temperatureService *services.TemperatureService
	validationService  *services.ValidationService
}

func NewWeatherHandler() *WeatherHandler {
	// Log de Inicialização (aparece no Cloud Logging sem problemas)
	fmt.Println("Starting NewWeatherHandler")
	return &WeatherHandler{
		viaCEPService:      services.NewViaCEPService(),
		weatherAPIService:  services.NewWeatherAPIService(),
		temperatureService: services.NewTemperatureService(),
		validationService:  services.NewValidationService(),
	}
}

// --- Métodos do Handler ---

func (h *WeatherHandler) GetWeatherByZipcode(w http.ResponseWriter, r *http.Request) {

	log.Printf("GetWeatherByZipcode")

	// return a simple text message
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	// _, err := w.Write([]byte("Weather API is up and running!"))
	// if err != nil {
	// 	h.sendErrorResponse(w, http.StatusInternalServerError, "internal server error: "+err.Error())
	// }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//err := json.NewEncoder(w).Encode(map[string]string{"message": "xablau"})
	//if err != nil {
	//	h.sendErrorResponse(w, http.StatusInternalServerError, "internal server errorxxx"+err.Error())
	//	}

	vars := mux.Vars(r)
	zipcode := vars["zipcode"]

	zipcode = strings.ReplaceAll(zipcode, "-", "")

	if !h.validationService.ValidateZipcode(zipcode) {
		// LOG: Erro de validação
		h.sendErrorResponse(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	location, err := h.viaCEPService.GetLocationByZipcode(zipcode)
	if err != nil {
		if strings.Contains(err.Error(), "zipcode not found") {
			// LOG: Erro de CEP não encontrado
			h.sendErrorResponse(w, http.StatusNotFound, "can not find zipcode")
			return
		}
		// LOG: Erro interno grave (ViaCEP)
		h.sendErrorResponse(w, http.StatusInternalServerError, "internal server errorx"+err.Error())
		return
	}

	cityQuery := location.Localidade + "," + location.UF + ",Brazil"

	weather, err := h.weatherAPIService.GetCurrentWeather(cityQuery)
	if err != nil {
		// LOG: Erro de serviço externo (WeatherAPI)
		h.sendErrorResponse(w, http.StatusInternalServerError, "weather service unavailable")
		return
	}

	response := h.temperatureService.ConvertTemperatures(weather.Current.TempC)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "internal server errorxxx"+err.Error())
	}
}

func (h *WeatherHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	//json.NewEncoder(w).Encode(models.ErrorResponse{Message: "xablau"})
	json.NewEncoder(w).Encode(models.ErrorResponse{Message: message})
}
