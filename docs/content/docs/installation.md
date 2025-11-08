---
title: Installation
weight: 20
---

# Installation

Complete guide to installing and setting up podlift.

## Prerequisites

### On Your Machine

- **Operating System**: Linux, macOS, or WSL2 on Windows
- **Git**: For version tracking
- **SSH Client**: For server communication
- **Docker** (optional): Only needed if testing locally

### On Your Servers

- **Operating System**: Ubuntu 20.04+, Debian 11+, or similar Linux distribution
- **SSH Server**: Accessible from your machine
- **Docker**: Version 20.10 or later
- **Open Ports**: 22 (SSH), 80 (HTTP), 443 (HTTPS)

## Install podlift

Choose the method that works best for you:

| Method | Best For | Updates |
|--------|----------|---------|
| Install Script | Quick setup, any platform | Re-run script |
| Download Binary | Offline installs, specific versions | Manual download |
| Go Install | Go developers | `go install ...@latest` |
| Build from Source | Contributors, custom builds | `git pull && go build` |

### Option 1: Install Script (Recommended)

**System-wide installation** (requires sudo):

```bash
curl -sSL https://raw.githubusercontent.com/ekinertac/podlift/main/install.sh | sh
```

**User-level installation** (no sudo required):

```bash
curl -sSL https://raw.githubusercontent.com/ekinertac/podlift/main/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
```

Then add to your shell profile (~/.bashrc or ~/.zshrc):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

The system-wide install places the binary in `/usr/local/bin`. The user-level install uses `~/.local/bin`, which doesn't require admin privileges.

### Option 2: Download Binary

Visit [GitHub Releases](https://github.com/ekinertac/podlift/releases) and download for your platform:

```bash
# Linux (amd64)
wget https://github.com/ekinertac/podlift/releases/latest/download/podlift-linux-amd64
chmod +x podlift-linux-amd64
sudo mv podlift-linux-amd64 /usr/local/bin/podlift

# macOS (arm64 - M1/M2)
wget https://github.com/ekinertac/podlift/releases/latest/download/podlift-darwin-arm64
chmod +x podlift-darwin-arm64
sudo mv podlift-darwin-arm64 /usr/local/bin/podlift

# macOS (amd64 - Intel)
wget https://github.com/ekinertac/podlift/releases/latest/download/podlift-darwin-amd64
chmod +x podlift-darwin-amd64
sudo mv podlift-darwin-amd64 /usr/local/bin/podlift
```

### Option 3: Go Install

If you have Go 1.21+ installed, use Go's package manager:

```bash
go install github.com/ekinertac/podlift@latest
```

This compiles and installs to `$GOPATH/bin` (usually `~/go/bin`).

Make sure `~/go/bin` is in your PATH:
```bash
export PATH=$PATH:~/go/bin
```

### Option 4: Build from Source

For development or custom builds:

```bash
git clone https://github.com/ekinertac/podlift.git
cd podlift
go build -o podlift .
sudo mv podlift /usr/local/bin/
```

### Verify Installation

```bash
podlift version
```

Should output:
```
podlift v0.1.0
Go version: go1.21.0
OS/Arch: linux/amd64
```

## Server Setup

### Install Docker

On each server you want to deploy to:

#### Ubuntu/Debian

```bash
# Quick install
curl -fsSL https://get.docker.com | sh

# Enable and start Docker
sudo systemctl enable docker
sudo systemctl start docker

# Verify
docker --version
```

#### Manual Installation (Ubuntu 22.04)

```bash
# Update package index
sudo apt update

# Install dependencies
sudo apt install -y ca-certificates curl gnupg lsb-release

# Add Docker's GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

# Enable and start
sudo systemctl enable docker
sudo systemctl start docker

# Verify
docker --version
```

### Configure SSH Access

#### Generate SSH Key (if needed)

On your machine:

```bash
# Generate key
ssh-keygen -t ed25519 -C "your_email@example.com"

# Press Enter to accept default location
# Set passphrase (or press Enter for none)
```

#### Copy SSH Key to Server

```bash
# Copy key
ssh-copy-id root@192.168.1.10

# Test connection
ssh root@192.168.1.10
```

If `ssh-copy-id` doesn't work:

```bash
# Manual method
cat ~/.ssh/id_ed25519.pub | ssh root@192.168.1.10 "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

#### Configure SSH Config (Optional)

Add to `~/.ssh/config`:

```
Host myserver
  HostName 192.168.1.10
  User root
  IdentityFile ~/.ssh/id_ed25519
```

Then use in `podlift.yml`:
```yaml
servers:
  - host: myserver  # Uses SSH config
```

### Open Firewall Ports

#### Using UFW (Ubuntu)

```bash
# SSH
sudo ufw allow 22/tcp

# HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

#### Using iptables

```bash
# SSH
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# HTTP/HTTPS
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Save rules
sudo iptables-save > /etc/iptables/rules.v4
```

### Test Server Readiness

From your machine:

```bash
# Test SSH
ssh root@192.168.1.10 'docker --version'
```

Should output Docker version.

## Initialize Your Project

### Navigate to Your Application

```bash
cd /path/to/your/app
```

Your app should have:
- A `Dockerfile`
- Git repository initialized

### Initialize podlift

```bash
podlift init
```

This creates:
- `podlift.yml` - Configuration file
- `.env.example` - Example environment variables

### Configure podlift.yml

Edit `podlift.yml`:

```yaml
service: myapp
image: myapp

servers:
  - host: 192.168.1.10
    user: root
    ssh_key: ~/.ssh/id_ed25519
```

Minimal working config. See [Configuration Reference](configuration.md) for all options.

### Set Up Environment Variables

Create `.env` in the same directory as `podlift.yml`:

```bash
# Copy example
cp .env.example .env

# Edit with your values
vim .env
```

Example `.env`:
```bash
SECRET_KEY=your-django-secret-key
DB_PASSWORD=secure-postgres-password
```

**Important**: 
- The `.env` file must be in the same directory as `podlift.yml`
- Add `.env` to `.gitignore`:

```bash
echo ".env" >> .gitignore
git add .gitignore
git commit -m "Ignore .env file"
```

### Validate Configuration

```bash
podlift validate
```

Should output:
```
✓ Configuration valid
✓ SSH connection to 192.168.1.10
✓ Docker installed (v24.0.5)
✓ Required ports available
✓ Disk space sufficient
✓ Git state clean

Ready to deploy!
```

If validation fails, see [Troubleshooting](troubleshooting.md).

## First Deployment

### Ensure Clean Git State

```bash
git status
```

If you have uncommitted changes:
```bash
git add -A
git commit -m "Initial podlift setup"
```

### Deploy

```bash
podlift deploy
```

You'll see:
```
[1/7] Validating configuration...
  ✓ Configuration valid

[2/7] Building image myapp:a1b2c3d...
  ✓ Built in 45s

[3/7] Pushing to server...
  ✓ Uploaded 234MB in 12s

[4/7] Loading image on server...
  ✓ Loaded in 8s

[5/7] Starting containers...
  ✓ myapp-web-a1b2c3d-1 started
  ✓ Health check passed (5s)

[6/7] Updating nginx configuration...
  ✓ Traffic routing to new version

[7/7] Stopping old containers...
  ✓ Complete

✓ Deployment successful!

URL: http://192.168.1.10
```

### Verify Deployment

```bash
# Check status
podlift ps

# View logs
podlift logs web

# Test in browser
curl http://192.168.1.10
```

## Set Up SSL (Optional)

### Prerequisites

- Domain pointing to your server
- Ports 80 and 443 open

### Update Configuration

Edit `podlift.yml`:
```yaml
domain: myapp.com

proxy:
  ssl: letsencrypt
  ssl_email: admin@myapp.com
```

### Set Up SSL

```bash
podlift ssl setup --email admin@myapp.com
```

Output:
```
Setting up SSL for myapp.com...

[1/3] Installing certbot...
  ✓ Installed

[2/3] Obtaining certificate...
  ✓ Certificate obtained

[3/3] Configuring nginx...
  ✓ HTTPS enabled

✓ SSL configured!
Your site is now available at: https://myapp.com
```

### Test HTTPS

```bash
curl https://myapp.com
```

Certificate auto-renews via cron.

## Next Steps

You're ready to deploy! Common workflows:

### Make Changes and Deploy

```bash
# Make code changes
vim app/views.py

# Commit changes
git add -A
git commit -m "Fix bug"

# Deploy
podlift deploy
```

### Rollback if Needed

```bash
podlift rollback
```

### View Logs

```bash
podlift logs web --follow
```

### Execute Commands

```bash
podlift exec web python manage.py migrate
```

## Upgrading podlift

### Check Current Version

```bash
podlift version
```

### Upgrade to Latest

**Go install:**
```bash
go install github.com/ekinertac/podlift@latest
```

**Install script:**
```bash
curl -sSL https://raw.githubusercontent.com/ekinertac/podlift/main/install.sh | sh
```

**Manual:** Download latest release from GitHub.

### Verify Upgrade

```bash
podlift version
```

## Uninstall

Remove binary:
```bash
sudo rm /usr/local/bin/podlift
```

Remove configuration (optional):
```bash
rm podlift.yml
rm .env
```

Server cleanup (if no longer using):
```bash
ssh root@192.168.1.10

# Stop all containers
docker stop $(docker ps -q)

# Remove containers
docker rm $(docker ps -aq)

# Remove images
docker rmi $(docker images -q)

# Remove nginx config
rm /etc/nginx/sites-available/myapp
rm /etc/nginx/sites-enabled/myapp
systemctl reload nginx
```

## Cloud Provider Setup

### DigitalOcean

Create droplet:
```bash
# Via web UI or API
doctl compute droplet create myapp \
  --image ubuntu-22-04-x64 \
  --size s-1vcpu-1gb \
  --region nyc1
```

Get IP and use in `podlift.yml`.

### AWS EC2

Launch instance:
```bash
# Via AWS Console or CLI
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.micro \
  --key-name your-key
```

Update security group to allow ports 22, 80, 443.

### Hetzner

Create server:
```bash
# Via Hetzner Cloud Console
hcloud server create \
  --name myapp \
  --type cx11 \
  --image ubuntu-22.04
```

Get IP and configure podlift.

### Linode

Launch instance via Linode Manager or API.

All cloud providers work the same way:
1. Create Ubuntu instance
2. Get public IP
3. Configure SSH access
4. Use IP in `podlift.yml`

## Getting Help

- [Commands Reference](commands.md)
- [Configuration Reference](configuration.md)
- [Troubleshooting](troubleshooting.md)
- [How It Works](how-it-works.md)

Still stuck? [Open an issue](https://github.com/ekinertac/podlift/issues).

