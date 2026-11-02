<div align="center">

# ⚡ go-csupa

**Ultra-lightweight, zero-dependency Go CLI & Daemon to keep Supabase Free Tier projects active.**

[![CI](https://github.com/kevadamar/go-csupa/actions/workflows/ci.yml/badge.svg)](https://github.com/kevadamar/go-csupa/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/kevadamar/go-csupa.svg)](https://pkg.go.dev/github.com/kevadamar/go-csupa)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kevadamar/go-csupa)](https://golang.org)

</div>

---

## 📌 Overview

Supabase automatically pauses inactive free-tier projects after 7 days without database activity. **`go-csupa`** (*Go Cron Supabase*) is a minimal, production-grade CLI/daemon written entirely using the **Go Standard Library** with zero third-party dependencies.

It executes an initial ping immediately on boot, and then periodically queries your Supabase PostgREST endpoint every **3 days (72 hours)** by default, keeping your database warm and preventing auto-pauses.

---

## ✨ Features

- **Zero External Dependencies**: Built with 100% Go Standard Library (`net/http`, `os/signal`, `time`, `context`, etc.).
- **Immediate + Periodic Execution**: Executes an initial query on startup, followed by steady scheduled executions via `time.Ticker`.
- **Graceful Shutdown**: Intercepts `SIGINT` (Ctrl+C) and `SIGTERM` to safely terminate ongoing requests and clean up resources.
- **Configurable Interval & Timeout**: Defaults to 72 hours and 30-second timeouts, both customizable via environment variables.
- **Microsecond Timestamp Logging**: Formatted logging with latency duration and response body inspection.
- **Container Ready**: Includes a tiny, multi-stage Alpine Dockerfile (~15MB image size).

---

## ⚙️ Configuration (Environment Variables)

| Variable | Required | Default | Description |
| :--- | :---: | :---: | :--- |
| `SUPABASE_URL` | **Yes** | — | Supabase project URL (e.g., `https://xyzproject.supabase.co`) |
| `SUPABASE_ANON_KEY` | **Yes** | — | Supabase public anonymous API key |
| `SUPABASE_TABLE` | **Yes** | — | Table name to query (e.g., `users`, `profiles`) |
| `KEEP_ALIVE_INTERVAL` | No | `72h` | Ticker interval (e.g., `72h`, `24h`, `10s` for testing) |
| `HTTP_TIMEOUT` | No | `30s` | Maximum HTTP request timeout |

### Under the Hood
Each ping executes an authenticated PostgREST GET request:
```http
GET {SUPABASE_URL}/rest/v1/{SUPABASE_TABLE}?select=id&limit=1 HTTP/1.1
apikey: {SUPABASE_ANON_KEY}
Authorization: Bearer {SUPABASE_ANON_KEY}
Accept: application/json
```

---

## 🚀 Getting Started

### 1. Run with Go CLI

```bash
# Clone the repository
git clone https://github.com/kevadamar/go-csupa.git
cd go-csupa

# Copy and fill your environment variables
cp .env.example .env

# Export variables and run
export $(cat .env | xargs) && go run main.go
```

### 2. Compile Standalone Binary

```bash
CGO_ENABLED=0 go build -ldflags="-w -s" -o csupa .
./csupa
```

---

## 🐳 Deployment Options

### Option A: Docker / Docker Compose

Build and run with Docker:
```bash
docker build -t go-csupa .

docker run -d \
  --name go-csupa \
  --restart unless-stopped \
  --env-file .env \
  go-csupa
```

Using `docker-compose.yml`:
```yaml
services:
  supabase-keepalive:
    build: .
    restart: unless-stopped
    env_file:
      - .env
```

### Option B: Linux Systemd Service

Create `/etc/systemd/system/csupa.service`:

```ini
[Unit]
Description=Supabase Keep-Alive Daemon (go-csupa)
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/go-csupa
ExecStart=/opt/go-csupa/csupa
Restart=always
RestartSec=10

Environment="SUPABASE_URL=https://your-project.supabase.co"
Environment="SUPABASE_ANON_KEY=your-anon-key"
Environment="SUPABASE_TABLE=users"
Environment="KEEP_ALIVE_INTERVAL=72h"

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now csupa
sudo journalctl -u csupa -f
```

---

## 🧪 Testing

Run unit tests:
```bash
go test -v -race ./...
```

---

## 🤝 Contributing

Contributions are welcome! Please check [CONTRIBUTING.md](CONTRIBUTING.md) and the [Code of Conduct](CODE_OF_CONDUCT.md).

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
