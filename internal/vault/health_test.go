package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newHealthServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func TestHealthCheck_Healthy(t *testing.T) {
	srv := newHealthServer(http.StatusOK)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL, 5*time.Second)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Initialized {
		t.Error("expected Initialized to be true")
	}
	if status.Sealed {
		t.Error("expected Sealed to be false")
	}
	if !status.IsReady() {
		t.Error("expected IsReady to return true")
	}
}

func TestHealthCheck_Sealed(t *testing.T) {
	srv := newHealthServer(http.StatusServiceUnavailable)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL, 5*time.Second)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Sealed {
		t.Error("expected Sealed to be true")
	}
	if status.IsReady() {
		t.Error("expected IsReady to return false for sealed vault")
	}
}

func TestHealthCheck_Standby(t *testing.T) {
	srv := newHealthServer(http.StatusTooManyRequests)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL, 5*time.Second)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Standby {
		t.Error("expected Standby to be true")
	}
}

func TestHealthCheck_Uninitialized(t *testing.T) {
	srv := newHealthServer(http.StatusNotImplemented)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL, 5*time.Second)
	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Initialized {
		t.Error("expected Initialized to be false")
	}
}

func TestHealthCheck_UnexpectedStatus(t *testing.T) {
	srv := newHealthServer(http.StatusInternalServerError)
	defer srv.Close()

	checker := NewHealthChecker(srv.URL, 5*time.Second)
	_, err := checker.Check(context.Background())
	if err == nil {
		t.Fatal("expected error for unexpected status code")
	}
}

func TestHealthCheck_InvalidAddress(t *testing.T) {
	checker := NewHealthChecker("http://127.0.0.1:0", 1*time.Second)
	_, err := checker.Check(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable address")
	}
}
