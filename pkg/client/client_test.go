package client_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/client"
)

func TestNewClient(t *testing.T) {
	cfg := client.Config{
		BaseURL:   "https://api.confluent.cloud",
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if c == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestClientDo_SuccessfulRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-key:test-secret"))
		if authHeader != expectedAuth {
			t.Errorf("Expected auth header %q, got %q", expectedAuth, authHeader)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"result": "success"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/test/path",
	}

	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	err = resp.DecodeJSON(&result)
	if err != nil {
		t.Fatalf("DecodeJSON failed: %v", err)
	}

	if result["result"] != "success" {
		t.Errorf("Expected result 'success', got %q", result["result"])
	}
}

func TestClientDo_POSTWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]string{"id": "cluster-123"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	body := map[string]string{"name": "test-cluster"}
	req := client.Request{
		Method: "POST",
		Path:   "/clusters",
		Body:   body,
	}

	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestClientDo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "not found"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters/nonexistent",
	}

	resp, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 404 status, got nil")
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestClientDo_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := client.Request{
		Method: "GET",
		Path:   "/slow",
	}

	_, err = c.Do(ctx, req)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestClientDo_UnauthorizedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "bad-key",
		APISecret: "bad-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters",
	}

	resp, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 401 status, got nil")
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestResponse_DecodeJSON(t *testing.T) {
	testData := map[string]interface{}{
		"id":   "cluster-123",
		"name": "my-cluster",
	}

	jsonBody, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	resp := client.Response{
		StatusCode: http.StatusOK,
		Body:       jsonBody,
	}

	var result map[string]interface{}
	err = resp.DecodeJSON(&result)
	if err != nil {
		t.Fatalf("DecodeJSON failed: %v", err)
	}

	if result["id"] != "cluster-123" {
		t.Errorf("Expected id 'cluster-123', got %v", result["id"])
	}
}

func TestResponse_DecodeJSONError(t *testing.T) {
	resp := client.Response{
		StatusCode: http.StatusOK,
		Body:       []byte("invalid json"),
	}

	var result map[string]interface{}
	err := resp.DecodeJSON(&result)
	if err == nil {
		t.Fatal("Expected DecodeJSON to fail with invalid JSON")
	}
}

func TestClientDo_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters",
	}

	resp, err := c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 429 status, got nil")
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", resp.StatusCode)
	}
}

func BenchmarkClientDo(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"result": "ok"}); err != nil {
			b.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		b.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Do(context.Background(), req)
	}
}

// Error type tests
func TestClientDo_Error_IsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error_code": "NOT_FOUND", "message": "Resource not found"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters/nonexistent",
	}

	_, err = c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 404 status, got nil")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("Expected *api.Error, got %T", err)
	}

	if !apiErr.IsNotFound() {
		t.Errorf("Expected IsNotFound() to return true")
	}

	if apiErr.Code != http.StatusNotFound {
		t.Errorf("Expected code 404, got %d", apiErr.Code)
	}

	if apiErr.ErrorCode != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got %s", apiErr.ErrorCode)
	}
}

func TestClientDo_Error_IsUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]string{"error_code": "UNAUTHORIZED", "message": "Invalid credentials"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "bad-key",
		APISecret: "bad-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters",
	}

	_, err = c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 401 status, got nil")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("Expected *api.Error, got %T", err)
	}

	if !apiErr.IsUnauthorized() {
		t.Errorf("Expected IsUnauthorized() to return true")
	}
}

func TestClientDo_Error_IsRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		if err := json.NewEncoder(w).Encode(map[string]string{"error_code": "RATE_LIMIT_EXCEEDED", "message": "Too many requests"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters",
	}

	_, err = c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 429 status, got nil")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("Expected *api.Error, got %T", err)
	}

	if !apiErr.IsRateLimited() {
		t.Errorf("Expected IsRateLimited() to return true")
	}

	if !apiErr.IsRetryable() {
		t.Errorf("Expected IsRetryable() to return true for rate limiting")
	}

	retryAfter := apiErr.RetryAfter()
	if retryAfter == 0 {
		t.Errorf("Expected RetryAfter() > 0, got %d", retryAfter)
	}
}

func TestClientDo_Error_IsBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]string{"error_code": "INVALID_REQUEST", "message": "Invalid request payload"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "POST",
		Path:   "/clusters",
		Body:   map[string]string{"invalid": "body"},
	}

	_, err = c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 400 status, got nil")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("Expected *api.Error, got %T", err)
	}

	if !apiErr.IsBadRequest() {
		t.Errorf("Expected IsBadRequest() to return true")
	}
}

func TestClientDo_Error_IsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{"error_code": "INTERNAL_SERVER_ERROR", "message": "Internal server error"}); err != nil {
			t.Errorf("failed to write JSON response: %v", err)
		}
	}))
	defer server.Close()

	cfg := client.Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	req := client.Request{
		Method: "GET",
		Path:   "/clusters",
	}

	_, err = c.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for 500 status, got nil")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("Expected *api.Error, got %T", err)
	}

	if !apiErr.IsInternalServerError() {
		t.Errorf("Expected IsInternalServerError() to return true")
	}

	if !apiErr.IsRetryable() {
		t.Errorf("Expected IsRetryable() to return true for server errors")
	}
}
