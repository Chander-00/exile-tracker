# VPS Deployment Guide — Exile Tracker

Step-by-step guide to deploy exile-tracker on a fresh VPS with Docker, Nginx, and Cloudflare.

## Prerequisites

- A VPS (any cheap one works — 1 CPU, 1GB RAM minimum)
- A domain name (we'll use Cloudflare for DNS and TLS)
- Your VPS IP address
- SSH access to your VPS (root or a user with sudo)

## Architecture Overview

```
Internet
  │
  ├── Port 80/443 (HTTP/HTTPS) ──→ [Nginx container] ──→ [exile-tracker :3000]
  │
  └── Port 2222 (SSH) ──────────→ [exile-tracker :2222] (direct, no proxy)
```

Two Docker containers run on the VPS:
- **exile-tracker**: Your Go app (API + Web UI + SSH TUI + background fetcher)
- **nginx**: Reverse proxy that handles HTTPS and forwards traffic to the app

---

## Step 1: Buy a Domain on Cloudflare

1. Go to https://dash.cloudflare.com
2. Click "Register Domains" in the left sidebar
3. Search for and purchase your domain
4. Done — Cloudflare is now your registrar AND DNS provider (no extra setup)

If you bought the domain elsewhere, you'll need to change the nameservers to
Cloudflare's. Cloudflare will walk you through this when you add the domain.

---

## Step 2: Point Your Domain to the VPS

1. In Cloudflare dashboard, select your domain
2. Go to **DNS** → **Records**
3. Click **Add record**:
   - Type: `A`
   - Name: `@` (this means your root domain, e.g., `yourdomain.com`)
   - IPv4 address: `YOUR_VPS_IP_ADDRESS`
   - Proxy status: **Proxied** (orange cloud ON)
   - TTL: Auto
4. Click **Save**

If you also want `www.yourdomain.com`:
- Add another `A` record with Name: `www`, same IP, Proxied

DNS propagation can take a few minutes to a few hours.

---

## Step 3: Generate the Cloudflare Origin Certificate

### What is this?

When someone visits your site, the traffic goes: **User → Cloudflare → Your VPS**.
Cloudflare handles the HTTPS that users see (the green lock in the browser).
But the connection between Cloudflare and your VPS also needs to be encrypted — that's what
this certificate does. It's called an "origin certificate" because your VPS is the "origin server".

You do NOT use the `openssl` self-signed cert command here — that's only for local testing.
Cloudflare gives you a real certificate for free that lasts 15 years.

1. In Cloudflare dashboard, select your domain
2. Go to **SSL/TLS** → **Origin Server**
3. Click **Create Certificate**
4. Keep the defaults:
   - Key type: RSA (2048)
   - Hostnames: `yourdomain.com` and `*.yourdomain.com`
   - Validity: 15 years
5. Click **Create**
6. **IMPORTANT**: You'll see two text blocks:
   - **Origin Certificate** — copy this and save it somewhere (you'll need it soon)
   - **Private Key** — copy this and save it somewhere (this is shown ONCE, you cannot retrieve it later)
7. Click **OK**

Now set the SSL mode:

1. Go to **SSL/TLS** → **Overview**
2. Set SSL/TLS encryption mode to **Full (strict)**

---

## Step 4: Prepare the VPS

SSH into your VPS:

```bash
ssh root@YOUR_VPS_IP
```

### 4.1: Update the system

```bash
# For Debian/Ubuntu:
apt update && apt upgrade -y

# For other distros, use your package manager
```

### 4.2: Create a non-root user (if you're logged in as root)

It's good practice to not run everything as root:

```bash
adduser deploy
usermod -aG sudo deploy
```

Switch to that user:

```bash
su - deploy
```

### 4.3: Install Docker

```bash
curl -fsSL https://get.docker.com | sh
```

Add your user to the docker group (so you don't need sudo for docker commands):

```bash
sudo usermod -aG docker $USER
```

**Log out and log back in** for the group change to take effect:

```bash
exit
ssh deploy@YOUR_VPS_IP
```

Verify Docker works:

```bash
docker --version
docker compose version
```

### 4.4: Open the required ports

Make sure your VPS firewall allows these ports:

- **80** (HTTP — Nginx will redirect to HTTPS)
- **443** (HTTPS — Nginx serves your web app)
- **2222** (SSH TUI — users connect to your app via SSH)
- **22** (standard SSH — so you can manage your VPS)

If you're using `ufw` (common on Ubuntu):

```bash
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 2222/tcp
sudo ufw enable
sudo ufw status
```

If your VPS provider has a web-based firewall (like DigitalOcean, Hetzner, etc.),
make sure those ports are also allowed there.

---

## Step 5: Clone the Repository

```bash
cd ~
git clone https://github.com/Chander-00/exile-tracker.git
cd exile-tracker
```

---

## Step 6: Set Up the TLS Certificates

Remember the certificate and private key you copied from Cloudflare in Step 3? Now you need
to put them on your VPS so nginx can use them to encrypt traffic.

These are NOT the self-signed certs from local testing. These are the real Cloudflare ones.

```bash
mkdir -p nginx/certs
```

Create the certificate file:

```bash
nano nginx/certs/origin.pem
```

Paste the **Origin Certificate** text (the one that starts with `-----BEGIN CERTIFICATE-----`
and ends with `-----END CERTIFICATE-----` — paste the whole thing including those lines).
Save and exit (Ctrl+X, then Y, then Enter).

Create the private key file:

```bash
nano nginx/certs/origin-key.pem
```

Paste the **Private Key** text (the one that starts with `-----BEGIN PRIVATE KEY-----`
and ends with `-----END PRIVATE KEY-----` — again, paste the whole thing).
Save and exit.

Set restrictive permissions on the key so only the owner can read it:

```bash
chmod 600 nginx/certs/origin-key.pem
```

Verify both files exist and aren't empty:

```bash
ls -la nginx/certs/
# You should see origin.pem and origin-key.pem, both with non-zero file sizes
```

---

## Step 7: Configure Environment Variables

```bash
cp .env.example .env
nano .env
```

The values you MUST change:

```bash
# Generate a strong random API key:
API_KEY=GENERATE_A_STRONG_RANDOM_STRING_HERE

# Your SSH public key for admin access to the TUI.
# Find yours with: cat ~/.ssh/id_ed25519.pub (or id_rsa.pub)
# If you don't have one, generate it with: ssh-keygen -t ed25519
SSH_ADMIN_KEY=ssh-ed25519 AAAAC3NzaC1lZDI1N... your-email@example.com
```

The rest of the values can stay as defaults. They are already configured for Docker.

To generate a random API key, you can use:

```bash
openssl rand -hex 32
```

---

## Step 8: Build and Start

```bash
docker compose up -d --build
```

This will:
1. Build the exile-tracker Docker image (compiles Go, clones PoB — takes a few minutes first time)
2. Pull the nginx:alpine image
3. Start both containers
4. Create Docker volumes for the database and SSH host key

Watch the logs to make sure everything starts correctly:

```bash
docker compose logs -f
```

You should see:
- `Database connection established`
- `Migrations completed successfully`
- `Starting API server, listening on :3000`
- `Starting SSH server`
- `Starting fetcher service`

Press Ctrl+C to stop watching logs (the containers keep running).

---

## Step 9: Verify Everything Works

### 9.1: Check containers are running

```bash
docker compose ps
```

You should see both `exile-tracker` and `nginx` with status "Up".

### 9.2: Test the web app

From your local machine (not the VPS), open a browser and go to:

```
https://yourdomain.com
```

You should see the exile-tracker web interface.

### 9.3: Test the API

```bash
curl -H "X-API-Key: YOUR_API_KEY" https://yourdomain.com/api/v1/accounts
```

Should return a JSON response.

### 9.4: Test the SSH TUI

From your local machine:

```bash
ssh -p 2222 yourdomain.com
```

You'll see:
1. First time: "Are you sure you want to continue connecting?" — type `yes`
2. The TUI should load with the exile-tracker dashboard

---

## Day-to-Day Operations

### Updating the app (after pushing new code)

```bash
cd ~/exile-tracker
git pull
docker compose up -d --build
```

### Updating Path of Building (after pushing changes to your PoB fork)

Same command — PoB gets cloned fresh during each Docker build:

```bash
cd ~/exile-tracker
docker compose up -d --build
```

### Viewing logs

```bash
docker compose logs -f                # All services
docker compose logs -f exile-tracker  # Just the app
docker compose logs -f nginx          # Just nginx
```

### Stopping everything

```bash
docker compose down
```

Your data is safe — it lives in Docker volumes that persist across stops/restarts.

### Restarting

```bash
docker compose up -d
```

### Checking disk usage

```bash
docker system df
```

### Cleaning up old Docker images (free disk space)

```bash
docker image prune -f
```

---

## Troubleshooting

### "Cannot connect to the Docker daemon"

Docker isn't running. Start it:

```bash
sudo systemctl start docker
```

To make Docker start automatically on boot:

```bash
sudo systemctl enable docker
```

### "permission denied" when running docker commands

Your user isn't in the docker group. Fix:

```bash
sudo usermod -aG docker $USER
```

Then log out and back in.

### Container starts but web app doesn't load

Check if the container is actually running:

```bash
docker compose ps
docker compose logs exile-tracker
```

Common issues:
- **"unable to open database file"**: The `/app/data` directory doesn't exist or the volume isn't mounted correctly
- **"Failed to run migrations"**: Database file is corrupted — you may need to delete the volume and start fresh

### HTTPS not working / SSL errors

1. Check that your Cloudflare SSL mode is set to "Full (strict)"
2. Check that the cert files exist and are correct:
   ```bash
   ls -la nginx/certs/
   ```
3. Check nginx logs:
   ```bash
   docker compose logs nginx
   ```
4. Make sure the DNS A record is pointing to your VPS IP with the orange cloud (Proxied) enabled

### SSH TUI not accessible

Make sure port 2222 is open in your firewall:

```bash
sudo ufw status
```

Test from your local machine:

```bash
ssh -v -p 2222 yourdomain.com
```

The `-v` flag shows verbose output to help debug connection issues.

### Need to start fresh (nuclear option)

This deletes ALL data (database, SSH keys, everything):

```bash
docker compose down -v   # -v removes volumes too
docker compose up -d --build
```

---

## Security Checklist

Before going live, make sure:

- [ ] You changed the `API_KEY` from the default
- [ ] You set your `SSH_ADMIN_KEY` to your actual public key
- [ ] The `.env` file is NOT committed to git (it's in `.gitignore`)
- [ ] The cert files in `nginx/certs/` are NOT committed (they're in `.gitignore`)
- [ ] Your VPS firewall only allows ports 22, 80, 443, and 2222
- [ ] Cloudflare SSL mode is "Full (strict)"
- [ ] You're not running as root (use a regular user with docker group access)
