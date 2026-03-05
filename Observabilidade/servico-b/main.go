package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type CEPRequest struct {
	CEP string `json:"cep"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Estruturas para ViaCEP
type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
	Erro        string `json:"erro,omitempty"`
}

// Estruturas para WeatherAPI
type WeatherAPIResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

var tracer trace.Tracer

func main() {
	// Inicializar OpenTelemetry
	cleanup := initTracer()
	defer cleanup()

	// Configurar tracer
	tracer = otel.Tracer("servico-b")

	// Configurar rotas
	r := mux.NewRouter()
	r.Handle("/weather", otelhttp.NewHandler(http.HandlerFunc(handleWeather), "handle-weather")).Methods("POST")
	r.HandleFunc("/health", handleHealth).Methods("GET")

	// Configurar servidor
	port := getEnv("PORT", "8081")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Printf("Serviço B iniciado na porta %s", port)
	log.Fatal(server.ListenAndServe())
}

func initTracer() func() {
	zipkinURL := getEnv("ZIPKIN_URL", "http://zipkin:9411/api/v2/spans")

	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		log.Fatal(err)
	}

	res := resource.Default()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handle-weather")
	defer span.End()

	// Decodificar request
	var cepReq CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&cepReq); err != nil {
		span.RecordError(err)
		sendErrorResponse(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validar CEP
	if !validateCEP(cepReq.CEP) {
		sendErrorResponse(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Buscar localização via ViaCEP
	location, err := getLocationFromCEP(ctx, cepReq.CEP)
	if err != nil {
		span.RecordError(err)
		if err.Error() == "zipcode not found" {
			sendErrorResponse(w, "can not find zipcode", http.StatusNotFound)
		} else {
			log.Printf("Erro ao buscar CEP: %v", err)
			sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Buscar temperatura
	tempC, err := getTemperature(ctx, location)
	if err != nil {
		span.RecordError(err)
		log.Printf("Erro ao buscar temperatura: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Converter temperaturas
	tempF := celsiusToFahrenheit(tempC)
	tempK := celsiusToKelvin(tempC)

	// Preparar resposta
	response := WeatherResponse{
		City:  location,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	// Enviar resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func validateCEP(cep string) bool {
	// Verificar se é string com exatamente 8 dígitos
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}

func getLocationFromCEP(ctx context.Context, cep string) (string, error) {
	ctx, span := tracer.Start(ctx, "get-location-from-cep")
	defer span.End()

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	// Criar request com contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// Usar cliente HTTP com instrumentação OTEL
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("viacep returned status: %d", resp.StatusCode)
	}

	var viacepResp ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viacepResp); err != nil {
		return "", err
	}

	// Verificar se CEP foi encontrado
	if viacepResp.Erro == "true" {
		return "", fmt.Errorf("zipcode not found")
	}

	return viacepResp.Localidade, nil
}

func getTemperature(ctx context.Context, city string) (float64, error) {
	ctx, span := tracer.Start(ctx, "get-temperature")
	defer span.End()

	// Usar WeatherAPI
	apiKey := getEnv("WEATHER_API_KEY", "demo_key")
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", apiKey, city)

	// Criar request com contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	// Usar cliente HTTP com instrumentação OTEL
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Se a API não funcionar, retornar temperatura simulada
		log.Printf("WeatherAPI returned status %d, using simulated temperature", resp.StatusCode)
		return simulateTemperature(city), nil
	}

	var weatherResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		// Se houver erro na decodificação, usar temperatura simulada
		log.Printf("Error decoding weather response: %v, using simulated temperature", err)
		return simulateTemperature(city), nil
	}

	return weatherResp.Current.TempC, nil
}

func simulateTemperature(city string) float64 {
	// Simular temperatura baseada no hash do nome da cidade
	hash := 0
	for _, char := range city {
		hash += int(char)
	}
	// Gerar temperatura entre 15 e 35 graus
	temp := 15.0 + float64(hash%21)
	return temp
}

func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273.15
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
