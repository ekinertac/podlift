package docker

import (
	"fmt"
	"net/http"
	"time"
)

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	URL      string
	Expect   []int
	Timeout  time.Duration
	Interval time.Duration
	Retries  int
}

// CheckHealth performs HTTP health check
func CheckHealth(cfg HealthCheckConfig) error {
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	if cfg.Retries == 0 {
		cfg.Retries = 3
	}
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Second
	}

	var lastErr error
	
	for i := 0; i < cfg.Retries; i++ {
		if i > 0 {
			time.Sleep(cfg.Interval)
		}

		resp, err := client.Get(cfg.URL)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Check if status code is expected
		for _, expected := range cfg.Expect {
			if resp.StatusCode == expected {
				return nil // Success
			}
		}

		lastErr = fmt.Errorf("unexpected status code: %d (expected %v)", resp.StatusCode, cfg.Expect)
	}

	return fmt.Errorf("health check failed after %d attempts: %w", cfg.Retries, lastErr)
}

// WaitForHealth waits for a container to become healthy
func WaitForHealth(host string, port int, path string, timeout time.Duration) error {
	url := fmt.Sprintf("http://%s:%d%s", host, port, path)

	cfg := HealthCheckConfig{
		URL:      url,
		Expect:   []int{200, 301, 302},
		Timeout:  5 * time.Second,
		Interval: 2 * time.Second,
		Retries:  int(timeout.Seconds() / 2), // Retry for duration of timeout
	}

	return CheckHealth(cfg)
}

