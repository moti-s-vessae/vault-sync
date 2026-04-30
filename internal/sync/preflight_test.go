package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newPreflightServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func TestPreflightChecker_Run_Healthy(t *testing.T) {
	srv := newPreflightServer(http.StatusOK)
	defer srv.Close()

	checker := NewPreflightChecker(PreflightConfig{
		VaultAddress: srv.URL,
		Timeout:      5 * time.Second,
	})

	if err := checker.Run(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPreflightChecker_Run_Sealed(t *testing.T) {
	srv := newPreflightServer(http.StatusServiceUnavailable)
	defer srv.Close()

	checker := NewPreflightChecker(PreflightConfig{
		VaultAddress: srv.URL,
		Timeout:      5 * time.Second,
	})

	err := checker.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for sealed vault")
	}
	if !strings.Contains(err.Error(), "sealed") {
		t.Errorf("expected 'sealed' in error, got: %v", err)
	}
}

func TestPreflightChecker_Run_Uninitialized(t *testing.T) {
	srv := newPreflightServer(http.StatusNotImplemented)
	defer srv.Close()

	checker := NewPreflightChecker(PreflightConfig{
		VaultAddress: srv.URL,
		Timeout:      5 * time.Second,
	})

	err := checker.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for uninitialized vault")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected 'not initialized' in error, got: %v", err)
	}
}

func TestPreflightChecker_Run_Unreachable(t *testing.T) {
	checker := NewPreflightChecker(PreflightConfig{
		VaultAddress: "http://127.0.0.1:0",
		Timeout:      1 * time.Second,
	})

	err := checker.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable vault")
	}
	if !strings.Contains(err.Error(), "preflight failed") {
		t.Errorf("expected 'preflight failed' in error, got: %v", err)
	}
}

func TestNewPreflightChecker_DefaultTimeout(t *testing.T) {
	checker := NewPreflightChecker(PreflightConfig{
		VaultAddress: "http://localhost:8200",
	})
	if checker == nil {
		t.Fatal("expected non-nil checker")
	}
}
