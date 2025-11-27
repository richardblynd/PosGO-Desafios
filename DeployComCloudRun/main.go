package main

import (
	"log"
	"net/http"
	"os"

	"weather-api/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env file")
	}

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	router := mux.NewRouter()

	weatherHandler := handlers.NewWeatherHandler()
	router.HandleFunc("/weather/{zipcode}", weatherHandler.GetWeatherByZipcode).Methods("GET")

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
