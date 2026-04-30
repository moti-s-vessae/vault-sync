package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthStatus represents the result of a Vault health check.
type HealthStatus struct {
	Initialized bool
	Sealed      bool
	Standby     bool
	Version     string
	ClusterName string
}

// HealthChecker checks the health of a Vault server.
type HealthChecker struct {
	address    string
	httpClient *http.Client
}

// NewHealthChecker creates a new HealthChecker for the given Vault address.
func NewHealthChecker(address string, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		address: address,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Check performs a health check against the Vault server.
// It returns a HealthStatus and any error encountered.
func (h *HealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
	url := fmt.Sprintf("%s/v1/sys/health?standbyok=true&sealedok=false", h.address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building health request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault health check failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return &HealthStatus{Initialized: true, Sealed: false, Standby: false}, nil
	case http.StatusTooManyRequests:
		return &HealthStatus{Initialized: true, Sealed: false, Standby: true}, nil
	case http.StatusNotImplemented:
		return &HealthStatus{Initialized: false, Sealed: false}, nil
	case http.StatusServiceUnavailable:
		return &HealthStatus{Initialized: true, Sealed: true}, nil
	default:
		return nil, fmt.Errorf("unexpected vault health status code: %d", resp.StatusCode)
	}
}

// IsReady returns true if Vault is initialized and unsealed.
func (s *HealthStatus) IsReady() bool {
	return s.Initialized && !s.Sealed
}
