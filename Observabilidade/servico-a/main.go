package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type ErrorResponse struct {
	Message string `json:"message"`
}

var tracer trace.Tracer

func main() {
	// Inicializar OpenTelemetry
	cleanup := initTracer()
	defer cleanup()

	// Configurar tracer
	tracer = otel.Tracer("servico-a")

	// Configurar rotas
	r := mux.NewRouter()
	r.Handle("/cep", otelhttp.NewHandler(http.HandlerFunc(handleCEP), "handle-cep")).Methods("POST")
	r.HandleFunc("/health", handleHealth).Methods("GET")

	// Configurar servidor
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Printf("Serviço A iniciado na porta %s", port)
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

func handleCEP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "handle-cep")
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

	// Encaminhar para Serviço B
	statusCode, responseBody, err := forwardToServiceB(ctx, cepReq.CEP)
	if err != nil {
		span.RecordError(err)
		log.Printf("Erro ao encaminhar para Serviço B: %v", err)
		sendErrorResponse(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Retornar resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}

func validateCEP(cep string) bool {
	// Verificar se é string com exatamente 8 dígitos
	matched, _ := regexp.MatchString(`^\d{8}$`, cep)
	return matched
}

func forwardToServiceB(ctx context.Context, cep string) (int, []byte, error) {
	ctx, span := tracer.Start(ctx, "forward-to-service-b")
	defer span.End()

	serviceBURL := getEnv("SERVICE_B_URL", "http://servico-b:8081")
	url := fmt.Sprintf("%s/weather", serviceBURL)

	reqBody := CEPRequest{CEP: cep}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, nil, err
	}

	// Criar request com contexto para propagação de trace
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Usar cliente HTTP com instrumentação OTEL
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
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
