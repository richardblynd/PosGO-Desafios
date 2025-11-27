package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"weather-api/internal/models"
)

type ViaCEPService struct {
	client  *http.Client
	baseURL string
}

func NewViaCEPService() *ViaCEPService {
	return &ViaCEPService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://viacep.com.br/ws",
	}
}

func (v *ViaCEPService) GetLocationByZipcode(zipcode string) (*models.ViaCEPResponse, error) {
	url := fmt.Sprintf("%s/%s/json/", v.baseURL, zipcode)

	resp, err := v.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viaCEP API returned status %d", resp.StatusCode)
	}

	var cepResp models.ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&cepResp); err != nil {
		return nil, err
	}

	if cepResp.Erro == "true" {
		return nil, fmt.Errorf("zipcode not found")
	}

	return &cepResp, nil
}
