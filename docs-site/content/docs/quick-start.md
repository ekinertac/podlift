---
title: Quick Start
weight: 5
---
# Quick Start Guide

Get started with podlift in 5 minutes. This guide shows common workflows without reading all documentation.

## First Time Setup

### 1. Install podlift

```bash
curl -sSL https://raw.githubusercontent.com/ekinertac/podlift/main/install.sh | sh
```

### 2. Initialize Your Project

```bash
cd myapp/
podlift init
```

Creates `podlift.yml` and `.env.example`.

### 3. Configure Server

Edit `podlift.yml`:

```yaml
service:
  name: myapp
  image: myapp
  port: 8000

servers:
  - host: 192.168.1.10    # Your server IP
    user: root            # SSH user
```

### 4. Setup Server

First time only - installs Docker, configures firewall:

```bash
podlift setup
```

### 5. Deploy

```bash
podlift deploy
```

Done! Your app is live at `http://192.168.1.10`

---

## Common Workflows

### Daily Development

```bash
# 1. Make changes
vim app/main.py

# 2. Test locally
docker-compose up

# 3. Commit
git add -A
git commit -m "Add feature"

# 4. Deploy
podlift deploy
```

### Multiple Environments

**Project structure:**
```
myapp/
├── docker-compose.yml       # Local: docker-compose up
├── podlift.yml              # Staging deployment
├── podlift.production.yml   # Production deployment
├── .env
└── .env.production
```

**Deploy to staging:**
```bash
podlift deploy
```

**Deploy to production:**
```bash
podlift deploy --config podlift.production.yml
```

### Add SSL/HTTPS

```yaml
# podlift.yml
service:
  name: myapp
  domain: myapp.com    # Add domain

servers:
  - host: 1.2.3.4

ssl:
  enabled: true
  email: admin@myapp.com
```

```bash
podlift ssl setup
```

Your site is now at `https://myapp.com`

### Add Database

```yaml
# podlift.yml
dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

services:
  web:
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
```

```bash
# .env
DB_PASSWORD=your-secure-password

podlift deploy
```

Database is automatically deployed and connected.

### Multiple Servers (Load Balancing)

```yaml
# podlift.yml
servers:
  - host: 1.2.3.4
    labels: [primary]
  - host: 5.6.7.8
```

```bash
podlift deploy
```

Automatically sets up nginx load balancer across both servers.

### Registry Instead of SCP

```yaml
# podlift.yml
registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}
```

```bash
# .env
REGISTRY_USER=myuser
REGISTRY_PASSWORD=ghp_xxxxx

podlift deploy
```

Pushes to registry instead of uploading via SCP.

### Rollback Deployment

```bash
podlift rollback
```

Instantly reverts to previous version.

---

## Common Patterns

### Pattern 1: Simple Web App

**Use case:** Single server, no database, basic deployment

```yaml
service:
  name: myapp
  image: myapp
  port: 8000

servers:
  - host: 1.2.3.4
    user: deploy
```

**Deploy:**
```bash
podlift deploy
```

### Pattern 2: Web App + Database

**Use case:** App with PostgreSQL, single server

```yaml
service:
  name: myapp
  image: myapp
  port: 8000

servers:
  - host: 1.2.3.4
    user: deploy

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

services:
  web:
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
```

### Pattern 3: Production with SSL + Load Balancing

**Use case:** Multiple servers, SSL, high availability

```yaml
service:
  name: myapp
  domain: myapp.com
  image: myapp
  port: 8000

servers:
  - host: 1.2.3.4
    user: deploy
    labels: [primary]
  - host: 5.6.7.8
    user: deploy

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

services:
  web:
    replicas: 2
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379

ssl:
  enabled: true
  email: admin@myapp.com

hooks:
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
```

### Pattern 4: Worker Servers

**Use case:** Web servers separate from background workers

```yaml
service:
  name: myapp
  image: myapp

servers:
  web:
    - host: 1.2.3.4
    - host: 5.6.7.8
  
  worker:
    - host: 9.10.11.12
    - host: 13.14.15.16

services:
  web:
    port: 8000
  
  worker:
    command: celery -A myapp worker
    replicas: 2
```

---

## Environment-Specific Configs

### Staging vs Production

**podlift.yml (staging):**
```yaml
service:
  name: myapp-staging
  domain: staging.myapp.com
  image: myapp

env_file: .env.staging

servers:
  - host: staging.myapp.com

services:
  web:
    replicas: 1
    env:
      DEBUG: "true"

ssl:
  enabled: true
```

**podlift.production.yml:**
```yaml
service:
  name: myapp
  domain: myapp.com
  image: myapp

env_file: .env.production

servers:
  - host: prod1.myapp.com
    labels: [primary]
  - host: prod2.myapp.com

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data

services:
  web:
    replicas: 3
    env:
      DEBUG: "false"

ssl:
  enabled: true
```

**Deploy:**
```bash
podlift deploy                              # Staging
podlift deploy --config podlift.production.yml  # Production
```

---

## Secrets Management

### Local .env Files

```bash
# .env (for staging)
DB_PASSWORD=staging_password
SECRET_KEY=staging_secret

# .env.production (for production)
DB_PASSWORD=production_password
SECRET_KEY=production_secret
REGISTRY_PASSWORD=ghp_xxxxx
```

**In podlift.yml:**
```yaml
env_file: .env.production  # Point to specific env file

services:
  web:
    env:
      SECRET_KEY: ${SECRET_KEY}
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
```

### CI/CD Secrets

**GitHub Actions:**
```yaml
- name: Deploy
  env:
    DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
    SECRET_KEY: ${{ secrets.SECRET_KEY }}
  run: podlift deploy --config podlift.production.yml
```

**GitLab CI:**
```yaml
deploy:
  script:
    - export DB_PASSWORD=$DB_PASSWORD
    - export SECRET_KEY=$SECRET_KEY
    - podlift deploy --config podlift.production.yml
```

---

## Monitoring & Maintenance

### Check Status

```bash
podlift ps              # Running services
podlift status          # Detailed status
```

### View Logs

```bash
podlift logs web               # Last 100 lines
podlift logs web --follow      # Stream logs
podlift logs web --tail 500    # Last 500 lines
```

### Execute Commands

```bash
podlift exec web bash                      # Interactive shell
podlift exec web python manage.py migrate  # Run migration
```

### Health Checks

```bash
podlift validate        # Validate config and servers
```

---

## Makefile Shortcuts

Add to `Makefile`:

```makefile
deploy-staging:
	podlift deploy

deploy-production:
	podlift deploy --config podlift.production.yml

rollback-staging:
	podlift rollback

rollback-production:
	podlift rollback --config podlift.production.yml

logs:
	podlift logs web --follow

status:
	podlift status

shell:
	podlift exec web bash
```

**Usage:**
```bash
make deploy-staging
make deploy-production
make logs
```

---

## Troubleshooting

### Deployment Failed

```bash
# Check logs
podlift logs web

# Validate config
podlift validate

# Rollback
podlift rollback
```

### Health Check Failing

```yaml
# Adjust health check in podlift.yml
services:
  web:
    health_check:
      path: /health
      expect: [200, 301]
      timeout: 60s      # Increase timeout
```

### Port Already in Use

```bash
# Check what's running
podlift ps --all

# Clean up old containers if needed
ssh root@server 'docker ps -a'
```

---

## Next Steps

- [Full Configuration Reference](configuration.md)
- [All Commands](commands.md)
- [Deployment Strategies](deployment-guide.md)
- [Troubleshooting](troubleshooting.md)
- [How It Works](how-it-works.md)

## Getting Help

- [GitHub Issues](https://github.com/ekinertac/podlift/issues)
- [Documentation](https://ekinertac.github.io/podlift/)

