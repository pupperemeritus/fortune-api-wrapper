package handlers

import (
	"encoding/json"
	"errors"
	"fortune-api/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockFortuneService is a mock implementation of the FortuneService.
// It allows us to test the handler logic in isolation.
type MockFortuneService struct {
	mock.Mock
}

func (m *MockFortuneService) GetFortune(opts service.FortuneOptions) (*service.FortuneResponse, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.FortuneResponse), args.Error(1)
}

func (m *MockFortuneService) ListFiles() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFortuneService) SearchFortunes(pattern string, opts service.FortuneOptions) (*service.SearchResponse, error) {
	args := m.Called(pattern, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SearchResponse), args.Error(1)
}

// setupTestHandler initializes a handler with a mock service for testing.
func setupTestHandler() (*Handler, *MockFortuneService) {
	mockService := new(MockFortuneService)
	logger := zap.NewNop() // Use a no-op logger for tests to keep output clean.
	handler := NewHandler(mockService, logger)
	return handler, mockService
}

func TestHealthCheck(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Health check should return status OK")

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "fortune-api", response["service"])
}

func TestGetFortune_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	expectedFortune := &service.FortuneResponse{Fortune: "Your future is bright."}
	// We expect GetFortune to be called once with any FortuneOptions and return our expected fortune.
	mockService.On("GetFortune", mock.AnythingOfType("service.FortuneOptions")).Return(expectedFortune, nil)

	req := httptest.NewRequest("GET", "/fortune?short=true", nil)
	rr := httptest.NewRecorder()

	handler.GetFortune(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t) // Verify that the mock was called as expected.

	var actualFortune service.FortuneResponse
	err := json.Unmarshal(rr.Body.Bytes(), &actualFortune)
	assert.NoError(t, err)
	assert.Equal(t, expectedFortune.Fortune, actualFortune.Fortune)
}

func TestGetFortune_ServiceError(t *testing.T) {
	handler, mockService := setupTestHandler()

	// Configure the mock to return an error.
	mockService.On("GetFortune", mock.AnythingOfType("service.FortuneOptions")).Return(nil, errors.New("something went wrong"))

	req := httptest.NewRequest("GET", "/fortune", nil)
	rr := httptest.NewRecorder()

	handler.GetFortune(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)

	var errResponse ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to get fortune", errResponse.Error)
}

func TestListFiles_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	expectedFiles := []string{"star-trek", "wisdom"}
	mockService.On("ListFiles").Return(expectedFiles, nil)

	req := httptest.NewRequest("GET", "/fortune/files", nil)
	rr := httptest.NewRecorder()

	handler.ListFiles(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"]) // JSON numbers are float64 by default
	assert.ElementsMatch(t, expectedFiles, response["files"])
}

func TestSearchFortunes_Success(t *testing.T) {
	handler, mockService := setupTestHandler()

	expectedSearch := &service.SearchResponse{
		Matches: []service.FortuneResponse{{Fortune: "A search has found you."}},
		Count:   1,
	}
	mockService.On("SearchFortunes", "test", mock.AnythingOfType("service.FortuneOptions")).Return(expectedSearch, nil)

	req := httptest.NewRequest("GET", "/fortune/search?pattern=test", nil)
	rr := httptest.NewRecorder()

	handler.SearchFortunes(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var actualSearch service.SearchResponse
	err := json.Unmarshal(rr.Body.Bytes(), &actualSearch)
	assert.NoError(t, err)
	assert.Equal(t, *expectedSearch, actualSearch)
}

func TestSearchFortunes_MissingPattern(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/fortune/search", nil) // No pattern query param
	rr := httptest.NewRecorder()

	handler.SearchFortunes(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResponse ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Missing required parameter", errResponse.Error)
}
