package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Simpan env asli
	origURL := os.Getenv("SUPABASE_URL")
	origKey := os.Getenv("SUPABASE_ANON_KEY")
	origTable := os.Getenv("SUPABASE_TABLE")
	origSchema := os.Getenv("SUPABASE_SCHEMA")
	origInterval := os.Getenv("KEEP_ALIVE_INTERVAL")
	origTimeout := os.Getenv("HTTP_TIMEOUT")

	defer func() {
		os.Setenv("SUPABASE_URL", origURL)
		os.Setenv("SUPABASE_ANON_KEY", origKey)
		os.Setenv("SUPABASE_TABLE", origTable)
		os.Setenv("SUPABASE_SCHEMA", origSchema)
		os.Setenv("KEEP_ALIVE_INTERVAL", origInterval)
		os.Setenv("HTTP_TIMEOUT", origTimeout)
	}()

	t.Run("Missing SUPABASE_URL", func(t *testing.T) {
		os.Unsetenv("SUPABASE_URL")
		os.Setenv("SUPABASE_ANON_KEY", "test-key")
		os.Setenv("SUPABASE_TABLE", "users")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("harus mengembalikan error jika SUPABASE_URL kosong")
		}
	})

	t.Run("Missing SUPABASE_ANON_KEY", func(t *testing.T) {
		os.Setenv("SUPABASE_URL", "https://xyz.supabase.co")
		os.Unsetenv("SUPABASE_ANON_KEY")
		os.Setenv("SUPABASE_TABLE", "users")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("harus mengembalikan error jika SUPABASE_ANON_KEY kosong")
		}
	})

	t.Run("Missing SUPABASE_TABLE", func(t *testing.T) {
		os.Setenv("SUPABASE_URL", "https://xyz.supabase.co")
		os.Setenv("SUPABASE_ANON_KEY", "test-key")
		os.Unsetenv("SUPABASE_TABLE")

		_, err := loadConfig()
		if err == nil {
			t.Fatal("harus mengembalikan error jika SUPABASE_TABLE kosong")
		}
	})

	t.Run("Valid Config with Defaults", func(t *testing.T) {
		os.Setenv("SUPABASE_URL", "https://xyz.supabase.co/")
		os.Setenv("SUPABASE_ANON_KEY", "secret-anon-key")
		os.Setenv("SUPABASE_TABLE", "profiles")
		os.Unsetenv("SUPABASE_SCHEMA")
		os.Unsetenv("KEEP_ALIVE_INTERVAL")
		os.Unsetenv("HTTP_TIMEOUT")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.URL != "https://xyz.supabase.co" {
			t.Errorf("expected URL without trailing slash, got: %s", cfg.URL)
		}
		if cfg.AnonKey != "secret-anon-key" {
			t.Errorf("expected secret-anon-key, got: %s", cfg.AnonKey)
		}
		if cfg.Table != "profiles" {
			t.Errorf("expected profiles, got: %s", cfg.Table)
		}
		if cfg.Schema != "" {
			t.Errorf("expected empty schema by default, got: %s", cfg.Schema)
		}
		if cfg.Interval != 72*time.Hour {
			t.Errorf("expected default interval 72h, got: %v", cfg.Interval)
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("expected default timeout 30s, got: %v", cfg.Timeout)
		}
	})

	t.Run("Auto Split Schema from SUPABASE_TABLE", func(t *testing.T) {
		os.Setenv("SUPABASE_URL", "https://xyz.supabase.co")
		os.Setenv("SUPABASE_ANON_KEY", "secret-anon-key")
		os.Setenv("SUPABASE_TABLE", "memfinance.user_profiles")
		os.Unsetenv("SUPABASE_SCHEMA")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Schema != "memfinance" {
			t.Errorf("expected schema 'memfinance', got: %s", cfg.Schema)
		}
		if cfg.Table != "user_profiles" {
			t.Errorf("expected table 'user_profiles', got: %s", cfg.Table)
		}
	})

	t.Run("Explicit SUPABASE_SCHEMA", func(t *testing.T) {
		os.Setenv("SUPABASE_URL", "https://xyz.supabase.co")
		os.Setenv("SUPABASE_ANON_KEY", "secret-anon-key")
		os.Setenv("SUPABASE_TABLE", "user_profiles")
		os.Setenv("SUPABASE_SCHEMA", "memfinance")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Schema != "memfinance" {
			t.Errorf("expected schema 'memfinance', got: %s", cfg.Schema)
		}
		if cfg.Table != "user_profiles" {
			t.Errorf("expected table 'user_profiles', got: %s", cfg.Table)
		}
	})

	t.Run("Custom Interval and Timeout", func(t *testing.T) {
		os.Setenv("SUPABASE_URL", "https://xyz.supabase.co")
		os.Setenv("SUPABASE_ANON_KEY", "secret-anon-key")
		os.Setenv("SUPABASE_TABLE", "users")
		os.Setenv("KEEP_ALIVE_INTERVAL", "1h30m")
		os.Setenv("HTTP_TIMEOUT", "10s")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Interval != 90*time.Minute {
			t.Errorf("expected interval 1h30m, got: %v", cfg.Interval)
		}
		if cfg.Timeout != 10*time.Second {
			t.Errorf("expected timeout 10s, got: %v", cfg.Timeout)
		}
	})
}

func TestPingSupabase(t *testing.T) {
	var requestReceived bool
	var receivedPath string
	var receivedRawQuery string
	var receivedAPIKey string
	var receivedAuth string
	var receivedProfile string

	// Buat mock server yang meniru Supabase PostgREST API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		receivedPath = r.URL.Path
		receivedRawQuery = r.URL.RawQuery
		receivedAPIKey = r.Header.Get("apikey")
		receivedAuth = r.Header.Get("Authorization")
		receivedProfile = r.Header.Get("Accept-Profile")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id": "a3b1-4c2d"}]`))
	}))
	defer server.Close()

	cfg := &Config{
		URL:      server.URL,
		AnonKey:  "my-anon-key-123",
		Schema:   "memfinance",
		Table:    "user_profiles",
		Interval: 72 * time.Hour,
		Timeout:  5 * time.Second,
	}

	client := server.Client()
	ctx := context.Background()

	pingSupabase(ctx, client, cfg)

	if !requestReceived {
		t.Fatal("expected request to be received by mock server")
	}

	expectedPath := "/rest/v1/user_profiles"
	if receivedPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, receivedPath)
	}

	expectedQuery := "select=id&limit=1"
	if receivedRawQuery != expectedQuery {
		t.Errorf("expected query %s, got %s", expectedQuery, receivedRawQuery)
	}

	if receivedAPIKey != cfg.AnonKey {
		t.Errorf("expected apikey header '%s', got '%s'", cfg.AnonKey, receivedAPIKey)
	}

	expectedAuth := "Bearer " + cfg.AnonKey
	if receivedAuth != expectedAuth {
		t.Errorf("expected Authorization header '%s', got '%s'", expectedAuth, receivedAuth)
	}

	if receivedProfile != "memfinance" {
		t.Errorf("expected Accept-Profile header 'memfinance', got '%s'", receivedProfile)
	}
}
