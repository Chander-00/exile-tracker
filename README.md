# Exile Tracker

A Go backend service for tracking Path of Exile characters and their build snapshots.
It provides a REST API, an SSH TUI, a background data fetcher, and generates Path of Building import codes automatically.

## Why

I am really bad at understanding build upgrades in Path of Exile — when and which upgrades to do.
So I wanted to track good players to see which updates they make to their builds over time.

## Features

- **REST API** for managing accounts, characters, and passive skill snapshots
- **SSH TUI** — connect via SSH to browse snapshots directly from a terminal
- **PoB Export** — automatically generates Path of Building import codes for each snapshot
- **Background fetcher** periodically pulls character data from the PoE API
- **SQLite** database with migration support (Goose)
- **Structured logging** with zerolog
- **Docker deployment** with nginx reverse proxy

## Project Structure

```
cmd/
  main.go             # Application entrypoint
  api/                # API server setup
config/               # Configuration loading
db/                   # Database and migrations
models/               # Data models (internal and API)
poeclient/            # Path of Exile API client
repository/           # Database access layer
services/             # Business logic and background fetcher
utils/                # Logging and helpers
nginx/                # Nginx reverse proxy config
migrations/           # SQL migration files
```

## Getting Started

### Prerequisites

- Go 1.20+
- [Goose](https://github.com/pressly/goose) for migrations
- [templ](https://templ.guide/) for template generation

### Local Development

1. **Clone the repository**
   ```sh
   git clone https://github.com/Chander-00/exile-tracker.git
   cd exile-tracker
   ```

2. **Copy and configure environment variables**
   ```sh
   cp .env.example .env
   # Edit .env and set your values (API_KEY, SSH_ADMIN_KEY, etc.)
   ```

3. **Run in development mode**
   ```sh
   make dev
   ```

   Or build and run manually:
   ```sh
   make build
   ./bin/exile-tracker
   ```

### Docker

1. **Copy and configure environment variables**
   ```sh
   cp .env.example .env
   ```

   At minimum, edit these in `.env`:
   - `API_KEY` — set a real secret for API authentication
   - `SSH_ADMIN_KEY` — your SSH public key for admin access to the TUI

   The rest of the defaults work out of the box for Docker.

2. **Build and start all services**
   ```sh
   make docker-up
   ```

   This builds the Docker image and starts both the app and the nginx reverse proxy.

3. **Verify it's running**
   ```sh
   # Check container status
   docker compose ps

   # Check logs
   make docker-logs

   # Test the API (HTTP, nginx redirects to HTTPS)
   curl -k https://localhost/api/v1/pobsnapshots/character/1

   # Test SSH TUI
   ssh -p 2222 localhost
   ```

4. **Stop everything**
   ```sh
   make docker-down
   ```

#### SSL Certificates for Local Testing

The nginx service needs SSL certificates to start. For local development you need to generate
self-signed certificates. These are fake certificates that make HTTPS work on your machine — browsers
will show a "Not Secure" warning, but everything works fine for testing.

Run this once (you don't need to run it again unless you delete the files):

```sh
mkdir -p nginx/certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/certs/origin-key.pem \
  -out nginx/certs/origin.pem \
  -subj '/CN=localhost'
```

This creates two files:
- `nginx/certs/origin.pem` — the certificate (public)
- `nginx/certs/origin-key.pem` — the private key

Both are gitignored so they won't be committed.

When testing with `curl`, use the `-k` flag to skip certificate verification (since it's self-signed):
```sh
curl -k https://localhost/api/v1/pobsnapshots/character/1
```

> **For production/VPS deployment**, you do NOT use self-signed certs. Instead you use real certificates
> from Cloudflare. See the [VPS Deployment Guide](docs/VPS_DEPLOYMENT_GUIDE.md) for details.

### Make Targets

Run `make help` to see all available targets:

| Target | Description |
|--------|-------------|
| `make dev` | Run in development mode |
| `make build` | Build the application |
| `make run` | Build and run |
| `make test` | Run tests |
| `make docker-up` | Build and start Docker services |
| `make docker-down` | Stop Docker services |
| `make docker-logs` | Tail logs from all services |
| `make lint` | Run linter |
| `make fmt` | Format code |

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### PoB Snapshots

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/pobsnapshots/character/{characterId}` | List snapshots for a character |
| `GET` | `/pobsnapshots/character/{characterId}/latest` | Get latest snapshot |
| `GET` | `/pobsnapshots/{id}` | Get snapshot by ID |

## Environment Variables

See [`.env.example`](.env.example) for all available configuration options with descriptions.

## License

MIT

## Credits

- [Path of Exile](https://www.pathofexile.com/)
- [Path of Building](https://github.com/PathOfBuildingCommunity/PathOfBuilding)
- [Goose](https://github.com/pressly/goose)
- [templ](https://templ.guide/)
- [chi](https://github.com/go-chi/chi)
- [zerolog](https://github.com/rs/zerolog)
- [wish](https://github.com/charmbracelet/wish) (SSH TUI)
