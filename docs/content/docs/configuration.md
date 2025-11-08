# Configuration Reference

Complete reference for `podlift.yml` configuration file.

## Minimal Configuration

The smallest valid configuration:

```yaml
service: myapp
image: myapp

servers:
  - host: 192.168.1.10
```

This is enough to deploy. Everything else has sensible defaults.

## Full Configuration Example

```yaml
service: myapp
domain: myapp.com
image: myapp

# Git configuration (optional, auto-detected)
git:
  repo: git@github.com:user/repo.git
  branch: main

# Servers
servers:
  web:
    - host: 192.168.1.10
      user: root
      ssh_key: ~/.ssh/id_rsa
      labels: [primary]
    - host: 192.168.1.11
      user: deploy
      ssh_key: ~/.ssh/deploy_key

  worker:
    - host: 192.168.1.12
      user: root

# Container registry (optional)
registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}

# Dependencies (run on primary server)
dependencies:
  postgres:
    image: postgres:16
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: myapp
    
  redis:
    image: redis:7-alpine
    port: 6379

# Services
services:
  web:
    port: 8000
    replicas: 2
    healthcheck:
      path: /health
      expect: [200, 301, 302]
      timeout: 30s
      interval: 10s
    env:
      SECRET_KEY: ${SECRET_KEY}
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379
  
  worker:
    command: celery -A myapp worker
    replicas: 1
    env:
      SECRET_KEY: ${SECRET_KEY}
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp

# Proxy configuration
proxy:
  enabled: true
  ssl: letsencrypt
  ssl_email: admin@myapp.com

# Deployment hooks
hooks:
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
    - docker exec myapp-web-1 python manage.py collectstatic --noinput
```

## Configuration Fields

### env_file

**Optional**. Path to custom environment file.

```yaml
env_file: /path/to/custom.env
```

By default, podlift looks for `.env` in the same directory as `podlift.yml`. Use this field to specify a different location.

Supports:
- Absolute paths: `/etc/myapp/.env`
- Relative paths: `../shared/.env.production`
- Home directory: `~/.config/myapp/.env`

### service

**Required**. The name of your service. Used as container name prefix.

```yaml
service: myapp
```

Containers will be named: `myapp-web-abc123-1`, `myapp-worker-abc123-1`, etc.

### domain

**Optional**. Your application domain. Required for SSL.

```yaml
domain: myapp.com
```

If specified, nginx will be configured to serve this domain.

### image

**Required**. Docker image name (without registry prefix).

```yaml
image: myapp
```

With registry configured, full image is: `ghcr.io/username/myapp:abc123`

Without registry, image is built and transferred via SCP.

### git

**Optional**. Git repository configuration. Auto-detected from current directory.

```yaml
git:
  repo: git@github.com:user/repo.git
  branch: main
```

Used for displaying repository info in deployment logs.

### servers

**Required**. Server configuration. Can be simple list or grouped by role.

#### Simple list

```yaml
servers:
  - host: 192.168.1.10
  - host: 192.168.1.11
```

All services deploy to all servers.

#### Grouped by role

```yaml
servers:
  web:
    - host: 192.168.1.10
      user: root
      ssh_key: ~/.ssh/id_rsa
      labels: [primary]
    - host: 192.168.1.11
  
  worker:
    - host: 192.168.1.12
```

Services deploy only to their designated role.

#### Server fields

- `host` - **Required**. IP address or hostname
- `user` - Username for SSH (default: `root`)
- `ssh_key` - Path to SSH private key (default: `~/.ssh/id_rsa`)
- `port` - SSH port (default: `22`)
- `labels` - Array of labels (e.g., `[primary]` for dependency hosting)

### registry

**Optional**. Container registry configuration.

```yaml
registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}
```

If omitted, images are transferred via SCP (no registry needed).

#### Supported registries

- GitHub Container Registry: `ghcr.io`
- Docker Hub: `docker.io` (or omit server)
- Google Container Registry: `gcr.io`
- AWS ECR: `<account>.dkr.ecr.<region>.amazonaws.com`
- Any private registry

#### Fields

- `server` - Registry server (default: `docker.io`)
- `username` - Registry username (supports env vars)
- `password` - Registry password (supports env vars)

### dependencies

**Optional**. Services that run once (databases, caches, etc.).

```yaml
dependencies:
  postgres:
    image: postgres:16
    port: 5432
    volume: postgres_data:/var/lib/postgresql/data
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
```

By default, dependencies run on the server labeled `primary` (or first server if no label). You can deploy dependencies to specific servers using `host`, `role`, or `labels`.

#### Dependency fields

- `image` - **Required**. Docker image
- `host` - **Optional**. Deploy to specific server by IP/hostname
- `role` - **Optional**. Deploy to first server in specified role (e.g., `db`, `cache`)
- `labels` - **Optional**. Deploy to server matching these labels
- `port` - Port to expose (default: container default)
- `volume` - Persistent volume mapping
- `env` - Environment variables (supports `${VAR}` syntax)
- `command` - Override container command
- `options` - Additional Docker run options

**Placement priority:** If multiple placement options are specified, they're checked in this order:
1. `host` - Exact host match
2. `role` - First server in role
3. `labels` - Server with matching label
4. Default - Primary server

#### Common dependencies

PostgreSQL:
```yaml
postgres:
  image: postgres:16
  port: 5432
  volume: postgres_data:/var/lib/postgresql/data
  env:
    POSTGRES_PASSWORD: ${DB_PASSWORD}
    POSTGRES_DB: ${DB_NAME}
```

Redis:
```yaml
redis:
  image: redis:7-alpine
  port: 6379
  volume: redis_data:/data
```

MongoDB:
```yaml
mongo:
  image: mongo:7
  port: 27017
  volume: mongo_data:/data/db
  env:
    MONGO_INITDB_ROOT_USERNAME: ${MONGO_USER}
    MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD}
```

### services

**Optional**. Application services configuration.

```yaml
services:
  web:
    port: 8000
    replicas: 2
    healthcheck:
      path: /health
      expect: [200]
    env:
      SECRET_KEY: ${SECRET_KEY}
```

If omitted, single `web` service is assumed with defaults.

#### Service fields

- `port` - Port the service listens on (default: `8000`)
- `replicas` - Number of containers per server (default: `1`)
- `command` - Override container command
- `healthcheck` - Health check configuration
- `env` - Environment variables
- `volumes` - Volume mounts

#### Health check configuration

```yaml
healthcheck:
  path: /health              # HTTP path to check
  expect: [200, 301, 302]    # Accepted status codes
  timeout: 30s               # Timeout per check
  interval: 10s              # Time between checks
  retries: 3                 # Retries before failure
```

If `healthcheck` is omitted, Docker's `HEALTHCHECK` instruction is used.

Set `healthcheck: false` to disable (for workers).

#### Environment variables

```yaml
env:
  SECRET_KEY: ${SECRET_KEY}                    # From .env file
  DATABASE_URL: postgres://user:pass@host/db   # Literal value
  DEBUG: "false"                                # Quoted for booleans
```

Environment variables support:
- `${VAR}` - Read from `.env` file or environment
- `${VAR:-default}` - Default value if not set
- Literal values

### proxy

**Optional**. Reverse proxy configuration.

```yaml
proxy:
  enabled: true
  ssl: letsencrypt
  ssl_email: admin@myapp.com
```

#### Fields

- `enabled` - Enable nginx proxy (default: `true`)
- `ssl` - SSL mode: `letsencrypt`, `manual`, or `false`
- `ssl_email` - Email for Let's Encrypt notifications

If `enabled: false`, containers are exposed directly on their ports.

### hooks

**Optional**. Commands to run at specific deployment stages.

```yaml
hooks:
  after_deploy:
    - docker exec myapp-web-1 python manage.py migrate
    - docker exec myapp-web-1 python manage.py collectstatic --noinput
```

#### Available hooks

- `after_deploy` - After new containers start
- `before_deploy` - Before deployment begins
- `after_rollback` - After rollback completes

Hooks run on the primary server via SSH.

## Environment Variables

Environment variables are read from `.env` file **in the same directory as `podlift.yml`** (by default).

Example `.env`:
```bash
REGISTRY_USER=myuser
REGISTRY_PASSWORD=ghp_xxx
SECRET_KEY=django-secret-key
DB_PASSWORD=postgres-password
```

**Default file location:**
```
myapp/
├── podlift.yml
├── .env          ← Default location (same directory as podlift.yml)
├── Dockerfile
└── src/
```

**Custom env file path:**

You can specify a custom path in `podlift.yml`:

```yaml
# Use a custom env file
env_file: /path/to/custom.env

# Or use relative path
env_file: ../shared/.env.production

# Or use tilde for home directory
env_file: ~/.config/myapp/.env
```

**Important**: 
- Never commit `.env` to git. Add to `.gitignore`.
- By default, `.env` is in the same directory as `podlift.yml`
- Use `env_file` to specify a custom location
- Absolute paths, relative paths, and `~` expansion are supported

Reference in `podlift.yml`:
```yaml
registry:
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}

services:
  web:
    env:
      SECRET_KEY: ${SECRET_KEY}
```

## Defaults

If a field is omitted, these defaults apply:

```yaml
# Server defaults
user: root
ssh_key: ~/.ssh/id_rsa
port: 22

# Service defaults
services:
  web:
    port: 8000
    replicas: 1
    healthcheck:
      path: /health
      expect: [200]
      timeout: 30s
      interval: 10s
      retries: 3

# Proxy defaults
proxy:
  enabled: true
  ssl: false
```

## Configuration Validation

Run `podlift validate` to check configuration:

```bash
$ podlift validate

✓ Configuration valid
✓ All required fields present
✓ Server hostnames valid
✓ Port numbers in valid range
✓ Environment variables set
```

Invalid configuration example:
```bash
$ podlift validate

✗ Configuration invalid

Error: services.web.port must be between 1-65535 (got: 70000)
Error: registry.username is required when registry.server is set
Error: Environment variable SECRET_KEY not set
```

## Multiple Environments

Use different config files for different environments:

```bash
# Production
podlift deploy --config podlift.prod.yml

# Staging
podlift deploy --config podlift.staging.yml
```

Or use environment variable substitution:

```yaml
# podlift.yml
servers:
  - host: ${SERVER_HOST}

services:
  web:
    env:
      ENVIRONMENT: ${ENVIRONMENT}
```

Then:
```bash
# Production
ENVIRONMENT=production SERVER_HOST=192.168.1.10 podlift deploy

# Staging
ENVIRONMENT=staging SERVER_HOST=192.168.2.20 podlift deploy
```

## Advanced Patterns

### Separated Infrastructure

Dependencies can be deployed to separate servers for better separation of concerns:

```yaml
servers:
  web:
    - host: 192.168.1.10
    - host: 192.168.1.11
  
  db:
    - host: 192.168.1.20
      labels: [database]
  
  cache:
    - host: 192.168.1.30
      labels: [cache]

dependencies:
  postgres:
    image: postgres:16
    role: db  # On 192.168.1.20
    port: 5432
  
  redis:
    image: redis:7
    labels: [cache]  # On 192.168.1.30
    port: 6379

services:
  web:
    env:
      # Connect to dependencies by their server IPs
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@192.168.1.20:5432/myapp
      REDIS_URL: redis://192.168.1.30:6379
```

This allows complete separation:
- Web servers: 192.168.1.10, 192.168.1.11
- Database: 192.168.1.20
- Cache: 192.168.1.30

### Multiple Replicas Per Server

Run multiple containers on each server for better resource usage:

```yaml
services:
  web:
    replicas: 4  # 4 containers per server
```

nginx load balances between all replicas across all servers.

### Worker-Only Servers

Deploy workers separately from web servers:

```yaml
servers:
  web:
    - host: 192.168.1.10
  
  worker:
    - host: 192.168.1.20
    - host: 192.168.1.21

services:
  web:
    port: 8000
  
  worker:
    command: celery -A myapp worker
```

### Custom Docker Run Options

Pass additional options to `docker run`:

```yaml
services:
  web:
    options:
      memory: 2g
      cpus: 2
      restart: always
```

Translates to: `docker run --memory=2g --cpus=2 --restart=always ...`

