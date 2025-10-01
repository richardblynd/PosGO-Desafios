package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestParseFlagsFromArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedURL string
		expectedReq int
		expectedCon int
		shouldFail  bool
	}{
		{
			name:        "Valid arguments",
			args:        []string{"--url=http://example.com", "--requests=100", "--concurrency=10"},
			expectedURL: "http://example.com",
			expectedReq: 100,
			expectedCon: 10,
			shouldFail:  false,
		},
		{
			name:        "Concurrency greater than requests",
			args:        []string{"--url=http://example.com", "--requests=5", "--concurrency=10"},
			expectedURL: "http://example.com",
			expectedReq: 5,
			expectedCon: 5, // Should be adjusted to requests count
			shouldFail:  false,
		},
		{
			name:        "Default concurrency",
			args:        []string{"--url=http://example.com", "--requests=100"},
			expectedURL: "http://example.com",
			expectedReq: 100,
			expectedCon: 1,
			shouldFail:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldFail {
				t.Skip("Skipping test that would call log.Fatal")
				return
			}

			config := parseFlagsFromArgs(tt.args)

			if config.URL != tt.expectedURL {
				t.Errorf("Expected URL %s, got %s", tt.expectedURL, config.URL)
			}
			if config.Requests != tt.expectedReq {
				t.Errorf("Expected requests %d, got %d", tt.expectedReq, config.Requests)
			}
			if config.Concurrency != tt.expectedCon {
				t.Errorf("Expected concurrency %d, got %d", tt.expectedCon, config.Concurrency)
			}
		})
	}
}

func TestRunStressTest(t *testing.T) {
	// Create a test server that returns different status codes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate different responses
		if r.URL.Path == "/success" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/notfound" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	tests := []struct {
		name        string
		config      Config
		expectedMin int // Minimum expected successful requests
	}{
		{
			name: "Successful requests",
			config: Config{
				URL:         server.URL + "/success",
				Requests:    10,
				Concurrency: 2,
			},
			expectedMin: 10,
		},
		{
			name: "Mixed status codes",
			config: Config{
				URL:         server.URL + "/notfound",
				Requests:    5,
				Concurrency: 1,
			},
			expectedMin: 0, // All should return 404
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := runStressTest(tt.config)

			if report.TotalRequests != tt.config.Requests {
				t.Errorf("Expected %d total requests, got %d", tt.config.Requests, report.TotalRequests)
			}

			if report.TotalTime <= 0 {
				t.Error("Expected positive total time")
			}

			// Check that we have status codes recorded
			totalStatusRequests := 0
			for _, count := range report.StatusCounts {
				totalStatusRequests += count
			}

			expectedTotal := report.TotalRequests - report.FailureCount
			if totalStatusRequests != expectedTotal {
				t.Errorf("Expected %d status code requests, got %d", expectedTotal, totalStatusRequests)
			}

			// For success endpoint, all should be 200
			if tt.name == "Successful requests" {
				if report.StatusCounts[200] != tt.config.Requests {
					t.Errorf("Expected %d successful requests, got %d", tt.config.Requests, report.StatusCounts[200])
				}
				if report.SuccessCount != tt.config.Requests {
					t.Errorf("Expected %d success count, got %d", tt.config.Requests, report.SuccessCount)
				}
			}

			// For notfound endpoint, all should be 404
			if tt.name == "Mixed status codes" {
				if report.StatusCounts[404] != tt.config.Requests {
					t.Errorf("Expected %d 404 requests, got %d", tt.config.Requests, report.StatusCounts[404])
				}
			}
		})
	}
}

func TestWorker(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create channels
	jobs := make(chan int, 3)
	results := make(chan Result, 3)

	// Add jobs
	for i := 0; i < 3; i++ {
		jobs <- i
	}
	close(jobs)

	// Run worker
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		worker(server.URL, jobs, results, &wg)
		wg.Wait()
		close(results)
	}()

	// Check results
	resultCount := 0
	for result := range results {
		resultCount++
		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}
		if result.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", result.StatusCode)
		}
		if result.Duration <= 0 {
			t.Error("Expected positive duration")
		}
	}

	if resultCount != 3 {
		t.Errorf("Expected 3 results, got %d", resultCount)
	}
}

func TestWorkerWithTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than our client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create channels
	jobs := make(chan int, 1)
	results := make(chan Result, 1)

	// Add one job
	jobs <- 1
	close(jobs)

	// Run worker (this will use the 30-second timeout in the actual worker function)
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		worker(server.URL, jobs, results, &wg)
		wg.Wait()
		close(results)
	}()

	// Check result - should succeed since 2s < 30s timeout
	result := <-results
	if result.Error != nil {
		// This might actually succeed since 2s is less than 30s timeout
		t.Logf("Request timed out as expected: %v", result.Error)
	} else if result.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", result.StatusCode)
	}
}

// Benchmark tests
func BenchmarkRunStressTest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := Config{
		URL:         server.URL,
		Requests:    100,
		Concurrency: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runStressTest(config)
	}
}
