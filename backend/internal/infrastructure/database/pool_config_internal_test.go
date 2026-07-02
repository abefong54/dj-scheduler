package database

import (
	"testing"
	"time"
)

// TestPoolConfigAppliesLimits checks the connection-pool tuning is wired onto the
// parsed config. Keeping a warm floor of connections (MinConns) avoids paying a
// TLS + auth handshake to Neon on the first query after an idle gap, which is
// what makes buttons feel laggy on a serverless Postgres.
func TestPoolConfigAppliesLimits(t *testing.T) {
	cfg, err := poolConfig("postgres://user:pass@localhost:5432/db?sslmode=disable")
	if err != nil {
		t.Fatalf("poolConfig: %v", err)
	}

	if cfg.MaxConns != 10 {
		t.Errorf("MaxConns = %d, want 10", cfg.MaxConns)
	}
	if cfg.MinConns != 2 {
		t.Errorf("MinConns = %d, want 2", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != 30*time.Minute {
		t.Errorf("MaxConnLifetime = %s, want 30m", cfg.MaxConnLifetime)
	}
	if cfg.MaxConnIdleTime != 10*time.Minute {
		t.Errorf("MaxConnIdleTime = %s, want 10m", cfg.MaxConnIdleTime)
	}
	if cfg.HealthCheckPeriod != time.Minute {
		t.Errorf("HealthCheckPeriod = %s, want 1m", cfg.HealthCheckPeriod)
	}
}

func TestPoolConfigRejectsBadURL(t *testing.T) {
	if _, err := poolConfig("://not-a-url"); err == nil {
		t.Fatal("expected error for malformed connection string, got nil")
	}
}
