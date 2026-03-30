package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/podland/backend/pkg/response"
)

// LogsHandler handles logs API requests
type LogsHandler struct {
	lokiURL    string
	httpClient *http.Client
	wsUpgrader websocket.Upgrader
}

// LokiResponse represents the response from Loki API
type LokiResponse struct {
	Status string        `json:"status"`
	Data   LokiData      `json:"data"`
	Stats  LokiStats     `json:"stats,omitempty"`
}

// LokiData represents log data from Loki
type LokiData struct {
	ResultType string       `json:"resultType"`
	Result     []LokiResult `json:"result"`
}

// LokiResult represents a single log result
type LokiResult struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// LokiStats represents query statistics
type LokiStats struct {
	Summary map[string]interface{} `json:"summary"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// LogsResponse represents the API response for logs
type LogsResponse struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler() *LogsHandler {
	lokiURL := os.Getenv("LOKI_URL")
	if lokiURL == "" {
		lokiURL = "http://loki.monitoring.svc:3100"
	}

	return &LogsHandler{
		lokiURL:    lokiURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for WebSocket
			},
		},
	}
}

// GetVMLogs returns logs for a specific VM
func (h *LogsHandler) GetVMLogs(w http.ResponseWriter, r *http.Request) {
	vmIDStr := chi.URLParam(r, "id")
	if vmIDStr == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	vmID, err := uuid.Parse(vmIDStr)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid VM ID format")
		return
	}

	// Parse query params
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "1000"
	}

	level := r.URL.Query().Get("level")
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")

	// Build LogQL query
	query := fmt.Sprintf(`{vm_id="%s"}`, vmID.String())
	if level != "" {
		query += fmt.Sprintf(` |= "%s"`, level)
	}

	// Query Loki
	entries, err := h.queryLoki(query, limit, startTime, endTime)
	if err != nil {
		log.Printf("Failed to query Loki: %v", err)
		pkgresponse.InternalError(w, "Failed to fetch logs")
		return
	}

	pkgresponse.Success(w, http.StatusOK, LogsResponse{
		Entries: entries,
		Total:   len(entries),
	})
}

// StreamVMLogs streams logs for a specific VM via WebSocket
func (h *LogsHandler) StreamVMLogs(w http.ResponseWriter, r *http.Request) {
	vmIDStr := chi.URLParam(r, "id")
	if vmIDStr == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	vmID, err := uuid.Parse(vmIDStr)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid VM ID format")
		return
	}

	// Upgrade to WebSocket
	conn, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Build LogQL query
	query := fmt.Sprintf(`{vm_id="%s"}`, vmID.String())

	// Send initial logs
	entries, err := h.queryLoki(query, "100", "", "")
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "Failed to fetch initial logs"})
		return
	}

	conn.WriteJSON(map[string]interface{}{
		"type":    "initial",
		"entries": entries,
	})

	// Poll for new logs every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastTimestamp := time.Now()
	if len(entries) > 0 {
		lastTimestamp = entries[0].Timestamp
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Query for new logs since last timestamp
			newEntries, err := h.queryLokiSince(query, "100", lastTimestamp)
			if err != nil {
				log.Printf("Failed to query new logs: %v", err)
				continue
			}

			if len(newEntries) > 0 {
				conn.WriteJSON(map[string]interface{}{
					"type":    "update",
					"entries": newEntries,
				})
				lastTimestamp = newEntries[0].Timestamp
			}
		}
	}
}

func (h *LogsHandler) queryLoki(query, limit, startTime, endTime string) ([]LogEntry, error) {
	// Build Loki query URL
	lokiURL := fmt.Sprintf("%s/loki/api/v1/query_range", h.lokiURL)
	req, err := http.NewRequest("GET", lokiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("limit", limit)

	// Set time range (default: last 24 hours)
	if startTime == "" {
		startTime = strconv.FormatInt(time.Now().Add(-24*time.Hour).UnixNano(), 10)
	}
	if endTime == "" {
		endTime = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	q.Add("start", startTime)
	q.Add("end", endTime)
	q.Add("direction", "BACKWARD")

	req.URL.RawQuery = q.Encode()

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lokiResp LokiResponse
	if err := json.Unmarshal(body, &lokiResp); err != nil {
		return nil, err
	}

	if lokiResp.Status != "success" {
		return nil, fmt.Errorf("loki query failed: %s", string(body))
	}

	// Parse log entries
	var entries []LogEntry
	for _, result := range lokiResp.Data.Result {
		for _, value := range result.Values {
			if len(value) < 2 {
				continue
			}

			timestamp, err := strconv.ParseInt(value[0], 10, 64)
			if err != nil {
				continue
			}

			entries = append(entries, LogEntry{
				Timestamp: time.Unix(0, timestamp),
				Line:      value[1],
				Labels:    result.Stream,
			})
		}
	}

	return entries, nil
}

func (h *LogsHandler) queryLokiSince(query, limit string, since time.Time) ([]LogEntry, error) {
	endTime := strconv.FormatInt(time.Now().UnixNano(), 10)
	startTime := strconv.FormatInt(since.UnixNano(), 10)
	return h.queryLoki(query, limit, startTime, endTime)
}

// GetLogLevels returns available log levels for a VM
func (h *LogsHandler) GetLogLevels(w http.ResponseWriter, r *http.Request) {
	vmIDStr := chi.URLParam(r, "id")
	if vmIDStr == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	vmID, err := uuid.Parse(vmIDStr)
	if err != nil {
		pkgresponse.BadRequest(w, "Invalid VM ID format")
		return
	}

	// Query Loki for distinct log levels
	query := fmt.Sprintf(`{vm_id="%s"} | pattern "<_> <_> <_> <_> <_> <msg>"`, vmID.String())
	entries, err := h.queryLoki(query, "1000", "", "")
	if err != nil {
		log.Printf("Failed to query log levels: %v", err)
		pkgresponse.InternalError(w, "Failed to fetch log levels")
		return
	}

	// Extract unique log levels
	levelSet := make(map[string]bool)
	for _, entry := range entries {
		// Try to extract level from log line
		line := entry.Line
		if len(line) > 0 {
			// Simple heuristic: look for common log level patterns
			if len(line) > 5 {
				potentialLevel := line[:5]
				if isLogLevel(potentialLevel) {
					levelSet[potentialLevel] = true
				}
			}
		}
	}

	levels := make([]string, 0, len(levelSet))
	for level := range levelSet {
		levels = append(levels, level)
	}

	pkgresponse.Success(w, http.StatusOK, map[string]interface{}{
		"levels": levels,
	})
}

func isLogLevel(s string) bool {
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL", "TRACE"}
	for _, level := range validLevels {
		if s == level {
			return true
		}
	}
	return false
}
