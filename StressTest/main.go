package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Config holds the CLI parameters
type Config struct {
	URL         string
	Requests    int
	Concurrency int
}

// Result holds the response information
type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

// Report holds the test statistics
type Report struct {
	TotalTime       time.Duration
	TotalRequests   int
	StatusCounts    map[int]int
	SuccessCount    int
	FailureCount    int
	AverageDuration time.Duration
}

func main() {
	config := parseFlags()

	fmt.Printf("Starting stress test...\n")
	fmt.Printf("URL: %s\n", config.URL)
	fmt.Printf("Total requests: %d\n", config.Requests)
	fmt.Printf("Concurrency: %d\n", config.Concurrency)
	fmt.Println("---")

	report := runStressTest(config)
	printReport(report)
}

func parseFlags() Config {
	return parseFlagsFromArgs(os.Args[1:])
}

func parseFlagsFromArgs(args []string) Config {
	var config Config

	fs := flag.NewFlagSet("stresstest", flag.ExitOnError)
	fs.StringVar(&config.URL, "url", "", "URL do serviço a ser testado")
	fs.IntVar(&config.Requests, "requests", 0, "Número total de requests")
	fs.IntVar(&config.Concurrency, "concurrency", 1, "Número de chamadas simultâneas")

	fs.Parse(args)

	if config.URL == "" {
		log.Fatal("Parâmetro --url é obrigatório")
	}
	if config.Requests <= 0 {
		log.Fatal("Parâmetro --requests deve ser maior que 0")
	}
	if config.Concurrency <= 0 {
		log.Fatal("Parâmetro --concurrency deve ser maior que 0")
	}
	if config.Concurrency > config.Requests {
		config.Concurrency = config.Requests
	}

	return config
}

func runStressTest(config Config) Report {
	startTime := time.Now()

	// Channel para distribuir trabalho
	jobs := make(chan int, config.Requests)
	// Channel para coletar resultados
	results := make(chan Result, config.Requests)

	// Criar workers
	var wg sync.WaitGroup
	for w := 0; w < config.Concurrency; w++ {
		wg.Add(1)
		go worker(config.URL, jobs, results, &wg)
	}

	// Enviar jobs
	go func() {
		for i := 0; i < config.Requests; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Aguardar workers terminarem
	wg.Wait()
	close(results)

	// Processar resultados
	report := Report{
		TotalTime:     time.Since(startTime),
		TotalRequests: config.Requests,
		StatusCounts:  make(map[int]int),
	}

	var totalDuration time.Duration

	for result := range results {
		if result.Error == nil {
			report.StatusCounts[result.StatusCode]++
			if result.StatusCode == 200 {
				report.SuccessCount++
			}
			totalDuration += result.Duration
		} else {
			report.FailureCount++
		}
	}

	if report.TotalRequests > 0 {
		report.AverageDuration = totalDuration / time.Duration(report.TotalRequests)
	}

	return report
}

func worker(url string, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for range jobs {
		start := time.Now()

		resp, err := client.Get(url)
		duration := time.Since(start)

		result := Result{
			Duration: duration,
			Error:    err,
		}

		if err == nil {
			result.StatusCode = resp.StatusCode
			resp.Body.Close()
		}

		results <- result
	}
}

func printReport(report Report) {
	fmt.Println("\n=== RELATÓRIO DO TESTE DE CARGA ===")
	fmt.Printf("Tempo total de execução: %v\n", report.TotalTime)
	fmt.Printf("Total de requests realizados: %d\n", report.TotalRequests)
	fmt.Printf("Requests com status 200 (sucesso): %d\n", report.SuccessCount)
	fmt.Printf("Tempo médio por request: %v\n", report.AverageDuration)
	fmt.Printf("Requests por segundo: %.2f\n", float64(report.TotalRequests)/report.TotalTime.Seconds())

	fmt.Println("\nDistribuição de códigos de status HTTP:")
	for statusCode, count := range report.StatusCounts {
		fmt.Printf("  %d: %d requests\n", statusCode, count)
	}

	fmt.Println("\n=== FIM DO RELATÓRIO ===")
}
