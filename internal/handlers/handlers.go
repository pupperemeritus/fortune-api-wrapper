package handlers

import (
	"encoding/json"
	"fortune-api/internal/service"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type Handler struct {
	// Depend on the interface, not the concrete type.
	fortuneService service.FortuneServiceInterface
	logger         *zap.Logger
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// NewHandler now accepts the interface.
func NewHandler(fortuneService service.FortuneServiceInterface, logger *zap.Logger) *Handler {
	return &Handler{
		fortuneService: fortuneService,
		logger:         logger,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "fortune-api",
	})
}

func (h *Handler) GetFortune(w http.ResponseWriter, r *http.Request) {
	opts := h.parseFortuneOptions(r)

	fortune, err := h.fortuneService.GetFortune(opts)
	if err != nil {
		h.logger.Error("Failed to get fortune", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get fortune", err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, fortune)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := h.fortuneService.ListFiles()
	if err != nil {
		h.logger.Error("Failed to list fortune files", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list files", err.Error())
		return
	}

	response := map[string]interface{}{
		"files": files,
		"count": len(files),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) SearchFortunes(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	if pattern == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing required parameter", "pattern parameter is required")
		return
	}

	opts := h.parseFortuneOptions(r)

	results, err := h.fortuneService.SearchFortunes(pattern, opts)
	if err != nil {
		h.logger.Error("Failed to search fortunes", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Search failed", err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, results)
}

func (h *Handler) parseFortuneOptions(r *http.Request) service.FortuneOptions {
	query := r.URL.Query()

	opts := service.FortuneOptions{
		All:        query.Get("all") == "true",
		ShowCookie: query.Get("show_cookie") == "true",
		Equal:      query.Get("equal") == "true",
		Long:       query.Get("long") == "true",
		Short:      query.Get("short") == "true",
		IgnoreCase: query.Get("ignore_case") == "true",
		Wait:       query.Get("wait") == "true",
		Pattern:    query.Get("pattern"),
	}

	if lengthStr := query.Get("length"); lengthStr != "" {
		if length, err := strconv.Atoi(lengthStr); err == nil {
			opts.Length = length
		}
	}

	if filesStr := query.Get("files"); filesStr != "" {
		opts.Files = strings.Split(filesStr, ",")
	}

	if percentagesStr := query.Get("percentages"); percentagesStr != "" {
		opts.Percentages = strings.Split(percentagesStr, ",")
	}

	return opts
}

func (h *Handler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	response := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSONResponse(w, statusCode, response)
}
