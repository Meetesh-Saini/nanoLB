package lb

import (
	"fmt"
	"nanoLB/internal/about"
	configProvider "nanoLB/internal/config"
	"net/http"
)

type ctxKey int
type ERROR int

const (
	Attempts ctxKey = iota
	Retry
)

const (
	ServiceUnavailable ERROR = iota
)

// Get attempts for request from context
func GetAttempts(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

// Get retries for request from context
func GetRetries(r *http.Request) int {
	if retries, ok := r.Context().Value(Retry).(int); ok {
		return retries
	}
	return 0
}

func (e ERROR) String() string {
	return [...]string{fmt.Sprintf("<h1>%d Service is unavailable</h1><hr><p style='width:100%%;text-align:center;'>nanoLB (%s)</p>", http.StatusServiceUnavailable, about.Version)}[e]
}

// Send custom html errors
func HttpHtmlError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, error)
}

// Provides algorithm based on config enum
func GetAlgo(a configProvider.ALGORITHM_TYPE) (algo Algorithm) {
	algo = GetRoundRobin()
	switch a {
	case configProvider.RoundRobin:
		algo = GetRoundRobin()
	}
	return
}
