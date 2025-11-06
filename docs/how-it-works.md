# How podlift Works

Understanding what happens under the hood. No magic, no black boxes.

## Architecture Overview

```
┌─────────────────┐
│  Your Machine   │
│                 │
│  podlift CLI    │
└────────┬────────┘
         │
         │ SSH
         ▼
┌─────────────────────────────────┐
│  Server (192.168.1.10)          │
│                                 │
│  ┌────────────────────────┐    │
│  │  nginx (port 80/443)   │    │
│  │  ├─> web-abc123-1:8000 │    │
│  │  └─> web-abc123-2:8000 │    │
│  └────────────────────────┘    │
│                                 │
│  ┌──────────────────────────┐  │
│  │  App Containers          │  │
│  │  ├─ myapp-web-abc123-1   │  │
│  │  └─ myapp-web-abc123-2   │  │
│  └──────────────────────────┘  │
│                                 │
│  ┌──────────────────────────┐  │
│  │  Dependencies            │  │
│  │  ├─ postgres:16          │  │
│  │  └─ redis:7              │  │
│  └──────────────────────────┘  │
└─────────────────────────────────┘
```

## Components

### 1. podlift CLI

Single Go binary that runs on your machine. It:
- Parses `podlift.yml`
- Validates configuration
- Executes Docker commands locally
- Runs commands on servers via SSH
- Manages nginx configuration

### 2. Docker

Standard Docker installation on servers. podlift uses:
- `docker build` - Build images
- `docker save/load` - Transfer images via SCP
- `docker run` - Start containers
- `docker exec` - Run commands in containers
- `docker ps` - Check container status

No Docker Compose, no Swarm, no custom orchestration. Just plain Docker commands.

### 3. nginx

Reverse proxy that handles:
- Routing traffic to app containers
- SSL termination
- Zero-downtime deploys (upstream switching)
- Load balancing between replicas

Standard nginx. No custom builds, no plugins.

### 4. Certbot

Optional, for SSL. Standard Let's Encrypt client.

## Deployment Process

Step-by-step breakdown of `podlift deploy`.

### Step 1: Validation

```bash
[1/7] Validating configuration...
```

**What happens:**
1. Parse `podlift.yml` for syntax errors
2. Check git working tree is clean
3. Get current commit hash (e.g., `a1b2c3d`)
4. Verify all required environment variables exist
5. Test SSH connection to all servers
6. Check Docker is installed on servers
7. Verify required ports are available

**Commands executed:**
```bash
# On your machine
git status --porcelain
git rev-parse --short HEAD

# On each server (via SSH)
ssh root@192.168.1.10 'docker --version'
ssh root@192.168.1.10 'netstat -tuln | grep :80'
```

If any check fails, deployment stops before doing any work.

### Step 2: Build Image

```bash
[2/7] Building image myapp:a1b2c3d...
```

**What happens:**
1. Build Docker image from your `Dockerfile`
2. Tag with git commit: `myapp:a1b2c3d`
3. Save image to tar file (if using SCP method)

**Commands executed:**
```bash
# On your machine
docker build -t myapp:a1b2c3d .
docker save myapp:a1b2c3d -o /tmp/myapp-a1b2c3d.tar
```

The git commit is the version identifier. This ensures reproducibility.

### Step 3: Transfer Image

```bash
[3/7] Pushing to server 192.168.1.10...
```

**What happens (SCP method):**
1. SCP image tar to server
2. Progress bar shows upload

**What happens (Registry method):**
1. Push to registry: `docker push ghcr.io/user/myapp:a1b2c3d`
2. Pull on server: `docker pull ghcr.io/user/myapp:a1b2c3d`

**Commands executed:**
```bash
# SCP method
scp /tmp/myapp-a1b2c3d.tar root@192.168.1.10:/tmp/

# Registry method
docker push ghcr.io/user/myapp:a1b2c3d
ssh root@192.168.1.10 'docker pull ghcr.io/user/myapp:a1b2c3d'
```

### Step 4: Load Image

```bash
[4/7] Loading image on server...
```

**What happens (SCP only):**
1. Load tar file into Docker

**Commands executed:**
```bash
ssh root@192.168.1.10 'docker load -i /tmp/myapp-a1b2c3d.tar'
ssh root@192.168.1.10 'rm /tmp/myapp-a1b2c3d.tar'
```

### Step 5: Start New Containers

```bash
[5/7] Starting new containers...
```

**What happens:**
1. Start new containers alongside old ones
2. Wait for health checks to pass
3. Do NOT route traffic yet

**Commands executed:**
```bash
# Start container 1
ssh root@192.168.1.10 'docker run -d \
  --name myapp-web-a1b2c3d-1 \
  --label podlift.version=a1b2c3d \
  --label podlift.service=web \
  -p 8001:8000 \
  -e SECRET_KEY=xxx \
  myapp:a1b2c3d'

# Wait for health check
ssh root@192.168.1.10 'for i in {1..30}; do \
  curl -f http://localhost:8001/health && break; \
  sleep 1; \
done'
```

**Port allocation:**
- New containers use temporary ports (8001, 8002, etc.)
- Old containers still on primary ports
- nginx still routes to old containers

If health checks fail, deployment stops. Old containers keep running.

### Step 6: Update nginx

```bash
[6/7] Updating nginx configuration...
```

**What happens:**
1. Generate new nginx config
2. Update upstream to point to new containers
3. Test config: `nginx -t`
4. Reload nginx (zero downtime)
5. Traffic now flows to new containers

**nginx config generated:**
```nginx
upstream myapp_web {
    server 127.0.0.1:8001;  # New container 1
    server 127.0.0.1:8002;  # New container 2
}

server {
    listen 80;
    server_name myapp.com;
    
    location / {
        proxy_pass http://myapp_web;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Commands executed:**
```bash
# Upload new config
scp /tmp/nginx.conf root@192.168.1.10:/etc/nginx/sites-available/myapp

# Test and reload
ssh root@192.168.1.10 'nginx -t && systemctl reload nginx'
```

**Why this is zero-downtime:**
- `nginx -t` validates config before applying
- `systemctl reload` reloads without dropping connections
- Old containers still running during reload
- New requests go to new containers
- In-flight requests to old containers complete

### Step 7: Stop Old Containers

```bash
[7/7] Stopping old containers...
```

**What happens:**
1. Wait 30 seconds for connection draining
2. Stop old containers
3. Keep old images (for rollback)
4. Cleanup temporary files

**Commands executed:**
```bash
# Wait for draining
sleep 30

# Stop old containers
ssh root@192.168.1.10 'docker stop myapp-web-x9y8z7w-1'
ssh root@192.168.1.10 'docker stop myapp-web-x9y8z7w-2'

# Note: docker images NOT removed (kept for rollback)
```

Old containers are stopped but not removed. Images remain on disk.

## Rollback Process

How `podlift rollback` works.

### Finding Previous Version

```bash
# Query Docker labels
ssh root@192.168.1.10 'docker ps -a \
  --filter "label=podlift.service=web" \
  --format "{{.Labels}}"'
```

Containers have labels:
```
podlift.version=a1b2c3d
podlift.service=web
podlift.deployed_at=2025-11-05T10:30:00Z
```

podlift finds the most recent stopped container and uses its version.

### Starting Old Containers

```bash
# Start old containers
ssh root@192.168.1.10 'docker start myapp-web-x9y8z7w-1'
ssh root@192.168.1.10 'docker start myapp-web-x9y8z7w-2'
```

If old containers are removed, podlift uses the old image (still on disk):
```bash
ssh root@192.168.1.10 'docker run -d \
  --name myapp-web-x9y8z7w-1 \
  myapp:x9y8z7w'
```

### Update nginx

Same as deploy: update upstream, test, reload.

### Stop Current Containers

After old version is healthy and serving traffic, stop current containers.

## State Management

podlift tracks deployments in two places:

### 1. Docker Labels

Every container has labels:
```json
{
  "podlift.version": "a1b2c3d",
  "podlift.service": "web",
  "podlift.deployed_at": "2025-11-05T10:30:00Z",
  "podlift.deployed_by": "user@hostname",
  "podlift.git_branch": "main",
  "podlift.git_tag": "v1.2.3"
}
```

Query with: `docker ps --filter "label=podlift.version=a1b2c3d"`

### 2. State File on Server

`/opt/myapp/.podlift/state.json`:
```json
{
  "current": {
    "version": "a1b2c3d",
    "deployed_at": "2025-11-05T10:30:00Z",
    "containers": ["myapp-web-a1b2c3d-1", "myapp-web-a1b2c3d-2"]
  },
  "previous": {
    "version": "x9y8z7w",
    "deployed_at": "2025-11-04T15:20:00Z",
    "containers": ["myapp-web-x9y8z7w-1", "myapp-web-x9y8z7w-2"]
  },
  "history": [...]
}
```

If state file is lost, podlift rebuilds it from Docker labels.

## Multi-Server Deployment

With multiple servers, podlift:

1. **Deploys to all servers** (serial by default, `--parallel` for simultaneous)
2. **Automatically sets up nginx load balancing** on the primary server
3. **Configures upstreams** to distribute traffic across all servers

### Automatic Load Balancing

When you deploy to 2+ servers, podlift automatically:

- Installs nginx on the primary server (labeled `primary` or first server)
- Configures upstream backends pointing to all servers
- Uses `least_conn` algorithm for connection distribution
- Sets up health checks (`max_fails=3`, `fail_timeout=30s`)
- Maintains persistent connections (`keepalive 32`)

No additional configuration required - it just works.

```yaml
servers:
  web:
    - host: 192.168.1.10
    - host: 192.168.1.11
```

### Serial Deployment (Default)

```
Deploy to 192.168.1.10:
  [1/7] Validate
  [2/7] Build (once, locally)
  [3/7] Push to .10
  [4/7] Load on .10
  [5/7] Start containers on .10
  [6/7] Update nginx on .10
  [7/7] Stop old containers on .10

Deploy to 192.168.1.11:
  [3/7] Push to .11
  [4/7] Load on .11
  [5/7] Start containers on .11
  [6/7] Update nginx on .11
  [7/7] Stop old containers on .11
```

**Why serial?**
- If first server fails, deployment stops
- No wasted work on remaining servers
- Clear error reporting

### Parallel Deployment (--parallel)

```bash
podlift deploy --parallel
```

All servers deploy simultaneously. Faster, but if one fails, others continue.

## Dependency Management

Dependencies run once on the "primary" server.

```yaml
servers:
  web:
    - host: 192.168.1.10
      labels: [primary]
    - host: 192.168.1.11

dependencies:
  postgres:
    image: postgres:16
```

### What Happens

On first deploy:
```bash
# Check if postgres exists on .10
ssh root@192.168.1.10 'docker ps | grep postgres'

# If not, start it
ssh root@192.168.1.10 'docker run -d \
  --name myapp-postgres \
  --label podlift.dependency=postgres \
  -v postgres_data:/var/lib/postgresql/data \
  postgres:16'
```

On subsequent deploys:
- Dependencies are not restarted
- Data persists across deployments

### Network Configuration

podlift creates a Docker network for service communication:

```bash
ssh root@192.168.1.10 'docker network create myapp_network'

# Attach all containers
docker run --network myapp_network --name myapp-postgres postgres:16
docker run --network myapp_network --name myapp-web-a1b2c3d-1 myapp:a1b2c3d
```

Containers can reach dependencies by name:
- `postgres` resolves to postgres container
- `redis` resolves to redis container

For multi-server setups, dependencies are reachable via primary server IP.

## SSL with Let's Encrypt

How `podlift ssl setup` works.

### Step 1: Install Certbot

```bash
ssh root@192.168.1.10 'apt-get install -y certbot python3-certbot-nginx'
```

### Step 2: Obtain Certificate

```bash
ssh root@192.168.1.10 'certbot --nginx \
  -d myapp.com \
  --non-interactive \
  --agree-tos \
  --email admin@myapp.com'
```

Certbot automatically:
- Updates nginx config
- Adds SSL directives
- Configures HTTP → HTTPS redirect

### Step 3: Auto-Renewal

```bash
ssh root@192.168.1.10 'systemctl enable certbot.timer'
```

Cron job renews certificate automatically before expiration.

## What podlift Does NOT Do

**No custom infrastructure:**
- No custom proxy (uses nginx)
- No custom orchestration (uses Docker)
- No agents running on servers
- No persistent connections

**No state on your machine:**
- All state lives on servers
- You can deploy from any machine
- No local database

**No lock-in:**
- Everything is standard Docker commands
- You can manage manually if podlift fails
- nginx config is editable
- Full access to all components

## Debugging

Everything is transparent. You can inspect and modify:

### Check What's Running

```bash
ssh root@192.168.1.10 'docker ps'
```

### View nginx Config

```bash
ssh root@192.168.1.10 'cat /etc/nginx/sites-available/myapp'
```

### Check nginx Logs

```bash
ssh root@192.168.1.10 'tail -f /var/log/nginx/access.log'
```

### Manually Fix Issues

If deployment fails, you can:
- Manually start containers
- Edit nginx config
- Debug with `docker exec`
- Restart nginx

podlift won't fight you. It's a helper, not a dictator.

## Security

### SSH Keys

All communication uses SSH key authentication. No passwords.

### Secrets

Environment variables never logged or transmitted insecurely:
```bash
# Secrets are passed via SSH
ssh root@192.168.1.10 'docker run -e SECRET_KEY="$SECRET" myapp'
```

### Least Privilege

podlift needs:
- SSH access to servers (root or docker group)
- Docker permissions

Nothing more.

## Performance

Typical deployment timeline:

- Validation: 2-5 seconds
- Build: 30-60 seconds (cached: 5-10 seconds)
- Transfer (SCP): 10-30 seconds for 200MB image
- Load: 5-10 seconds
- Start: 5-15 seconds (including health checks)
- nginx reload: <1 second
- Total: **1-2 minutes**

Registry method is faster for multi-server (push once, pull many).

## Comparison to Kamal

| Aspect | podlift | Kamal |
|--------|---------|-------|
| Proxy | nginx | kamal-proxy (custom) |
| Image transfer | SCP or registry | Registry only |
| State | Docker labels + JSON | Unknown |
| Healthcheck | HTTP + Docker | kamal-proxy |
| Validation | Pre-flight checks | Fails late |
| Transparency | Show all commands | Black box |
| Debugging | Standard tools | Custom tooling |

podlift's philosophy: Use proven tools. Be transparent. Fail fast.

