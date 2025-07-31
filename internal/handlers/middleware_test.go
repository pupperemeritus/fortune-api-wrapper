package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// dummyHandler is a simple handler that writes OK.
var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
})

func TestCORSMiddleware(t *testing.T) {
	// Create a test server with the CORS middleware.
	ts := httptest.NewServer(CORSMiddleware(dummyHandler))
	defer ts.Close()

	// Test a standard GET request
	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))

	// Test a pre-flight OPTIONS request
	req, _ = http.NewRequest("OPTIONS", ts.URL, nil)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestLoggingMiddleware(t *testing.T) {
	// Use an observed core to capture log output.
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	// Create a test server with the logging middleware.
	ts := httptest.NewServer(LoggingMiddleware(logger)(dummyHandler))
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/testpath?q=1", nil)
	req.Header.Set("User-Agent", "test-agent")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Check that one log entry was created.
	assert.Equal(t, 1, recorded.Len())
	log := recorded.All()[0]

	assert.Equal(t, "HTTP Request", log.Message)
	fields := log.ContextMap()
	assert.Equal(t, "GET", fields["method"])
	assert.Equal(t, "/testpath", fields["path"])
	assert.Equal(t, "q=1", fields["query"])
	assert.Equal(t, "test-agent", fields["user_agent"])
	assert.Equal(t, int64(http.StatusOK), fields["status_code"]) // Note: zapcore encodes ints as int64
	assert.NotZero(t, fields["duration"])
}

func TestPerClientRateLimit(t *testing.T) {
	// Reset the clients map for a clean test run.
	clients = make(map[string]*client)

	// Create a test server with the rate limit middleware.
	ts := httptest.NewServer(PerClientRateLimit(dummyHandler))
	defer ts.Close()

	// The default limiter is 2 requests per second with a burst of 4.
	// We should be able to make 4 requests successfully.
	for i := 0; i < 4; i++ {
		resp, err := http.Get(ts.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Request %d should be allowed", i+1)
		resp.Body.Close()
	}

	// The 5th request should be rate limited.
	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "5th request should be rate limited")
	resp.Body.Close()

	// Wait for the token bucket to refill slightly.
	time.Sleep(500 * time.Millisecond)

	// The next request should now be allowed.
	resp, err = http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Request after delay should be allowed")
	resp.Body.Close()
}
