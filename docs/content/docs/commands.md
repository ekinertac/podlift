---
title: Commands Reference
weight: 20
---

# Command Reference

Complete reference for all podlift commands.

## Global Flags

Available on all commands:

- `--config` - Path to config file (default: `podlift.yml`)
- `--verbose` - Show detailed output
- `--dry-run` - Show what would happen without executing

## podlift init

Initialize podlift configuration in the current directory.

```bash
podlift init
```

Creates:
- `podlift.yml` - Main configuration file
- `.env.example` - Example environment variables

Output:
```
Created podlift.yml
Created .env.example

Next steps:
  1. Edit podlift.yml and add your server
  2. Copy .env.example to .env and set secrets
  3. Run: podlift setup (prepare fresh servers)
  4. Run: podlift validate
```

## podlift setup

Prepare fresh servers for deployment.

```bash
podlift setup
```

Installs and configures everything needed on your servers:
- Docker installation
- Firewall configuration (ports 22, 80, 443)
- Basic security hardening
- Verification

### Flags

- `--no-firewall` - Skip firewall configuration
- `--no-security` - Skip security hardening

### What It Does

```bash
$ podlift setup

Setting up servers from podlift.yml...

Server: 192.168.1.10
  [1/4] Installing Docker...
    Downloading Docker installation script...
    Installing Docker 24.0.5...
    ✓ Docker installed

  [2/4] Configuring firewall...
    Opening port 22 (SSH)...
    Opening port 80 (HTTP)...
    Opening port 443 (HTTPS)...
    Enabling firewall...
    ✓ Firewall configured

  [3/4] Applying security settings...
    Disabling password authentication...
    Installing fail2ban...
    ✓ Security configured

  [4/4] Verification...
    ✓ Docker running
    ✓ Ports open
    ✓ SSH access working

✓ Server setup complete!

Next step: podlift deploy
```

### Idempotent

Safe to run multiple times:

```bash
$ podlift setup

Server: 192.168.1.10
  [1/4] Installing Docker...
    ✓ Docker 24.0.5 already installed

  [2/4] Configuring firewall...
    ✓ Firewall already configured

  # ... all checks pass ...

✓ Server already configured correctly!
```

### Error Handling

Clear errors if setup fails:

```bash
$ podlift setup

Server: 192.168.1.10
  [1/4] Installing Docker...
    ✗ Installation failed

ERROR: Docker installation failed

apt-get command not found. 

Supported operating systems:
  - Ubuntu 20.04+
  - Debian 11+
  - CentOS 8+

Manual installation: https://docs.docker.com/engine/install/
```

## podlift validate

Validate configuration and verify server readiness before deployment.

```bash
podlift validate
```

Checks:
- Configuration syntax
- SSH connectivity to all servers
- Docker installation and version
- Required ports availability (80, 443, app ports)
- Disk space
- Git repository state
- Environment variables
- Service name conflicts (detects if different app uses same name)

Output:
```
Validating configuration...

✓ Configuration valid
✓ SSH connection to 192.168.1.10
✓ SSH connection to 192.168.1.11
✓ Docker 24.0.5 installed on all servers
✓ Required ports available (80, 443, 8000)
✓ 45GB disk space available
✓ Git state clean (commit: a1b2c3d)
✓ Environment variables set (5 required)

Ready to deploy!
```

Error examples:

**Port conflict:**
```
Validating configuration...

✓ Configuration valid
✓ SSH connection to 192.168.1.10
✗ Port 5432 in use on 192.168.1.10

Port 5432 is required for postgres but already in use.

Check what's using it: ssh root@192.168.1.10 'lsof -i :5432'
```

**Service name conflict:**
```
Validating configuration...

✓ Configuration valid
✓ SSH connection to 192.168.1.10
✓ Docker 24.0.5 installed
⚠ Service 'myapp' already deployed on 192.168.1.10

Existing deployment:
  Version: x9y8z7w
  Deployed: 2 hours ago
  Containers: 2

If this is the same application:
  This is normal - deploying will update the existing deployment

If this is a DIFFERENT application:
  Change the service name in podlift.yml to avoid conflicts

Service names must be unique per server.
```

## podlift deploy

Deploy your application to all configured servers with zero-downtime (default).

```bash
podlift deploy
```

### Flags

- `--skip-build` - Skip building Docker image
- `--skip-healthcheck` - Skip health check
- `--parallel` - Deploy to all servers in parallel (default: serial)
- `--dry-run` - Show what would happen without executing
- `--zero-downtime` - Use zero-downtime deployment with nginx (default: true)

### Process

1. Validate configuration and git state
2. Build Docker image (tagged with git commit)
3. Transfer image to servers (SCP or registry)
4. Start new containers alongside old
5. Wait for health checks
6. Update nginx to route to new containers
7. Stop old containers
8. Cleanup unused images

### Examples

Standard deployment:
```bash
podlift deploy
```

Skip build if image already exists:
```bash
podlift deploy --skip-build
```

Deploy to all servers at once:
```bash
podlift deploy --parallel
```

See what would happen:
```bash
podlift deploy --dry-run
```

### Output

```
[1/7] Validating configuration...
  ✓ Configuration valid
  ✓ Git state clean (commit: a1b2c3d)

[2/7] Building image myapp:a1b2c3d...
  Building production stage...
  ✓ Built in 45s

[3/7] Pushing to server 192.168.1.10...
  ✓ Uploaded 234MB in 12s

[4/7] Loading image on server...
  ✓ Loaded in 8s

[5/7] Starting new containers...
  Starting myapp-web-a1b2c3d-1...
    ✓ Health check passed (5s)
  Starting myapp-web-a1b2c3d-2...
    ✓ Health check passed (4s)

[6/7] Updating nginx configuration...
  ✓ Traffic routing to new version

[7/7] Stopping old containers...
  ✓ myapp-web-x9y8z7w-1 stopped
  ✓ myapp-web-x9y8z7w-2 stopped

✓ Deployment successful!

Deployed: a1b2c3d "Fix critical bug"
Previous: x9y8z7w (available for rollback)
URL: http://192.168.1.10
Time: 1m 34s
```

### Error Handling

If deployment fails, old containers keep running. No downtime.

```
[5/7] Starting new containers...
  Starting myapp-web-a1b2c3d-1...
    ✗ Health check failed (timeout after 30s)

ERROR: Deployment failed

The new container started but failed health checks.

Container logs (last 20 lines):
  django.core.exceptions.ImproperlyConfigured: 
  ALLOWED_HOSTS must be set

Check logs: podlift logs web
Debug: podlift exec web bash

Old containers still running. No downtime occurred.
```

## podlift rollback

Revert to the previous deployment.

```bash
podlift rollback
```

### Flags

- `--to <version>` - Rollback to specific git commit or tag
- `--skip-healthcheck` - Don't wait for health checks

### Examples

Rollback to previous deployment:
```bash
podlift rollback
```

Rollback to specific version:
```bash
podlift rollback --to v1.2.3
podlift rollback --to a1b2c3d
```

### Output

```
Rolling back to x9y8z7w...

[1/4] Starting old containers...
  ✓ myapp-web-x9y8z7w-1 started
  ✓ Health check passed (3s)

[2/4] Updating nginx configuration...
  ✓ Traffic routing to x9y8z7w

[3/4] Stopping current containers...
  ✓ myapp-web-a1b2c3d-1 stopped

[4/4] Cleanup...
  ✓ Complete

✓ Rollback successful!

Current: x9y8z7w "Previous working version"
Time: 34s
```

## podlift ps

Show status of running services.

```bash
podlift ps
```

### Flags

- `--all`, `-a` - Show all containers (including stopped)

### Output

```
SERVICE    REPLICAS  STATUS   VERSION  UPTIME   
web        2/2       healthy  a1b2c3d  5m 23s
worker     1/1       healthy  a1b2c3d  5m 20s
postgres   1/1       healthy  -        2d 3h
redis      1/1       healthy  -        2d 3h
```

With `--all`:
```
SERVICE    REPLICAS  STATUS   VERSION  UPTIME   
web        2/2       healthy  a1b2c3d  5m 23s
web-old    0/2       stopped  x9y8z7w  -
worker     1/1       healthy  a1b2c3d  5m 20s
postgres   1/1       healthy  -        2d 3h
redis      1/1       healthy  -        2d 3h
```

## podlift logs

View container logs.

```bash
podlift logs <service>
```

### Flags

- `--follow`, `-f` - Stream logs in real-time
- `--tail <n>`, `-n` - Show last N lines (default: 100)
- `--since <time>` - Show logs since timestamp (e.g., "2h", "30m")

### Examples

View last 100 lines:
```bash
podlift logs web
```

Stream logs:
```bash
podlift logs web --follow
```

Last 500 lines:
```bash
podlift logs web --tail 500
```

Logs from last hour:
```bash
podlift logs web --since 1h
```

### Output

```
[web-1] [2025-11-05 10:30:15] INFO Starting server on :8000
[web-1] [2025-11-05 10:30:16] INFO Connected to database
[web-2] [2025-11-05 10:30:15] INFO Starting server on :8000
[web-2] [2025-11-05 10:30:16] INFO Connected to database
```

## podlift exec

Execute command in a running container.

```bash
podlift exec <service> <command>
```

### Flags

- `--replica <n>` - Execute on specific replica (default: 1)

### Examples

Interactive shell:
```bash
podlift exec web bash
```

Run management command:
```bash
podlift exec web python manage.py migrate
```

Check database connection:
```bash
podlift exec web python manage.py dbshell
```

Execute on specific replica:
```bash
podlift exec web --replica 2 bash
```

## podlift ssl

Manage SSL certificates.

```bash
podlift ssl <command>
```

### Commands

#### podlift ssl setup

Set up SSL with Let's Encrypt.

```bash
podlift ssl setup --email admin@example.com
```

Required:
- Domain configured in `podlift.yml`
- DNS pointing to your server
- Ports 80 and 443 open

Output:
```
Setting up SSL for myapp.com...

[1/3] Installing certbot...
  ✓ Installed

[2/3] Obtaining certificate...
  Validating domain ownership...
  ✓ Certificate obtained

[3/3] Configuring nginx...
  ✓ HTTPS enabled
  ✓ HTTP → HTTPS redirect enabled

✓ SSL configured!

Certificate: /etc/letsencrypt/live/myapp.com/fullchain.pem
Expires: 2026-02-03
Auto-renewal: enabled

Your site is now available at: https://myapp.com
```

#### podlift ssl renew

Manually renew SSL certificate.

```bash
podlift ssl renew
```

Normally runs automatically via cron.

#### podlift ssl status

Check SSL certificate status.

```bash
podlift ssl status
```

Output:
```
SSL Status for myapp.com:

Certificate: /etc/letsencrypt/live/myapp.com/fullchain.pem
Issued: 2025-11-05
Expires: 2026-02-03 (89 days remaining)
Auto-renewal: enabled (certbot.timer active)

✓ SSL is configured correctly
```

## podlift status

Show current deployment status.

```bash
podlift status
```

### Output

```
Current deployment:
  Version: a1b2c3d "Fix critical bug"
  Deployed: 5 minutes ago
  Services: 4/4 healthy
  URL: https://myapp.com

Previous deployment:
  Version: x9y8z7w "Add new feature"
  Status: Stopped (can rollback)

Servers:
  192.168.1.10 - healthy (2/2 web, postgres, redis)
  192.168.1.11 - healthy (2/2 web)

Available commands:
  podlift logs web
  podlift exec web bash
  podlift rollback
```

## podlift config

Show current configuration.

```bash
podlift config
```

### Flags

- `--show-secrets` - Include environment variable values

### Output

Without secrets:
```yaml
service: myapp
domain: myapp.com

servers:
  web:
    - host: 192.168.1.10
      user: root

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

services:
  web:
    port: 8000
    replicas: 2
```

## podlift version

Show podlift version.

```bash
podlift version
```

Output:
```
podlift v0.1.0
Go version: go1.21.0
OS/Arch: linux/amd64
```

## Exit Codes

All commands use standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Validation error
- `4` - Deployment error
- `5` - Connection error

This allows reliable scripting:

```bash
if podlift validate; then
  podlift deploy
else
  echo "Validation failed"
  exit 1
fi
```

