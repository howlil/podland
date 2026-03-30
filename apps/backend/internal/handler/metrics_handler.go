package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/podland/backend/pkg/response"
)

// MetricsHandler handles metrics API requests
type MetricsHandler struct {
	prometheusURL string
	httpClient    *http.Client
}

// PrometheusResponse represents the response from Prometheus API
type PrometheusResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// MetricsData represents time series data from Prometheus
type MetricsData struct {
	ResultType string          `json:"resultType"`
	Result     []MetricsResult `json:"result"`
}

// MetricsResult represents a single metric result
type MetricsResult struct {
	Metric map[string]interface{} `json:"metric"`
	Value  []interface{}          `json:"value"`
	Values []interface{}          `json:"values"`
}

// VMMetrics represents aggregated VM metrics
type VMMetrics struct {
	CPU    *MetricSeries `json:"cpu,omitempty"`
	Memory *MetricSeries `json:"memory,omitempty"`
	NetRx  *MetricSeries `json:"network_rx,omitempty"`
	NetTx  *MetricSeries `json:"network_tx,omitempty"`
}

// MetricSeries represents a time series of metric values
type MetricSeries struct {
	Current float64 `json:"current"`
	Average float64 `json:"average"`
	Max     float64 `json:"max"`
	Min     float64 `json:"min"`
	Points  []Point `json:"points,omitempty"`
}

// Point represents a single data point
type Point struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://prometheus.monitoring.svc:9090"
	}

	return &MetricsHandler{
		prometheusURL: prometheusURL,
		httpClient:    &http.Client{},
	}
}

// GetVMMetrics returns metrics for a specific VM
func (h *MetricsHandler) GetVMMetrics(w http.ResponseWriter, r *http.Request) {
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
	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "24h"
	}

	step := r.URL.Query().Get("step")
	if step == "" {
		step = "5m"
	}

	// Query Prometheus for various metrics
	cpuData, err := h.queryCPUMetrics(vmID.String(), rangeStr, step)
	if err != nil {
		log.Printf("Failed to query CPU metrics: %v", err)
	}

	memoryData, err := h.queryMemoryMetrics(vmID.String(), rangeStr, step)
	if err != nil {
		log.Printf("Failed to query memory metrics: %v", err)
	}

	netRxData, err := h.queryNetworkRxMetrics(vmID.String(), rangeStr, step)
	if err != nil {
		log.Printf("Failed to query network RX metrics: %v", err)
	}

	netTxData, err := h.queryNetworkTxMetrics(vmID.String(), rangeStr, step)
	if err != nil {
		log.Printf("Failed to query network TX metrics: %v", err)
	}

	metrics := VMMetrics{
		CPU:    cpuData,
		Memory: memoryData,
		NetRx:  netRxData,
		NetTx:  netTxData,
	}

	pkgresponse.Success(w, http.StatusOK, metrics)
}

// RedirectToGrafana redirects to Grafana dashboard for detailed metrics
func (h *MetricsHandler) RedirectToGrafana(w http.ResponseWriter, r *http.Request) {
	vmIDStr := chi.URLParam(r, "id")
	if vmIDStr == "" {
		pkgresponse.BadRequest(w, "VM ID is required")
		return
	}

	grafanaURL := os.Getenv("GRAFANA_URL")
	if grafanaURL == "" {
		grafanaURL = "http://grafana.monitoring.svc:3000"
	}

	dashboardURL := fmt.Sprintf("%s/d/vm-metrics?var-vm_id=%s", grafanaURL, url.QueryEscape(vmIDStr))
	http.Redirect(w, r, dashboardURL, http.StatusSeeOther)
}

func (h *MetricsHandler) queryCPUMetrics(vmID, rangeStr, step string) (*MetricSeries, error) {
	query := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{vm_id="%s", image!=""}[5m])`, vmID)
	return h.queryMetric(query, rangeStr, step, true)
}

func (h *MetricsHandler) queryMemoryMetrics(vmID, rangeStr, step string) (*MetricSeries, error) {
	query := fmt.Sprintf(`container_memory_usage_bytes{vm_id="%s", image!=""}`, vmID)
	return h.queryMetric(query, rangeStr, step, false)
}

func (h *MetricsHandler) queryNetworkRxMetrics(vmID, rangeStr, step string) (*MetricSeries, error) {
	query := fmt.Sprintf(`rate(container_network_receive_bytes_total{vm_id="%s"}[5m])`, vmID)
	return h.queryMetric(query, rangeStr, step, true)
}

func (h *MetricsHandler) queryNetworkTxMetrics(vmID, rangeStr, step string) (*MetricSeries, error) {
	query := fmt.Sprintf(`rate(container_network_transmit_bytes_total{vm_id="%s"}[5m])`, vmID)
	return h.queryMetric(query, rangeStr, step, true)
}

func (h *MetricsHandler) queryMetric(query, rangeStr, step string, isRate bool) (*MetricSeries, error) {
	// Build Prometheus range query URL
	promURL := fmt.Sprintf("%s/api/v1/query_range", h.prometheusURL)
	req, err := http.NewRequest("GET", promURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", "-"+rangeStr)
	q.Add("end", "now")
	q.Add("step", step)
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

	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, err
	}

	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s", string(promResp.Data))
	}

	var metricsData MetricsData
	if err := json.Unmarshal(promResp.Data, &metricsData); err != nil {
		return nil, err
	}

	if len(metricsData.Result) == 0 {
		return nil, nil
	}

	// Parse the results
	series := &MetricSeries{
		Points: make([]Point, 0),
	}

	var sum float64
	var count int
	var max, min float64 = -1e9, 1e9

	for _, result := range metricsData.Result {
		// Handle both instant query (value) and range query (values)
		var values []interface{}
		if result.Values != nil && len(result.Values) > 0 {
			values = result.Values
		} else if result.Value != nil && len(result.Value) > 0 {
			values = []interface{}{result.Value}
		}

		for _, v := range values {
			point, ok := v.([]interface{})
			if !ok || len(point) < 2 {
				continue
			}

			timestamp, ok := point[0].(float64)
			if !ok {
				continue
			}

			valueStr, ok := point[1].(string)
			if !ok {
				continue
			}

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			series.Points = append(series.Points, Point{
				Timestamp: int64(timestamp),
				Value:     value,
			})

			sum += value
			count++
			if value > max {
				max = value
			}
			if value < min {
				min = value
			}
		}
	}

	if count > 0 {
		series.Average = sum / float64(count)
		series.Max = max
		series.Min = min
		if len(series.Points) > 0 {
			series.Current = series.Points[len(series.Points)-1].Value
		}
	}

	return series, nil
}
