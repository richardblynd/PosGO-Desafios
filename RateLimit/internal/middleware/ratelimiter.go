package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/ratelimiter"
	"github.com/richardblynd/PosGO-Desafios/tree/main/ratelimiter/internal/storage"
)

// RateLimiterMiddleware provides HTTP middleware for rate limiting
type RateLimiterMiddleware struct {
	rateLimiter *ratelimiter.RateLimiter
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
	Timestamp string `json:"timestamp"`
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(rateLimiter *ratelimiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiter: rateLimiter,
	}
}

// Handler returns an HTTP handler that applies rate limiting
func (rlm *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Extract client IP
		clientIP := rlm.getClientIP(r)

		// Extract API token from header
		token := rlm.getAPIToken(r)

		// Check rate limit
		result, err := rlm.rateLimiter.Check(ctx, clientIP, token)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			rlm.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", err.Error())
			return
		}

		// Add rate limit headers
		rlm.setRateLimitHeaders(w, result)

		if !result.Allowed {
			// Rate limit exceeded
			w.Header().Set("Retry-After", formatRetryAfter(result.RetryAfter))
			rlm.writeErrorResponse(w, http.StatusTooManyRequests,
				"you have reached the maximum number of requests or actions allowed within a certain time frame",
				"Rate limit exceeded")
			return
		}

		// Request allowed, proceed to next handler
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the real client IP from the request
func (rlm *RateLimiterMiddleware) getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first (for proxy/load balancer setups)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				return ip
			}
		}
	}

	// Try X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if parsedIP := net.ParseIP(xri); parsedIP != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, assume RemoteAddr is just IP
		if parsedIP := net.ParseIP(r.RemoteAddr); parsedIP != nil {
			return r.RemoteAddr
		}
		// If all else fails, return a default IP
		return "127.0.0.1"
	}

	if parsedIP := net.ParseIP(host); parsedIP != nil {
		return host
	}

	return "127.0.0.1"
}

// getAPIToken extracts the API token from the API_KEY header
func (rlm *RateLimiterMiddleware) getAPIToken(r *http.Request) string {
	// Check API_KEY header
	if token := r.Header.Get("API_KEY"); token != "" {
		return strings.TrimSpace(token)
	}

	// Also check Authorization header for Bearer tokens (alternative format)
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}
	}

	return ""
}

// setRateLimitHeaders sets standard rate limiting headers
func (rlm *RateLimiterMiddleware) setRateLimitHeaders(w http.ResponseWriter, result *storage.RateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", formatInt(result.Limit))
	w.Header().Set("X-RateLimit-Remaining", formatInt(result.Remaining))
	w.Header().Set("X-RateLimit-Reset", formatReset())
}

// writeErrorResponse writes a JSON error response
func (rlm *RateLimiterMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:     http.StatusText(statusCode),
		Message:   message,
		Code:      statusCode,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// Helper functions

func formatRetryAfter(duration time.Duration) string {
	seconds := int(duration.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return formatInt(seconds)
}

func formatReset() string {
	// Next reset time (next second)
	nextReset := time.Now().Add(time.Second).Unix()
	return formatInt64(nextReset)
}

func formatInt(i int) string {
	return formatInt64(int64(i))
}

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}
