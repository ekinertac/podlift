# Test Configuration Examples

Example `podlift.yml` configurations for different use cases.

## Files

### minimal.yml
**Use case:** Simplest possible deployment (quick prototypes, testing)

Features:
- Single server
- No registry (uses SCP)
- No dependencies
- All defaults

**When to use:**
- Testing podlift for the first time
- Deploying static files (nginx)
- Simple apps with no database

---

### standard.yml ⭐ (Recommended)
**Use case:** Typical production deployment (most common scenario)

Features:
- Single production server
- PostgreSQL database
- Docker registry (GitHub Container Registry)
- SSL with Let's Encrypt
- Health checks
- Database migrations hook
- Environment variables from .env

**When to use:**
- Django/Flask/Rails apps
- Apps with database
- Production deployments
- 90% of real-world use cases

**What it covers:**
- Database setup (postgres)
- Secrets management (env vars)
- SSL/HTTPS
- Zero-downtime (replicas: 2)
- Post-deploy tasks (migrations)

This is what you'd actually use in production.

---

### full.yml
**Use case:** Complex deployment with all features (reference)

Features:
- Multiple servers (web + worker roles)
- Multiple dependencies (postgres + redis)
- Worker processes (Celery)
- All configuration options shown
- Advanced features

**When to use:**
- Learning all available options
- Complex applications
- Reference for documentation
- Advanced deployments

**What it covers:**
- Everything podlift can do
- All configuration fields
- Comments explaining options

---

## Comparison

| Feature | minimal.yml | standard.yml | full.yml |
|---------|-------------|--------------|----------|
| Servers | 1 | 1 | 3 |
| Roles | Default (web) | Default (web) | web + worker |
| Dependencies | None | postgres | postgres + redis |
| Registry | No (SCP) | Yes (ghcr.io) | Yes (ghcr.io) |
| SSL | No | Yes (Let's Encrypt) | Yes (Let's Encrypt) |
| Replicas | 1 | 2 | 2 |
| Health checks | Default | Yes | Yes |
| Hooks | No | Yes (migrations) | Yes (migrations + static) |
| Env vars | No | Yes | Yes |
| Workers | No | No | Yes |

## Quick Start Guide

### For First-Time Users
Start with **minimal.yml**:
```bash
cp testdata/minimal.yml podlift.yml
# Edit: change server IP
podlift deploy
```

### For Production Apps
Use **standard.yml**:
```bash
cp testdata/standard.yml podlift.yml
# Edit: 
#   - Change domain
#   - Change server IP
#   - Set up .env file
podlift deploy
```

### For Complex Apps
Start with **standard.yml**, then consult **full.yml** for additional features.

## Environment Variables

### For standard.yml

Create `.env` file:
```bash
# .env
REGISTRY_USER=your-github-username
REGISTRY_PASSWORD=ghp_your_github_token
DB_PASSWORD=secure-random-password
SECRET_KEY=your-app-secret-key
```

### For full.yml

Additional variables:
```bash
# .env (add to above)
REGISTRY_USER=your-github-username
REGISTRY_PASSWORD=ghp_your_github_token
DB_PASSWORD=secure-random-password
SECRET_KEY=your-app-secret-key
```

## Testing Configurations

All three configs are tested:

```bash
# Test loading
go test ./internal/config/... -v

# Load specific config
podlift validate --config testdata/standard.yml
```

## Common Modifications

### Add Redis to standard.yml

```yaml
dependencies:
  postgres:
    image: postgres:16
    # ... existing config ...
  
  redis:
    image: redis:7-alpine
    port: 6379

services:
  web:
    env:
      # ... existing vars ...
      REDIS_URL: redis://primary:6379
```

### Add Multiple Servers to standard.yml

```yaml
servers:
  - host: 192.168.1.10
    user: root
    labels: [primary]
  - host: 192.168.1.11
    user: root
```

### Remove SSL from standard.yml

```yaml
proxy:
  enabled: true
  ssl: false  # Change this
```

Or remove the proxy section entirely.

---

**Recommendation:** Start with `standard.yml` for any real deployment. It covers 90% of use cases without overwhelming complexity.

