package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type PrometheusCollector struct {
	serverURL string
	client    *http.Client
}

type PrometheusMetrics struct {
	HTTPRequestDuration map[string][]float64 `json:"http_request_duration"`
	HTTPRequestTotal    map[string]int64     `json:"http_request_total"`
	HTTPInFlight        int                  `json:"http_in_flight"`
	CustomMetrics       map[string]float64   `json:"custom_metrics"`
	Timestamp           time.Time            `json:"timestamp"`
}

type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func NewPrometheusCollector(serverURL string) *PrometheusCollector {
	return &PrometheusCollector{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *PrometheusCollector) Collect(ctx context.Context, duration time.Duration) (*PrometheusMetrics, error) {
	metrics := &PrometheusMetrics{
		HTTPRequestDuration: make(map[string][]float64),
		HTTPRequestTotal:    make(map[string]int64),
		CustomMetrics:       make(map[string]float64),
		Timestamp:           time.Now(),
	}

	endTime := time.Now()
	startTime := endTime.Add(-duration)

	queries := map[string]string{
		"request_duration_p99": fmt.Sprintf(
			`histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{job="bank-api"}[%s])) by (endpoint, le))`,
			duration.String(),
		),
		"request_duration_p95": fmt.Sprintf(
			`histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="bank-api"}[%s])) by (endpoint, le))`,
			duration.String(),
		),
		"request_duration_p50": fmt.Sprintf(
			`histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{job="bank-api"}[%s])) by (endpoint, le))`,
			duration.String(),
		),
		"request_total": fmt.Sprintf(
			`sum(rate(http_request_total{job="bank-api"}[%s])) by (endpoint, status)`,
			duration.String(),
		),
		"request_errors": fmt.Sprintf(
			`sum(rate(http_request_total{job="bank-api", status=~"4..|5.."}[%s])) by (endpoint)`,
			duration.String(),
		),
		"in_flight": `http_requests_in_flight{job="bank-api"}`,
	}

	for name, query := range queries {
		result, err := p.query(ctx, query, endTime)
		if err != nil {
			continue
		}

		for _, r := range result.Data.Result {
			endpoint := r.Metric["endpoint"]
			if endpoint == "" {
				endpoint = "total"
			}

			if len(r.Value) >= 2 {
				if valueStr, ok := r.Value[1].(string); ok {
					if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
						key := fmt.Sprintf("%s_%s", name, endpoint)
						
						switch name {
						case "request_total":
							status := r.Metric["status"]
							key = fmt.Sprintf("%s_%s_%s", name, endpoint, status)
							metrics.HTTPRequestTotal[key] = int64(value)
						case "in_flight":
							metrics.HTTPInFlight = int(value)
						default:
							metrics.CustomMetrics[key] = value
						}
					}
				}
			}
		}
	}

	rangeQueries := map[string]string{
		"request_duration": fmt.Sprintf(
			`http_request_duration_seconds{job="bank-api"}`,
		),
	}

	for name, query := range rangeQueries {
		result, err := p.queryRange(ctx, query, startTime, endTime, 30*time.Second)
		if err != nil {
			continue
		}

		for _, r := range result {
			endpoint := r.Metric["endpoint"]
			if endpoint == "" {
				continue
			}

			key := fmt.Sprintf("%s_%s", name, endpoint)
			if _, exists := metrics.HTTPRequestDuration[key]; !exists {
				metrics.HTTPRequestDuration[key] = []float64{}
			}

			for _, value := range r.Values {
				if len(value) >= 2 {
					if valueStr, ok := value[1].(string); ok {
						if v, err := strconv.ParseFloat(valueStr, 64); err == nil {
							metrics.HTTPRequestDuration[key] = append(metrics.HTTPRequestDuration[key], v)
						}
					}
				}
			}
		}
	}

	return metrics, nil
}

func (p *PrometheusCollector) query(ctx context.Context, query string, time time.Time) (*QueryResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("time", fmt.Sprintf("%d", time.Unix()))

	req, err := http.NewRequestWithContext(ctx, "GET", 
		fmt.Sprintf("%s/api/v1/query?%s", p.serverURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", result.Status)
	}

	return &result, nil
}

type RangeQueryResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

func (p *PrometheusCollector) queryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]RangeQueryResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", fmt.Sprintf("%d", int(step.Seconds())))

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/query_range?%s", p.serverURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string             `json:"resultType"`
			Result     []RangeQueryResult `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", result.Status)
	}

	return result.Data.Result, nil
}

func (p *PrometheusCollector) FilterAPIMetrics(metrics *PrometheusMetrics) *PrometheusMetrics {
	filtered := &PrometheusMetrics{
		HTTPRequestDuration: make(map[string][]float64),
		HTTPRequestTotal:    make(map[string]int64),
		CustomMetrics:       make(map[string]float64),
		Timestamp:           metrics.Timestamp,
	}

	for key, values := range metrics.HTTPRequestDuration {
		filtered.HTTPRequestDuration[key] = values
	}

	for key, value := range metrics.HTTPRequestTotal {
		filtered.HTTPRequestTotal[key] = value
	}

	filtered.HTTPInFlight = metrics.HTTPInFlight

	for key, value := range metrics.CustomMetrics {
		filtered.CustomMetrics[key] = value
	}

	return filtered
}