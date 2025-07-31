package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int("status_code", wrapper.statusCode),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Create a map of clients keyed by IP address
var (
	clients = make(map[string]*client)
	mu      sync.Mutex // Mutex to protect access to the clients map
)

// Run a background goroutine to remove old entries from the clients map.
func init() {
	go cleanupClients()
}

func getClient(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	// Check if the client exists.
	if c, found := clients[ip]; found {
		// Update the last seen time.
		c.lastSeen = time.Now()
		return c.limiter
	}

	// If the client does not exist, create a new one.
	// Allow 2 requests per second with a burst of 4.
	limiter := rate.NewLimiter(2, 4)
	clients[ip] = &client{limiter, time.Now()}

	return limiter
}

func cleanupClients() {
	// Run an endless loop every minute.
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		// Check for clients that haven't been seen in the last 3 minutes.
		for ip, c := range clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

// The new middleware function.
func PerClientRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the IP address for the current user.
		// `r.RemoteAddr` may include the port, so we use SplitHostPort.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If we can't get the IP, we can't rate limit.
			// You might want to log this error or block the request.
			// For this example, we'll allow it.
			ip = r.RemoteAddr
		}

		// Get the limiter for the specific IP address.
		limiter := getClient(ip)

		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
