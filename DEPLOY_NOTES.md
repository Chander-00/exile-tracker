# Deployment Notes

## Architecture

```
Internet
  │
  ├── Port 80/443 (HTTP/HTTPS) ──→ [Nginx container] ──→ [exile-tracker container :3000]
  │
  └── Port 2222 (SSH) ──────────→ [exile-tracker container :2222] (direct)
```

- **Nginx container**: Reverse proxy, TLS termination (Cloudflare Origin Certificate)
- **exile-tracker container**: Go monolith (API + Web UI + SSH TUI + background fetcher)
- **Docker volumes**: SQLite database, SSH host key, TLS certificates

## First Time Setup

```bash
# 1. SSH into your VPS
ssh user@your-vps-ip

# 2. Install Docker (one-time)
curl -fsSL https://get.docker.com | sh

# 3. Clone your repo
git clone https://github.com/Chander-00/exile-tracker.git
cd exile-tracker

# 4. Create your .env file with production values
cp .env.example .env
nano .env  # set your API_KEY, SSH_ADMIN_KEY, domain, etc.

# 5. Build and start everything
docker compose up -d --build
```

## Updating the App

```bash
cd exile-tracker
git pull                           # Get latest code
docker compose up -d --build       # Rebuild and restart
```

## Updating PoB

Same as updating the app — PoB gets cloned fresh during Docker build:

```bash
docker compose up -d --build
```

## Useful Commands

```bash
docker compose logs -f              # All services
docker compose logs -f exile-tracker # Just your app
docker compose logs -f nginx         # Just nginx
docker compose down                  # Stop containers (data persists in volumes)
docker compose ps                    # Check running containers
```

## Volumes (persistent data)

- `sqlite-data` — SQLite database file
- `ssh-keys` — SSH host key (prevents "host key changed" warnings)
- `./nginx/certs/` — TLS certificates (bind-mounted into Nginx, not a Docker volume)

## SSH Host Key (how it works)

The SSH host key identifies the server to clients. Without it persisting,
users would see "WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!" on every
container restart.

**You don't need to generate it manually.** Here's what happens:

1. Container starts for the first time
2. The app (via the `wish` library) looks for the key at `.ssh/exile_tracker_ed25519`
3. Key doesn't exist → `wish` auto-generates an ed25519 key pair and saves it
4. The `.ssh/` directory is mounted to a Docker volume (`ssh-keys`), so the key persists
5. On next restart → key already exists → same key → no client warnings

If you ever need to reset the host key (e.g., migrating to a new server):
```bash
# Remove the volume (clients will need to re-accept the new key)
docker compose down
docker volume rm exile-tracker_ssh-keys
docker compose up -d
```

## Cloudflare Origin Certificate Setup

1. Log into Cloudflare dashboard
2. Select your domain → SSL/TLS → Origin Server
3. Click "Create Certificate"
4. Keep defaults (RSA 2048, 15 years validity)
5. Copy the certificate and private key
6. On your VPS, save them:
   ```bash
   nano nginx/certs/origin.pem      # Paste the certificate
   nano nginx/certs/origin-key.pem  # Paste the private key
   ```
7. In Cloudflare DNS settings, add an A record pointing to your VPS IP with proxy enabled (orange cloud)
8. In SSL/TLS settings, set mode to "Full (strict)"
