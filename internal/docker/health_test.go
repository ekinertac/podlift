package docker

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckHealth_Success(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := HealthCheckConfig{
		URL:      ts.URL,
		Expect:   []int{200},
		Timeout:  5 * time.Second,
		Interval: 1 * time.Second,
		Retries:  3,
	}

	err := CheckHealth(cfg)
	if err != nil {
		t.Errorf("CheckHealth() error = %v, want nil", err)
	}
}

func TestCheckHealth_WrongStatusCode(t *testing.T) {
	// Create test server returning 500
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := HealthCheckConfig{
		URL:      ts.URL,
		Expect:   []int{200},
		Timeout:  1 * time.Second,
		Interval: 100 * time.Millisecond,
		Retries:  2,
	}

	err := CheckHealth(cfg)
	if err == nil {
		t.Error("CheckHealth() should fail with wrong status code")
	}

	if !strings.Contains(err.Error(), "unexpected status code") {
		t.Errorf("Error should mention status code, got: %v", err)
	}
}

func TestCheckHealth_ConnectionFailed(t *testing.T) {
	cfg := HealthCheckConfig{
		URL:      "http://localhost:99999", // Invalid port
		Expect:   []int{200},
		Timeout:  1 * time.Second,
		Interval: 100 * time.Millisecond,
		Retries:  2,
	}

	err := CheckHealth(cfg)
	if err == nil {
		t.Error("CheckHealth() should fail with connection error")
	}
}

func TestCheckHealth_EventualSuccess(t *testing.T) {
	attempts := 0
	
	// Server fails first 2 attempts, succeeds on 3rd
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	cfg := HealthCheckConfig{
		URL:      ts.URL,
		Expect:   []int{200},
		Timeout:  1 * time.Second,
		Interval: 100 * time.Millisecond,
		Retries:  5,
	}

	err := CheckHealth(cfg)
	if err != nil {
		t.Errorf("CheckHealth() should eventually succeed, got error: %v", err)
	}

	if attempts < 3 {
		t.Errorf("Should have retried, attempts = %d", attempts)
	}
}

func TestCheckHealth_MultipleExpectedCodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMovedPermanently) // 301
	}))
	defer ts.Close()

	cfg := HealthCheckConfig{
		URL:      ts.URL,
		Expect:   []int{200, 301, 302}, // Accept multiple codes
		Timeout:  1 * time.Second,
		Retries:  1,
	}

	err := CheckHealth(cfg)
	if err != nil {
		t.Errorf("CheckHealth() should accept 301, got error: %v", err)
	}
}

