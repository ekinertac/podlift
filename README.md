# podlift

**Simple, transparent deployment for containerized applications.**

podlift deploys your Docker containers to any server with SSH access. No black boxes, no magic, no broken promises.

## Why podlift?

**Transparent**: See exactly what commands run on your servers. No custom infrastructure, just standard Docker + nginx.

**Reliable**: Validate everything before deployment. Clear errors with actionable solutions.

**Non-Interactive**: Every command works in CI/CD. No prompts, no waiting for input.

**Zero-Downtime**: Deploy new versions without dropping requests. Rollback instantly if needed.

**Git-Native**: Every deployment maps to a commit. Always know what's running in production.

## Quick Start

Install podlift:
```bash
# macOS
brew install ekinertac/tap/podlift

# Linux (Debian/Ubuntu)
curl -fsSL https://apt.podlift.sh/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/podlift.gpg
echo "deb [signed-by=/usr/share/keyrings/podlift.gpg] https://apt.podlift.sh stable main" | sudo tee /etc/apt/sources.list.d/podlift.list
sudo apt update && sudo apt install podlift

# Or use install script
curl -sSL https://podlift.sh/install.sh | sh

# Or with Go
go install github.com/ekinertac/podlift@latest
```

Initialize your project:
```bash
cd myapp/
podlift init
```

Edit `podlift.yml` to add your server:
```yaml
service: myapp
image: myapp

servers:
  - host: 192.168.1.10
    user: root
    ssh_key: ~/.ssh/id_rsa
```

Deploy:
```bash
podlift deploy
```

That's it. Your app is live.

## What Happens During Deploy

```bash
$ podlift deploy

[1/7] Validating configuration...
  ✓ SSH connection to 192.168.1.10
  ✓ Docker installed (v24.0.5)
  ✓ Git state clean (commit: a1b2c3d)
  
[2/7] Building image myapp:a1b2c3d...
  ✓ Built in 45s

[3/7] Pushing to server...
  ✓ Uploaded 234MB in 12s
  
[4/7] Loading image on server...
  ✓ Loaded in 8s
  
[5/7] Starting new containers...
  ✓ myapp-web-a1b2c3d-1 started
  ✓ Health check passed (5s)
  
[6/7] Updating nginx configuration...
  ✓ Traffic routing to new version
  
[7/7] Stopping old containers...
  ✓ myapp-web-x9y8z7w-1 stopped

✓ Deployment successful!

Deployed: a1b2c3d "Fix critical bug"
URL: http://192.168.1.10
Time: 1m 34s
```

## Core Commands

```bash
# Initialize configuration
podlift init

# Validate setup before deploying
podlift validate

# Deploy your application
podlift deploy

# Check running services
podlift ps

# View logs
podlift logs web
podlift logs web --follow

# Rollback to previous version
podlift rollback
podlift rollback --to v1.2.3

# Set up SSL with Let's Encrypt
podlift ssl setup --email admin@myapp.com
```

## Features

### Zero-Downtime Deployments
New containers start alongside old ones. Traffic switches only after health checks pass. No dropped requests.

### Git-Based Versioning
Every deployment uses your current git commit. Dirty working trees are rejected. Production state always matches git history.

### Multi-Server Support
Deploy to multiple servers with shared dependencies. Workers, web servers, and databases all configured in one file.

### Dependency Management
Run PostgreSQL, Redis, or any Docker container alongside your app. Dependencies persist across deployments.

### Flexible Build Strategies
- Build locally + SCP to server (no registry needed)
- Build locally + push to registry
- Support for GitHub Container Registry, Docker Hub, or private registries

### Standard Tools
Uses nginx for reverse proxy, Docker for containers, standard SSH for deployment. Everything is debuggable.

## Configuration Example

```yaml
service: myapp
domain: myapp.com

servers:
  web:
    - host: 192.168.1.10
      user: root
      labels: [primary]
    - host: 192.168.1.11
      user: root

dependencies:
  postgres:
    image: postgres:16
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
  
  redis:
    image: redis:7

services:
  web:
    port: 8000
    replicas: 2
    healthcheck:
      path: /health
      expect: [200, 301]
    env:
      SECRET_KEY: ${SECRET_KEY}
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp

  worker:
    command: celery -A myapp worker
    replicas: 1

proxy:
  enabled: true
  ssl: letsencrypt
  ssl_email: admin@myapp.com
```

## Documentation

- [Installation](docs/installation.md) - Detailed setup guide
- [Commands](docs/commands.md) - Complete CLI reference
- [Configuration](docs/configuration.md) - All configuration options
- [Deployment Guide](docs/deployment-guide.md) - Strategies, rollbacks, multi-server
- [Troubleshooting](docs/troubleshooting.md) - Common errors and solutions
- [How It Works](docs/how-it-works.md) - Architecture and internals
- [Migration](docs/migration.md) - Moving from Kamal, Ansible, or manual deploys

## Principles

**Transparency Over Magic**: You should understand what's happening. Run `podlift deploy --dry-run` to see every command before execution.

**Standards Over Reinvention**: Uses nginx, Docker Compose patterns, and standard tools. No custom infrastructure.

**Fail Fast With Solutions**: Validation catches errors before deployment. Error messages include how to fix them.

**Non-Interactive By Default**: Every command works in CI/CD without prompts.

**Escape Hatches Everywhere**: Skip steps with flags when needed. Manual intervention is always possible.

## Requirements

- Docker on your deployment servers
- SSH access with key-based authentication
- Git (for version tracking)
- Linux servers (Ubuntu 20.04+, Debian 11+, or similar)

## Project Status

podlift is under active development. See [ROADMAP.md](ROADMAP.md) for planned features.

Current version: **v0.1.0-alpha** (MVP phase)

## License

MIT

## Acknowledgments

Inspired by Kamal's vision of simple deployment, but built for reliability and transparency.

