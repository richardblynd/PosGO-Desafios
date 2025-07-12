package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const quoteApiUrl = "https://brasilapi.com.br/api/cep/v1/"
const externaApiTimeout = 1000 * time.Millisecond

func callApi(apiUrl string) string {
	ctx, cancel := context.WithTimeout(context.Background(), externaApiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)

	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Erro ao chamar serviço externo: %s\n", resp.Status)
		return ""
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	bodyString := string(bodyBytes)
	return bodyString
}

func getCepBrasilApi(cep string, c chan<- string) {
	brasilCepApiUrlWithParams := quoteApiUrl + cep
	result := callApi(brasilCepApiUrlWithParams)
	c <- result
}

func getCepViaCepApi(cep string, c chan<- string) {
	viaCepApiUrl := "https://viacep.com.br/ws/" + cep + "/json/"
	result := callApi(viaCepApiUrl)
	c <- result
}

func main() {
	channelCepBrasil := make(chan string)
	channelCepViaCep := make(chan string)

	cep := "01001-000"

	go getCepBrasilApi(cep, channelCepBrasil)
	go getCepViaCepApi(cep, channelCepViaCep)

	select {
	case result := <-channelCepBrasil:
		fmt.Println("Resultado da API Brasil:", result)
	case result := <-channelCepViaCep:
		fmt.Println("Resultado da API ViaCep:", result)
	case <-time.After(externaApiTimeout):
		fmt.Println("Tempo limite excedido ao chamar as APIs externas")
		return
	}
}
