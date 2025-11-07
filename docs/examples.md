# Configuration Examples

Real-world configuration examples for common deployment scenarios.

## Single Server - Basic Setup

**Scenario:** Deploy a simple web app to one server.

**When to use:** Small projects, MVPs, hobby projects.

```yaml
service: myapp
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy

services:
  web:
    port: 8000
```

**Deploy:**
```bash
podlift deploy
```

## Single Server with SSL

**Scenario:** Production-ready single server with HTTPS.

**When to use:** Small business sites, blogs, documentation sites.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy

services:
  web:
    port: 8000
    env:
      SECRET_KEY: ${SECRET_KEY}

ssl:
  enabled: true
  email: admin@myapp.com
```

## Multi-Server Load Balanced

**Scenario:** High-availability app across multiple servers.

**When to use:** Production apps with traffic, need for redundancy.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy
    labels: [primary]
  - host: 192.168.1.11
    user: deploy
  - host: 192.168.1.12
    user: deploy

services:
  web:
    port: 8000
    replicas: 2  # 2 containers per server = 6 total

ssl:
  enabled: true
  email: devops@myapp.com
```

**Result:** Automatic nginx load balancer across all servers.

## With Database (PostgreSQL)

**Scenario:** Web app with PostgreSQL database.

**When to use:** Most production applications.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy
    labels: [primary]

dependencies:
  postgres:
    image: postgres:16
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_DB: myapp
      POSTGRES_PASSWORD: ${DB_PASSWORD}

services:
  web:
    port: 8000
    replicas: 2
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      SECRET_KEY: ${SECRET_KEY}

ssl:
  enabled: true
  email: admin@myapp.com
```

## With Redis Cache

**Scenario:** App with PostgreSQL and Redis.

**When to use:** Apps needing session storage, caching, queues.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy
    labels: [primary]

dependencies:
  postgres:
    image: postgres:16
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
  
  redis:
    image: redis:7-alpine
    port: 6379
    volume: redis_data:/data

services:
  web:
    port: 8000
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379

ssl:
  enabled: true
  email: admin@myapp.com
```

## Web + Background Workers

**Scenario:** Separate web and worker servers.

**When to use:** Apps with background jobs (Celery, Sidekiq, etc).

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  web:
    - host: 192.168.1.10
      user: deploy
      labels: [primary]
  
  worker:
    - host: 192.168.1.20
      user: deploy
    - host: 192.168.1.21
      user: deploy

dependencies:
  postgres:
    image: postgres:16
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
  
  redis:
    image: redis:7-alpine
    port: 6379

services:
  web:
    port: 8000
    replicas: 2
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379
  
  worker:
    command: celery -A myapp worker --loglevel=info
    replicas: 2
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379

ssl:
  enabled: true
  email: admin@myapp.com
```

## Using Docker Registry

**Scenario:** Use GitHub Container Registry instead of SCP.

**When to use:** CI/CD pipelines, team workflows, private registries.

```yaml
service: myapp
domain: myapp.com
image: myapp

registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_TOKEN}

servers:
  - host: 192.168.1.10
    user: deploy

services:
  web:
    port: 8000
    env:
      SECRET_KEY: ${SECRET_KEY}

ssl:
  enabled: true
  email: admin@myapp.com
```

**CI/CD workflow:**
```bash
# Build and push in CI
docker build -t ghcr.io/username/myapp:$GIT_SHA .
docker push ghcr.io/username/myapp:$GIT_SHA

# Deploy from registry
podlift deploy
```

## Staging vs Production

**Scenario:** Separate staging and production environments.

**When to use:** Professional development workflows.

### File Structure

```
myapp/
├── docker-compose.yml
├── podlift.yml              # Staging
├── podlift.production.yml   # Production
├── .env.staging
└── .env.production
```

### podlift.yml (Staging)

```yaml
service: myapp-staging
domain: staging.myapp.com
image: myapp

env_file: .env.staging

servers:
  - host: staging.myapp.com
    user: deploy

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

services:
  web:
    port: 8000
    replicas: 1
    env:
      ENVIRONMENT: staging
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp

ssl:
  enabled: true
  email: devops@myapp.com
```

### podlift.production.yml

```yaml
service: myapp
domain: myapp.com
image: myapp

env_file: .env.production

servers:
  - host: prod1.myapp.com
    user: deploy
    labels: [primary]
  - host: prod2.myapp.com
    user: deploy
  - host: prod3.myapp.com
    user: deploy

dependencies:
  postgres:
    image: postgres:16
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
  
  redis:
    image: redis:7-alpine
    port: 6379
    volume: redis_data:/data

services:
  web:
    port: 8000
    replicas: 3  # 9 containers total
    env:
      ENVIRONMENT: production
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379

ssl:
  enabled: true
  email: security@myapp.com
```

**Deploy:**
```bash
# Staging (default)
podlift deploy

# Production (explicit)
podlift deploy --config podlift.production.yml
```

## With Deployment Hooks

**Scenario:** Run database migrations and tasks after deployment.

**When to use:** Django, Rails, or any app with migrations.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy
    labels: [primary]

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

services:
  web:
    port: 8000
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp

hooks:
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
    - docker exec myapp-web-1 python manage.py collectstatic --noinput

ssl:
  enabled: true
  email: admin@myapp.com
```

## Separate Database Server

**Scenario:** Database on dedicated server.

**When to use:** Large apps, need database isolation, managed databases.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  web:
    - host: 192.168.1.10
      user: deploy
    - host: 192.168.1.11
      user: deploy
  
  db:
    - host: 192.168.1.20
      user: deploy
      labels: [database]

dependencies:
  postgres:
    image: postgres:16
    role: db  # Deploys to db servers
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

services:
  web:
    port: 8000
    replicas: 2
    env:
      # Connect to DB server directly
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@192.168.1.20:5432/myapp

ssl:
  enabled: true
  email: admin@myapp.com
```

## Air-Gapped Deployment (No Registry)

**Scenario:** Deploy without internet access or registry.

**When to use:** Secure environments, private networks.

```yaml
service: myapp
domain: myapp.internal
image: myapp

# No registry - uses SCP to transfer image

servers:
  - host: 192.168.10.10
    user: deploy

services:
  web:
    port: 8000
    env:
      SECRET_KEY: ${SECRET_KEY}

proxy:
  enabled: true
  ssl: false  # Internal network, no public SSL
```

**Deploy:**
```bash
# Builds image locally, transfers via SCP
podlift deploy
```

## Microservices (Multiple Services)

**Scenario:** Multiple services in one deployment.

**When to use:** API + frontend, multiple microservices.

```yaml
service: myapp
domain: myapp.com
image: myapp

servers:
  - host: 192.168.1.10
    user: deploy
    labels: [primary]

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

services:
  api:
    port: 8000
    replicas: 2
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
  
  frontend:
    port: 3000
    replicas: 2
    env:
      API_URL: http://localhost:8000
  
  worker:
    command: python worker.py
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp

proxy:
  enabled: true

ssl:
  enabled: true
  email: admin@myapp.com
```

## FAQ

### Which config should I start with?

Start with **"Single Server with SSL"** for most projects. It's production-ready and simple.

### Do I need a registry?

No. By default, podlift uses SCP to transfer images. Registries are optional and useful for CI/CD.

### How many replicas should I use?

Start with 1-2 replicas per service. Increase based on load. More replicas = more memory needed.

### Should I separate web and worker servers?

Only if you have heavy background jobs. Start with everything on one server.

### How do I handle multiple environments?

Create separate config files: `podlift.yml` (staging), `podlift.production.yml` (production).

### Can I use managed databases (RDS, etc)?

Yes! Don't define in `dependencies`. Just set `DATABASE_URL` to point to your managed database.

### What about MongoDB, MySQL, etc?

Add them in `dependencies` just like PostgreSQL. See [Configuration Reference](configuration.md).

## Next Steps

- [Configuration Reference](configuration.md) - All available options
- [Commands Reference](commands.md) - All commands and flags
- [Deployment Guide](deployment-guide.md) - Best practices and workflows

