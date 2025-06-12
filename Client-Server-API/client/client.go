package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type GetQuoteResult struct {
	Cotacao float64 `json:"cotacao"`
}

const quoteApiUrl = "http://localhost:8080/"
const externaApiTimeout = 300 * time.Millisecond

var (
	ErrTimeout = errors.New("timeout atingido")
)

func getQuote(moedaOrigem, moedaDestino string) (*GetQuoteResult, error) {
	quoteApiUrlWithParams := quoteApiUrl + "cotacao/" + moedaOrigem + "/" + moedaDestino

	ctx, cancel := context.WithTimeout(context.Background(), externaApiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", quoteApiUrlWithParams, nil)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {

		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Timeout atingido após " + externaApiTimeout.String() + "ms")
			return nil, ErrTimeout
		}

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao chamar serviço externo: %s", resp.Status)
	}

	defer resp.Body.Close()

	var quoteResult GetQuoteResult

	if err := json.NewDecoder(resp.Body).Decode(&quoteResult); err != nil {
		return nil, err
	}

	return &quoteResult, nil
}

func saveQuoteFile(quote *GetQuoteResult, fileName string) {
	fileContent := fmt.Sprintf("Dólar: %.2f", quote.Cotacao)

	err := os.WriteFile(fileName, []byte(fileContent), 0644)
	if err != nil {
		fmt.Println("Erro ao salvar o arquivo:", err)
		return
	}
}

func main() {
	quote, err := getQuote("USD", "BRL")

	if err != nil {
		fmt.Println("Erro ao obter cotação:", err)
		return
	}

	saveQuoteFile(quote, "cotacao.txt")
}
