# Deployment Guide

Complete guide to deploying applications with podlift.

## Basic Deployment Flow

Standard deployment workflow:

```bash
# 1. Make changes
vim app/views.py

# 2. Test locally
docker build -t myapp .
docker run -p 8000:8000 myapp

# 3. Commit changes
git add -A
git commit -m "Fix bug in views"

# 4. Deploy
podlift deploy
```

That's it. podlift handles the rest.

## Zero-Downtime Deployment

How podlift ensures no requests are dropped.

### The Process

```
Current state: web-v1 running, serving traffic

1. Start web-v2 containers (old still running)
2. Wait for web-v2 health checks
3. Update nginx to route to web-v2
4. Wait 30s for connection draining
5. Stop web-v1 containers

New state: web-v2 running, serving traffic
```

### During Deployment

```bash
$ podlift deploy
[5/7] Starting new containers...
  Starting myapp-web-abc123-1...
    Health check passed (5s)
  Starting myapp-web-abc123-2...
    Health check passed (4s)

[6/7] Updating nginx configuration...
  Traffic switching to new version...
  ✓ 100% of traffic on new version

[7/7] Connection draining...
  Waiting 30s for in-flight requests to complete...
  ✓ Old containers stopped
```

### What nginx Does

Before:
```nginx
upstream myapp_web {
    server 127.0.0.1:8001;  # web-v1
    server 127.0.0.1:8002;  # web-v1
}
```

After:
```nginx
upstream myapp_web {
    server 127.0.0.1:9001;  # web-v2
    server 127.0.0.1:9002;  # web-v2
}
```

nginx reloads gracefully. In-flight requests to v1 complete. New requests go to v2.

### Failure Handling

If new version fails health checks:

```bash
[5/7] Starting new containers...
  Starting myapp-web-abc123-1...
    ✗ Health check failed (timeout 30s)

ERROR: Deployment failed

Container started but failed health checks.
Old containers still running. No downtime occurred.

Check logs: podlift logs web
```

**Old version keeps running.** No traffic switches. No downtime.

## Rollback

Reverting to previous version.

### Quick Rollback

```bash
podlift rollback
```

Reverts to the last successful deployment.

### Rollback to Specific Version

```bash
# By git tag
podlift rollback --to v1.2.3

# By git commit
podlift rollback --to a1b2c3d

# By looking at deployment history
podlift ps --all
podlift rollback --to x9y8z7w
```

### How It Works

```bash
$ podlift rollback

Finding previous deployment...
  Found: x9y8z7w "Working version" (deployed 2h ago)

[1/4] Starting old containers...
  ✓ myapp-web-x9y8z7w-1 started
  ✓ Health check passed (3s)

[2/4] Updating nginx...
  ✓ Traffic routing to x9y8z7w

[3/4] Stopping current containers...
  ✓ myapp-web-abc123-1 stopped

[4/4] Cleanup...
  ✓ Complete

✓ Rollback successful!
Time: 34s
```

Same zero-downtime process, just in reverse.

### Rollback Failed Deployment

If deployment fails, you don't need to rollback—old version is still running:

```bash
$ podlift deploy
# Deployment fails

$ podlift ps
SERVICE  VERSION  STATUS
web      x9y8z7w  healthy  ← Still running

# Fix the issue, redeploy
$ podlift deploy
```

## Multi-Server Deployment

Deploying to multiple servers.

### Serial Deployment (Default)

```yaml
servers:
  web:
    - host: 192.168.1.10
    - host: 192.168.1.11
    - host: 192.168.1.12
```

```bash
$ podlift deploy

Deploying to 192.168.1.10...
  [1/7] Validate...
  [2/7] Build (local)...
  [3/7] Transfer...
  [4/7] Load...
  [5/7] Start containers...
  [6/7] Update nginx...
  [7/7] Cleanup...
  ✓ Complete

Deploying to 192.168.1.11...
  [3/7] Transfer...
  [4/7] Load...
  [5/7] Start containers...
  [6/7] Update nginx...
  [7/7] Cleanup...
  ✓ Complete

Deploying to 192.168.1.12...
  ✓ Complete

✓ All servers deployed successfully!
```

**Benefits:**
- If first server fails, stop before touching others
- Clear progress tracking
- Predictable order

**When to use:** Production deployments where safety matters.

### Parallel Deployment

```bash
podlift deploy --parallel
```

All servers deploy simultaneously:

```
Deploying to all servers in parallel...

192.168.1.10: [=====>    ] 50%
192.168.1.11: [=======>  ] 70%
192.168.1.12: [=========>] 90%
```

**Benefits:**
- Faster (3x faster with 3 servers)

**Drawbacks:**
- If one fails, others continue (potential inconsistency)
- Less clear error reporting

**When to use:** Staging environments or when speed matters more than safety.

## Load Balancing

Multiple servers serving the same app.

### Configuration

```yaml
servers:
  web:
    - host: 192.168.1.10
    - host: 192.168.1.11

services:
  web:
    replicas: 2  # 2 containers per server = 4 total
```

### nginx Configuration

podlift generates:

```nginx
upstream myapp_web {
    # Server 1
    server 192.168.1.10:8001;
    server 192.168.1.10:8002;
    
    # Server 2
    server 192.168.1.11:8001;
    server 192.168.1.11:8002;
}

server {
    listen 80;
    server_name myapp.com;
    
    location / {
        proxy_pass http://myapp_web;
    }
}
```

nginx distributes requests across all containers on all servers.

### Adding Load Balancer

For production, put a load balancer in front:

```
                 ┌─> Server 1 (192.168.1.10)
Client → LB ─────┼─> Server 2 (192.168.1.11)
                 └─> Server 3 (192.168.1.12)
```

Use:
- DigitalOcean Load Balancer
- AWS ALB
- Cloudflare
- HAProxy

Point LB to all server IPs on port 80/443.

## Worker Servers

Separate servers for background jobs.

### Configuration

```yaml
servers:
  web:
    - host: 192.168.1.10
      labels: [primary]  # Dependencies run here
    - host: 192.168.1.11
  
  worker:
    - host: 192.168.1.20
    - host: 192.168.1.21

dependencies:
  postgres:
    image: postgres:16
  redis:
    image: redis:7

services:
  web:
    port: 8000
    healthcheck:
      path: /health
  
  worker:
    command: celery -A myapp worker
    replicas: 2
    healthcheck: false  # Workers don't have HTTP endpoints
```

### What Happens

```
Server 192.168.1.10 (primary):
  - postgres
  - redis
  - web (2 containers)

Server 192.168.1.11:
  - web (2 containers)

Server 192.168.1.20:
  - worker (2 containers)

Server 192.168.1.21:
  - worker (2 containers)
```

Workers connect to postgres/redis on primary server.

### Environment Variables

```yaml
services:
  web:
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379
  
  worker:
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@192.168.1.10:5432/myapp
      REDIS_URL: redis://192.168.1.10:6379
```

`primary` resolves to the server with `labels: [primary]`.

## Deployment Strategies

### Strategy 1: Simple Single Server

Best for: Side projects, MVPs, small apps

```yaml
service: myapp
servers:
  - host: 192.168.1.10
```

**Pros:**
- Simple
- Cheap ($5-10/month)
- Easy to debug

**Cons:**
- Single point of failure
- Limited scale

### Strategy 2: Multi-Server Web

Best for: Growing apps with traffic spikes

```yaml
servers:
  web:
    - host: 192.168.1.10
      labels: [primary]
    - host: 192.168.1.11
    - host: 192.168.1.12
```

**Pros:**
- Handles more traffic
- Redundancy
- Can scale by adding servers

**Cons:**
- More expensive
- Dependencies still single point of failure

### Strategy 3: Separate Web + Workers

Best for: Apps with background jobs

```yaml
servers:
  web:
    - host: 192.168.1.10
      labels: [primary]
    - host: 192.168.1.11
  
  worker:
    - host: 192.168.1.20
    - host: 192.168.1.21
```

**Pros:**
- Workers don't affect web performance
- Can scale web and workers independently

**Cons:**
- More servers = more cost
- More complex

### Strategy 4: Separate Database Server

Best for: Apps with heavy database load

Not directly supported by podlift (use managed database):

```yaml
services:
  web:
    env:
      DATABASE_URL: postgres://user:pass@db-server.example.com:5432/myapp
```

Use DigitalOcean Managed Database, AWS RDS, etc.

## Deployment Hooks

Run commands after deployment.

### Configuration

```yaml
hooks:
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
    - docker exec myapp-web-1 python manage.py collectstatic --noinput
    - docker exec myapp-web-1 python manage.py clearsessions
```

### When Hooks Run

```bash
$ podlift deploy

[1/7] Validate...
[2/7] Build...
[3/7] Transfer...
[4/7] Load...
[5/7] Start containers...
[6/7] Update nginx...
[7/7] Cleanup...

Running post-deploy hooks...
  ✓ python manage.py migrate (2s)
  ✓ python manage.py collectstatic (3s)
  ✓ python manage.py clearsessions (1s)

✓ Deployment successful!
```

### Hook Types

```yaml
hooks:
  before_deploy:
    - echo "Deployment starting"
  
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
  
  after_rollback:
    - echo "Rolled back to previous version"
```

Hooks run on the primary server via SSH.

## Environment-Specific Deploys

Different configs for staging vs production.

### Option 1: Separate Config Files

```
podlift.staging.yml
podlift.production.yml
```

Deploy:
```bash
# Staging
podlift deploy --config podlift.staging.yml

# Production
podlift deploy --config podlift.production.yml
```

### Option 2: Environment Variables

```yaml
# podlift.yml
servers:
  - host: ${SERVER_HOST}

services:
  web:
    env:
      ENVIRONMENT: ${ENVIRONMENT}
      DEBUG: ${DEBUG}
```

Deploy:
```bash
# Staging
ENVIRONMENT=staging SERVER_HOST=staging.server.com podlift deploy

# Production
ENVIRONMENT=production SERVER_HOST=prod.server.com podlift deploy
```

### Option 3: Git Branches

```bash
# Staging (deploy from staging branch)
git checkout staging
podlift deploy --config podlift.yml

# Production (deploy from main branch)
git checkout main
podlift deploy --config podlift.yml
```

Use same config, different branches = different code versions.

## Database Migrations

Handling schema changes.

### Strategy 1: Post-Deploy Hook (Recommended)

```yaml
hooks:
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
```

**Flow:**
1. Deploy new code
2. New containers start
3. Run migrations
4. Traffic switches to new version

**Safe for:**
- Adding columns
- Adding tables
- Adding indexes (with CONCURRENT)

**Unsafe for:**
- Removing columns (old code still running)
- Renaming columns

### Strategy 2: Manual Migrations

```bash
# Deploy without traffic switch
podlift deploy --skip-healthcheck

# Run migrations
podlift exec web python manage.py migrate

# Manually test
curl http://server-ip:port/health

# If good, update nginx manually
ssh root@server 'systemctl reload nginx'
```

### Strategy 3: Two-Phase Deploy

For breaking schema changes:

**Phase 1: Make column optional**
```python
# Migration: make column nullable
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
```

Deploy and run migration.

**Phase 2: Remove column**
```python
# Migration: remove column
ALTER TABLE users DROP COLUMN email;
```

Deploy and run migration.

### Best Practices

1. **Make migrations backward compatible**
2. **Test migrations on staging first**
3. **Backup database before risky migrations**
4. **Use `--skip-healthcheck` for manual control**

## Monitoring Deployments

Track deployment success.

### During Deployment

```bash
# Watch logs during deploy
podlift deploy --verbose

# In another terminal
podlift logs web --follow
```

### After Deployment

```bash
# Check status
podlift ps

# View logs
podlift logs web --tail 100

# Test endpoint
curl https://myapp.com/health
```

### Automated Monitoring

Use external monitoring:

```yaml
# After deployment, ping healthcheck
hooks:
  after_deploy:
    - curl -fsS https://hc-ping.com/your-uuid
```

Services:
- Healthchecks.io
- UptimeRobot
- Pingdom

Get alerted if deployments fail.

## Troubleshooting Deployments

### Deployment Fails at Health Check

```bash
# Check logs
podlift logs web

# Common issues:
# - Missing env vars
# - Database not ready
# - Wrong healthcheck path

# Debug interactively
podlift exec web bash
curl http://localhost:8000/health
```

### Deployment Succeeds but App Broken

```bash
# Rollback immediately
podlift rollback

# Debug locally
git checkout <deployed-commit>
docker build -t myapp .
docker run -p 8000:8000 myapp
```

### Slow Deployments

```bash
# Skip build if image unchanged
podlift deploy --skip-build

# Use parallel for multi-server
podlift deploy --parallel

# Use registry instead of SCP
# (faster for multiple servers)
```

## Best Practices

### 1. Always Commit Before Deploy

```bash
git status  # Check for uncommitted changes
git add -A
git commit -m "Descriptive message"
podlift deploy
```

### 2. Test Locally First

```bash
docker build -t myapp .
docker run -p 8000:8000 myapp
curl http://localhost:8000/health
```

### 3. Deploy to Staging First

```bash
# Staging
podlift deploy --config podlift.staging.yml

# Test staging
curl https://staging.myapp.com

# Production
podlift deploy --config podlift.production.yml
```

### 4. Use Health Checks

```yaml
services:
  web:
    healthcheck:
      path: /health
      expect: [200]
      timeout: 30s
```

Don't deploy without health checks.

### 5. Monitor After Deploy

```bash
# Deploy
podlift deploy

# Watch for errors
podlift logs web --follow

# Check metrics
# (CPU, memory, error rate)
```

Give it 5-10 minutes before considering it stable.

### 6. Keep Rollback Ready

```bash
# If anything looks wrong
podlift rollback
```

Don't hesitate to rollback. Debug offline.

### 7. Backup Before Risky Changes

```bash
# Before schema migrations
ssh root@server
docker exec postgres pg_dump myapp > backup.sql

# Then deploy
podlift deploy
```

## Next Steps

- [Commands Reference](commands.md) - All available commands
- [Configuration Reference](configuration.md) - All config options
- [Troubleshooting](troubleshooting.md) - Common issues
- [How It Works](how-it-works.md) - Architecture details

