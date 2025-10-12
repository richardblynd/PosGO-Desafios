package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestWeatherHandler_GetWeatherByZipcode_InvalidZipcode(t *testing.T) {
	handler := NewWeatherHandler()

	tests := []struct {
		name           string
		zipcode        string
		expectedStatus int
	}{
		{
			name:           "Invalid zipcode - too short",
			zipcode:        "0100100",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Invalid zipcode - contains letters",
			zipcode:        "0100100A",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Invalid zipcode - too long",
			zipcode:        "010010001",
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/weather/"+tt.zipcode, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/weather/{zipcode}", handler.GetWeatherByZipcode)

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
