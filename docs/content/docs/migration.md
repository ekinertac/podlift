---
title: Migration Guide
weight: 80
---

# Migration Guide

How to migrate to podlift from other deployment tools.

## From Kamal

podlift was designed to fix Kamal's pain points while keeping the good ideas.

### Key Differences

| Aspect | Kamal | podlift |
|--------|-------|---------|
| **Proxy** | kamal-proxy (custom) | nginx (standard) |
| **Image transfer** | Registry only | SCP or registry |
| **Validation** | During deployment | Before deployment |
| **Error messages** | Opaque | Clear with solutions |
| **Prompts** | Interactive | Non-interactive |
| **State** | Unknown | Docker labels + JSON |
| **Escape hatches** | Limited | Many (--skip flags) |

### Configuration Comparison

#### Kamal (config/deploy.yml)

```yaml
service: myapp
image: myapp

servers:
  web:
    hosts:
      - 192.168.1.10
      - 192.168.1.11

registry:
  server: ghcr.io
  username:
    - KAMAL_REGISTRY_PASSWORD
  password:
    - KAMAL_REGISTRY_PASSWORD

accessories:
  db:
    image: postgres:16
    host: 192.168.1.10
    env:
      POSTGRES_PASSWORD: secret

proxy:
  ssl: true
  host: myapp.com
```

#### podlift (podlift.yml)

```yaml
service: myapp
image: myapp
domain: myapp.com

servers:
  web:
    - host: 192.168.1.10
    - host: 192.168.1.11

registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}

proxy:
  ssl: letsencrypt
  ssl_email: admin@myapp.com
```

### Migration Steps

#### 1. Install podlift

```bash
curl -sSL https://podlift.sh/install.sh | sh
```

#### 2. Convert Configuration

In your project root:

```bash
# Remove Kamal files
rm -rf .kamal/
rm -rf config/deploy.yml

# Initialize podlift
podlift init
```

Edit `podlift.yml` based on your Kamal config:

**Servers:**
```yaml
# Kamal
servers:
  web:
    hosts:
      - 192.168.1.10

# podlift
servers:
  web:
    - host: 192.168.1.10
```

**Accessories → Dependencies:**
```yaml
# Kamal
accessories:
  db:
    image: postgres:16

# podlift
dependencies:
  postgres:
    image: postgres:16
```

**Registry password:**
```yaml
# Kamal
password:
  - KAMAL_REGISTRY_PASSWORD

# podlift
password: ${REGISTRY_PASSWORD}
```

#### 3. Move Secrets

```bash
# Kamal secrets file
cat .kamal/secrets

# Create podlift .env
vim .env
```

Example `.env`:
```bash
REGISTRY_USER=myuser
REGISTRY_PASSWORD=ghp_xxx
DB_PASSWORD=postgres-password
SECRET_KEY=django-secret-key
```

#### 4. Clean Up Kamal Deployment (Optional)

On your servers:

```bash
# Remove Kamal containers
kamal remove

# Or manually via SSH
ssh root@192.168.1.10
docker stop $(docker ps -q --filter "label=service=myapp")
docker rm $(docker ps -aq --filter "label=service=myapp")
```

#### 5. Deploy with podlift

```bash
podlift validate
podlift deploy
```

### What You'll Notice

**Better:**
- Validation catches errors before deployment
- Clear error messages with solutions
- No interactive prompts
- Can use SCP instead of registry
- nginx instead of custom proxy (easier to debug)

**Different:**
- Must have clean git state (enforced)
- nginx config is generated (but editable)
- State is visible and inspectable

**Same:**
- Zero-downtime deployments still work
- Rollback still works
- SSH-based deployment

### Kamal Commands → podlift Commands

```bash
# Initialize
kamal init          → podlift init

# Deploy
kamal deploy        → podlift deploy

# Status
kamal details       → podlift ps
kamal app logs      → podlift logs web

# Rollback
kamal rollback      → podlift rollback

# Execute
kamal app exec      → podlift exec

# Remove (no direct equivalent)
kamal remove        → SSH and docker stop/rm

# Configuration
kamal config        → podlift config
```

---

## From Ansible + Docker

If you're using Ansible playbooks for Docker deployment.

### What podlift Replaces

Your Ansible setup probably has:
- Playbook for Docker installation
- Playbook for image build/push
- Playbook for container deployment
- Playbook for nginx configuration
- Playbook for SSL setup

podlift does all of this with one command: `podlift deploy`

### Migration Steps

#### 1. Identify Current Setup

Typical Ansible structure:
```
ansible/
├── playbooks/
│   ├── setup.yml
│   ├── deploy.yml
│   └── nginx.yml
├── inventory/
│   └── hosts.ini
└── vars/
    └── main.yml
```

#### 2. Convert to podlift Configuration

**Ansible inventory:**
```ini
[web]
192.168.1.10
192.168.1.11

[workers]
192.168.1.12
```

**podlift.yml:**
```yaml
servers:
  web:
    - host: 192.168.1.10
    - host: 192.168.1.11
  worker:
    - host: 192.168.1.12
```

**Ansible vars:**
```yaml
app_name: myapp
app_port: 8000
postgres_password: "{{ vault_postgres_password }}"
```

**podlift.yml + .env:**
```yaml
# podlift.yml
service: myapp
services:
  web:
    port: 8000
```

```bash
# .env
DB_PASSWORD=your-secure-password
```

#### 3. Deploy

```bash
podlift validate
podlift deploy
```

### What You'll Lose

- **Ansible's flexibility**: Can't run arbitrary tasks
- **Ansible Vault**: Use environment variables instead
- **Custom playbooks**: Use hooks for post-deploy tasks

### What You'll Gain

- **Simplicity**: One config file vs many playbooks
- **Speed**: Faster deploys (no Ansible overhead)
- **Git integration**: Automatic versioning
- **Better errors**: Clearer than Ansible output

### Keeping Ansible for Setup

You can keep Ansible for initial server setup:

```yaml
# ansible/setup.yml
- hosts: all
  tasks:
    - name: Install Docker
      shell: curl -fsSL https://get.docker.com | sh
    
    - name: Install podlift
      shell: curl -sSL https://podlift.sh/install.sh | sh
```

Then use podlift for all deployments:
```bash
ansible-playbook ansible/setup.yml  # Once
podlift deploy                      # Every deploy
```

---

## From Manual Docker Commands

If you SSH to servers and run `docker run` manually.

### Before (Manual)

```bash
# SSH to server
ssh root@192.168.1.10

# Pull image
docker pull myapp:latest

# Stop old container
docker stop myapp-web
docker rm myapp-web

# Start new container
docker run -d \
  --name myapp-web \
  -p 80:8000 \
  -e SECRET_KEY=xxx \
  -e DATABASE_URL=xxx \
  myapp:latest

# Check logs
docker logs -f myapp-web
```

**Problems:**
- Manual steps (error-prone)
- Downtime during container swap
- Hard to rollback
- No version tracking
- Inconsistent across servers

### After (podlift)

```yaml
# podlift.yml
service: myapp
image: myapp

servers:
  - host: 192.168.1.10

services:
  web:
    port: 8000
    env:
      SECRET_KEY: ${SECRET_KEY}
      DATABASE_URL: ${DATABASE_URL}
```

```bash
# Deploy
podlift deploy

# Logs
podlift logs web --follow

# Rollback if needed
podlift rollback
```

**Benefits:**
- One command deployment
- Zero-downtime
- Automatic rollback capability
- Version tracking via git
- Repeatable deploys

### Migration

#### 1. Document Current Setup

Write down:
- What environment variables you use
- What ports you expose
- What volumes you mount
- What other containers run (postgres, redis, etc.)

#### 2. Create podlift.yml

```yaml
service: myapp
image: myapp

servers:
  - host: YOUR_SERVER_IP

services:
  web:
    port: 8000
    env:
      SECRET_KEY: ${SECRET_KEY}
      DATABASE_URL: ${DATABASE_URL}
    volumes:
      - /data/uploads:/app/uploads

dependencies:
  postgres:
    image: postgres:16
    volume: postgres_data:/var/lib/postgresql/data
```

#### 3. Stop Manual Containers

```bash
ssh root@192.168.1.10

# Stop your manually started containers
docker stop myapp-web
docker rm myapp-web

# Don't stop databases (podlift will reuse them)
```

#### 4. Deploy

```bash
podlift deploy
```

podlift will:
- Build image
- Transfer to server
- Start containers
- Set up nginx
- Configure healthchecks

---

## From Heroku

Moving from Heroku to self-hosted with podlift.

### What Changes

| Heroku | podlift |
|--------|---------|
| `git push heroku main` | `podlift deploy` |
| Add-ons (postgres, redis) | `dependencies` in config |
| Config vars | `.env` file |
| Automatic SSL | `podlift ssl setup` |
| Logs | `podlift logs` |
| One-off dynos | `podlift exec` |

### Migration Steps

#### 1. Set Up Server

Unlike Heroku, you need a server:

```bash
# Create Ubuntu 22.04 server on any cloud provider
# DigitalOcean, AWS, Hetzner, etc.

# Get server IP: 192.168.1.10
```

#### 2. Create Dockerfile

Heroku uses buildpacks; you need a Dockerfile:

```dockerfile
# Example for Python/Django
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

CMD ["gunicorn", "myapp.wsgi:application", "--bind", "0.0.0.0:8000"]
```

#### 3. Export Heroku Config

```bash
heroku config --shell > .env
```

Edit `.env` and update:
```bash
# Change Heroku DATABASE_URL if needed
DATABASE_URL=postgres://postgres:${DB_PASSWORD}@primary:5432/myapp

# Remove Heroku-specific vars
# HEROKU_APP_NAME=...
```

#### 4. Create podlift.yml

```yaml
service: myapp
image: myapp
domain: myapp.com

servers:
  - host: 192.168.1.10

dependencies:
  postgres:
    image: postgres:16
    env:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
  
  redis:
    image: redis:7

services:
  web:
    port: 8000
    env:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@primary:5432/myapp
      REDIS_URL: redis://primary:6379

proxy:
  ssl: letsencrypt
  ssl_email: admin@myapp.com
```

#### 5. Deploy

```bash
podlift deploy
podlift ssl setup --email admin@myapp.com
```

#### 6. Update DNS

Point your domain from Heroku to your server:

```
A record: myapp.com → 192.168.1.10
```

### Cost Comparison

Heroku dyno ($25/mo) → DigitalOcean droplet ($6/mo)

You're now responsible for:
- Server maintenance
- Security updates
- Backups

But you gain:
- Full control
- Lower cost
- No vendor lock-in

---

## From Docker Compose

If you use `docker-compose.yml` for deployment.

### Before

```yaml
# docker-compose.yml
version: '3.8'

services:
  web:
    build: .
    ports:
      - "8000:8000"
    environment:
      - SECRET_KEY=${SECRET_KEY}
    depends_on:
      - postgres
  
  postgres:
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

Deploy manually:
```bash
ssh root@192.168.1.10
git pull
docker-compose up -d --build
```

### After (podlift)

```yaml
# podlift.yml
service: myapp
image: myapp

servers:
  - host: 192.168.1.10

dependencies:
  postgres:
    image: postgres:16
    volume: postgres_data:/var/lib/postgresql/data

services:
  web:
    port: 8000
    env:
      SECRET_KEY: ${SECRET_KEY}
```

Deploy:
```bash
podlift deploy
```

### What You Gain

- **Zero-downtime**: Compose stops containers before starting new ones
- **Git versioning**: Automatic tagging with commits
- **Rollback**: Easy revert to previous version
- **Multi-server**: Deploy to multiple servers
- **nginx proxy**: Automatic reverse proxy setup

### Migration

Your `docker-compose.yml` maps almost directly:

```yaml
# docker-compose
services:
  web:
    build: .
    ports: ["8000:8000"]

# podlift
services:
  web:
    port: 8000
```

```yaml
# docker-compose
services:
  postgres:
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data

# podlift
dependencies:
  postgres:
    image: postgres:16
    volume: postgres_data:/var/lib/postgresql/data
```

---

## Common Migration Gotchas

### 1. Environment Variables

**Problem**: Secrets in git

**Solution**: Use `.env` file (not committed)

```bash
echo ".env" >> .gitignore
git add .gitignore
git commit -m "Ignore .env"
```

### 2. Uncommitted Changes

**Problem**: podlift requires clean git state

**Solution**: Commit before deploying

```bash
git add -A
git commit -m "Changes"
podlift deploy
```

### 3. Port Conflicts

**Problem**: Existing services on ports 80/443

**Solution**: Stop conflicting services or use different ports

```bash
ssh root@192.168.1.10
systemctl stop apache2
systemctl disable apache2
```

### 4. Docker Not Installed

**Problem**: podlift assumes Docker is installed

**Solution**: Install Docker first

```bash
ssh root@192.168.1.10
curl -fsSL https://get.docker.com | sh
```

### 5. SSH Key Issues

**Problem**: Password-based SSH not supported

**Solution**: Use SSH keys

```bash
ssh-copy-id root@192.168.1.10
```

---

## Getting Help

Migrating from another tool? We want to help!

- [Installation Guide](installation.md)
- [Configuration Reference](configuration.md)
- [Troubleshooting](troubleshooting.md)

Still stuck? [Open an issue](https://github.com/ekinertac/podlift/issues) with:
- What you're migrating from
- Your current setup
- What's not working

We'll help you migrate.

