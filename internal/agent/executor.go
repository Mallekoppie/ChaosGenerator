package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

// UseCasePayloads defines request payloads for each use case
var UseCasePayloads = map[string]string{
	"uc1":  "",                                                                                                                       // Health check - no body
	"uc2":  "",                                                                                                                       // Status - no body
	"uc3":  "",                                                                                                                       // Simple query - no body
	"uc4":  `{"username":"newuser","email":"newuser@example.com","firstName":"Jane","lastName":"Smith","password":"encrypted_hash"}`, // Create resource
	"uc5":  "",                                                                                                                       // List collection - no body
	"uc6":  generateBulkUpdatePayload(),                                                                                              // Bulk update
	"uc7":  `{"reportType":"transactionSummary","dateRange":{"start":"2026-01-01","end":"2026-02-14"},"format":"json"}`,              // Report generation
	"uc8":  generateLargePayload(10 * 1024 * 1024),                                                                                   // File upload - 10MB
	"uc9":  "",                                                                                                                       // Data export - no body
	"uc10": generateLargePayload(100 * 1024 * 1024),                                                                                  // Network saturation - 100MB
	"uc11": `{"streamId":"stream-12345","windowSize":60,"aggregations":["count","sum","average"],"buffer":{"size":10000}}`,           // Memory pressure
}

// HTTPExecutor handles HTTP requests to target services
type HTTPExecutor struct {
	client           *http.Client
	clientNoPool     *http.Client
	targetAddress    string
	targetProtocol   string
	connectionPooled bool
	useCaseID        string
	method           string
	path             string
	testExecutionID  string
	sni              string
}

// NewHTTPExecutor creates a new HTTP executor
func NewHTTPExecutor(targetAddress, targetProtocol string, connectionPooled bool, useCaseID, method, path, testExecutionID, sni string) *HTTPExecutor {
	// Configure TLS with SNI if provided
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // For testing purposes
	}
	if sni != "" {
		tlsConfig.ServerName = sni
	}

	// Client with connection pooling
	pooledTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     tlsConfig,
	}
	pooledTransport.Protocols = new(http.Protocols)
	pooledTransport.Protocols.SetHTTP1(true)
	pooledTransport.Protocols.SetHTTP2(true)

	// Client without connection pooling (recreate connection each time)
	noPoolTransport := &http.Transport{
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 0,
		DisableKeepAlives:   true,
		TLSClientConfig:     tlsConfig,
	}
	noPoolTransport.Protocols = new(http.Protocols)
	noPoolTransport.Protocols.SetHTTP1(true)
	// Disable HTTP/2 for no-pool client to ensure new connection each time (HTTP/2 connections are long-lived)
	// This is also how we can simulate the overhead of connection establishment for each request when connection pooling is disabled
	noPoolTransport.Protocols.SetHTTP2(false)

	executor := &HTTPExecutor{
		client: &http.Client{
			Transport: pooledTransport,
			Timeout:   120 * time.Second,
		},
		clientNoPool: &http.Client{
			Transport: noPoolTransport,
			Timeout:   120 * time.Second,
		},
		targetAddress:    targetAddress,
		targetProtocol:   targetProtocol,
		connectionPooled: connectionPooled,
		useCaseID:        useCaseID,
		method:           method,
		path:             path,
		testExecutionID:  testExecutionID,
		sni:              sni,
	}

	return executor
}

// Execute performs a single HTTP request and records metrics
func (e *HTTPExecutor) Execute(ctx context.Context) error {
	start := time.Now()
	var connStart, dnsStart, tlsStart time.Time
	var connDuration time.Duration

	// Create HTTP trace to measure connection time
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				_ = time.Since(dnsStart) // DNS resolution time (not separately tracked)
			}
		},
		ConnectStart: func(network, addr string) {
			connStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if !connStart.IsZero() {
				connDuration = time.Since(connStart)
				AgentConnectionTime.WithLabelValues(e.targetProtocol, e.useCaseID, fmt.Sprintf("%v", e.connectionPooled)).Observe(connDuration.Seconds())
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if !tlsStart.IsZero() {
				_ = time.Since(tlsStart) // TLS handshake time (not separately tracked)
			}
		},
	}

	// Build full URL
	url := strings.TrimSuffix(e.targetAddress, "/") + e.path

	// Add query params for use case 5 (list collection)
	if e.useCaseID == "uc5" {
		url += "?page=1&limit=50"
	}

	// Add path parameter for use case 3 (simple query)
	if e.useCaseID == "uc3" {
		url = strings.TrimSuffix(url, "/") + "/user-12345"
	}

	// Get payload for use case
	payload := UseCasePayloads[e.useCaseID]
	var req *http.Request
	var err error

	if payload != "" {
		req, err = http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), e.method, url, bytes.NewBufferString(payload))
	} else {
		req, err = http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), e.method, url, nil)
	}

	if err != nil {
		AgentErrorsTotal.WithLabelValues(e.targetProtocol, e.useCaseID, "request_creation", fmt.Sprintf("%v", e.connectionPooled)).Inc()
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "ChaosAgent/1.0")
	if payload != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Select client based on connection pooling setting
	client := e.client
	if !e.connectionPooled {
		client = e.clientNoPool
	}

	// Execute request
	responseStart := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		duration := time.Since(start).Seconds()
		AgentErrorsTotal.WithLabelValues(e.targetProtocol, e.useCaseID, "request_failed", fmt.Sprintf("%v", e.connectionPooled)).Inc()
		AgentRequestDuration.WithLabelValues(e.targetProtocol, e.useCaseID, e.method, "error", fmt.Sprintf("%v", e.connectionPooled)).Observe(duration)
		AgentRequestsTotal.WithLabelValues(e.targetProtocol, e.useCaseID, e.method, "error", fmt.Sprintf("%v", e.connectionPooled)).Inc()
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Record response time (time to first byte)
	responseTime := time.Since(responseStart).Seconds()
	AgentResponseTime.WithLabelValues(e.targetProtocol, e.useCaseID, fmt.Sprintf("%v", e.connectionPooled)).Observe(responseTime)

	// Read response body
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		AgentErrorsTotal.WithLabelValues(e.targetProtocol, e.useCaseID, "response_read", fmt.Sprintf("%v", e.connectionPooled)).Inc()
	}

	// Record metrics
	duration := time.Since(start).Seconds()
	statusCode := fmt.Sprintf("%d", resp.StatusCode)

	AgentRequestDuration.WithLabelValues(e.targetProtocol, e.useCaseID, e.method, statusCode, fmt.Sprintf("%v", e.connectionPooled)).Observe(duration)
	AgentProcessingTime.WithLabelValues(e.targetProtocol, e.useCaseID, fmt.Sprintf("%v", e.connectionPooled)).Observe(duration)
	AgentRequestsTotal.WithLabelValues(e.targetProtocol, e.useCaseID, e.method, statusCode, fmt.Sprintf("%v", e.connectionPooled)).Inc()

	// Record success/error
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		AgentSuccessTotal.WithLabelValues(e.targetProtocol, e.useCaseID, fmt.Sprintf("%v", e.connectionPooled)).Inc()
	} else {
		AgentErrorsTotal.WithLabelValues(e.targetProtocol, e.useCaseID, fmt.Sprintf("http_%d", resp.StatusCode), fmt.Sprintf("%v", e.connectionPooled)).Inc()
	}

	return nil
}

// generateBulkUpdatePayload generates a bulk update payload (~100KB)
func generateBulkUpdatePayload() string {
	updates := make([]string, 100)
	for i := 0; i < 100; i++ {
		updates[i] = fmt.Sprintf(`{"id":"user-%d","fields":{"lastLogin":"2026-02-14T10:30:00Z","preferences.theme":"dark"}}`, i)
	}
	return fmt.Sprintf(`{"updates":[%s],"options":{"validateAll":true,"atomic":false}}`, strings.Join(updates, ","))
}

// generateLargePayload generates a large payload of specified size in bytes
func generateLargePayload(sizeBytes int) string {
	// Generate records with realistic-looking data
	recordSize := 200 // Approximate size per record
	recordCount := sizeBytes / recordSize

	var sb strings.Builder
	sb.WriteString(`{"records":[`)

	for i := 0; i < recordCount; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`{"id":%d,"timestamp":"2026-02-14T10:30:%02d.000Z","value":%d.%d,"data":"sample_data_payload_%d"}`,
			i, i%60, i%100, i%100, i))
	}

	sb.WriteString(`]}`)
	return sb.String()
}
