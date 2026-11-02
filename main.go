package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Config menyimpan konfigurasi dari Environment Variables
type Config struct {
	URL      string
	AnonKey  string
	Table    string
	Interval time.Duration
	Timeout  time.Duration
}

// loadConfig membaca dan memvalidasi environment variables
func loadConfig() (*Config, error) {
	supabaseURL := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	if supabaseURL == "" {
		return nil, fmt.Errorf("environment variable SUPABASE_URL wajib diisi")
	}

	anonKey := strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
	if anonKey == "" {
		return nil, fmt.Errorf("environment variable SUPABASE_ANON_KEY wajib diisi")
	}

	table := strings.TrimSpace(os.Getenv("SUPABASE_TABLE"))
	if table == "" {
		return nil, fmt.Errorf("environment variable SUPABASE_TABLE wajib diisi")
	}

	// Default interval: 72 jam (3 hari) sesuai ketentuan
	interval := 72 * time.Hour
	if rawInterval := strings.TrimSpace(os.Getenv("KEEP_ALIVE_INTERVAL")); rawInterval != "" {
		parsedInterval, err := time.ParseDuration(rawInterval)
		if err != nil {
			log.Printf("[WARN] Format KEEP_ALIVE_INTERVAL tidak valid ('%s'), menggunakan default 72h: %v", rawInterval, err)
		} else if parsedInterval <= 0 {
			log.Printf("[WARN] KEEP_ALIVE_INTERVAL harus > 0, menggunakan default 72h")
		} else {
			interval = parsedInterval
		}
	}

	// Default timeout HTTP: 30 detik
	timeout := 30 * time.Second
	if rawTimeout := strings.TrimSpace(os.Getenv("HTTP_TIMEOUT")); rawTimeout != "" {
		if parsedTimeout, err := time.ParseDuration(rawTimeout); err == nil && parsedTimeout > 0 {
			timeout = parsedTimeout
		}
	}

	return &Config{
		URL:      strings.TrimRight(supabaseURL, "/"),
		AnonKey:  anonKey,
		Table:    table,
		Interval: interval,
		Timeout:  timeout,
	}, nil
}

// pingSupabase mengirim HTTP GET request ke PostgREST API Supabase
func pingSupabase(ctx context.Context, client *http.Client, cfg *Config) {
	endpoint := fmt.Sprintf("%s/rest/v1/%s?select=id&limit=1", cfg.URL, url.PathEscape(cfg.Table))

	reqCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("[ERROR] Gagal membuat HTTP request: %v", err)
		return
	}

	// Header wajib PostgREST Supabase
	req.Header.Set("apikey", cfg.AnonKey)
	req.Header.Set("Authorization", "Bearer "+cfg.AnonKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Supabase-KeepAlive-Daemon/1.0")

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		if ctx.Err() != nil {
			log.Printf("[WARN] Request dibatalkan karena proses shutdown")
			return
		}
		log.Printf("[ERROR] Request gagal setelah %v: %v", duration.Round(time.Millisecond), err)
		return
	}
	defer resp.Body.Close()

	// Batasi pembacaan respons maksimal 2KB untuk log
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	bodyPreview := strings.TrimSpace(string(bodyBytes))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("[SUCCESS] Ping berhasil! Status: %d (%s) | Waktu respons: %v | Respons: %s",
			resp.StatusCode, resp.Status, duration.Round(time.Millisecond), bodyPreview)
	} else {
		log.Printf("[WARN] Supabase merespons status non-2xx: %d (%s) | Waktu respons: %v | Detail: %s",
			resp.StatusCode, resp.Status, duration.Round(time.Millisecond), bodyPreview)
	}
}

func main() {
	// Konfigurasi format logger dengan tanggal, waktu, dan mikrodetik
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[supabase-keepalive] ")

	log.Println("Memulai Supabase Keep-Alive Daemon...")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("[FATAL] Konfigurasi tidak valid: %v", err)
	}

	log.Printf("[INFO] Target URL: %s", cfg.URL)
	log.Printf("[INFO] Target Table: %s", cfg.Table)
	log.Printf("[INFO] Interval Ticker: %v (Next run in ~%v)", cfg.Interval, cfg.Interval)
	log.Printf("[INFO] HTTP Request Timeout: %v", cfg.Timeout)

	// Inisialisasi HTTP Client dengan timeout dasar
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Setup context untuk menangani Graceful Shutdown (SIGINT & SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Eksekusi pertama kali secara langsung saat aplikasi dimulai
	log.Println("[INFO] Menjalankan ping perdana (initial execution)...")
	pingSupabase(ctx, client, cfg)

	// Buat time.Ticker untuk jadwal berikutnya
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.Printf("[INFO] Daemon aktif. Menunggu jadwal ticker berikutnya setiap %v...", cfg.Interval)

	// Loop utama daemon
	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] Sinyal terminasi (SIGINT/SIGTERM) diterima. Melakukan graceful shutdown...")
			// Berikan jeda sejenak jika ada pembersihan tambahan sebelum exit
			log.Println("[INFO] Daemon berhenti dengan bersih. Sampai jumpa!")
			return

		case t := <-ticker.C:
			log.Printf("[INFO] Ticker aktif pada %s. Menjalankan keep-alive ping...", t.Format(time.RFC3339))
			pingSupabase(ctx, client, cfg)
			log.Printf("[INFO] Ping selesai. Jadwal berikutnya: %s", t.Add(cfg.Interval).Format(time.RFC3339))
		}
	}
}
