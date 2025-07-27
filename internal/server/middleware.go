package server

import (
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	visitors = make(map[string]time.Time)
	mu       sync.Mutex
	limit    = time.Minute // 1 request per minute
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		mu.Lock()
		last, seen := visitors[ip]
		if seen && time.Since(last) < limit {
			mu.Unlock()
			http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
			return
		}
		visitors[ip] = time.Now()
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap the ResponseWriter to capture status code if needed
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, duration)
	})
}

func RecoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

// Chain composes multiple middleware handlers into one.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middlewares {
		h = mw(h)
	}
	return h
}
