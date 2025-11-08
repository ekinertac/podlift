---
title: podlift
type: docs
---

# podlift

**Production-ready Docker deployment tool with zero-downtime deployments**

podlift is a modern deployment tool that makes deploying containerized applications simple, fast, and reliable. Built for developers who want production-grade deployments without the complexity of Kubernetes.

## Quick Start

Install podlift:

```bash
curl -sSL https://raw.githubusercontent.com/ekinertac/podlift/main/install.sh | sh
```

Deploy your first app:

```bash
podlift init                    # Create configuration
podlift setup                   # Prepare server
podlift deploy                  # Deploy with zero-downtime
```

## Key Features

### Zero-Downtime Deployments (Default)
Deploy new versions without any service interruption. podlift automatically:
- Starts new containers on temporary ports
- Waits for health checks to pass
- Switches nginx upstream atomically
- Drains connections from old containers
- Cleans up gracefully

### Automatic Load Balancing
Deploy to multiple servers and podlift automatically:
- Sets up nginx load balancer
- Configures health checks
- Manages upstream pools
- Handles server failures

### SSL/HTTPS Made Easy
One command to set up HTTPS:
```bash
podlift ssl setup
```
podlift handles Let's Encrypt certificates, nginx configuration, and automatic renewals.

### Smart Dependency Management
Deploy and manage PostgreSQL, Redis, MongoDB, and more:
```yaml
dependencies:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: secret
    volumes:
      - pgdata:/var/lib/postgresql/data
    health_check:
      cmd: pg_isready -U postgres
```

### Universal Registry Support
Works with any Docker registry:
- Docker Hub
- GitHub Container Registry
- GitLab Registry
- AWS ECR
- Or use SCP for air-gapped deployments

### Deployment Hooks
Run custom scripts at any stage:
```yaml
hooks:
  before_deploy: ./scripts/backup.sh
  after_deploy: ./scripts/notify-slack.sh
  after_rollback: ./scripts/alert-team.sh
```

## Why podlift?

**Simple** - One configuration file, intuitive commands  
**Fast** - Zero-downtime deployments in seconds  
**Reliable** - Automatic health checks and rollbacks  
**Production-Ready** - SSL, load balancing, dependencies  
**Non-Interactive** - Perfect for CI/CD pipelines

## Learn More

{{% columns %}}

### Getting Started
- [Installation](docs/installation)
- [Configuration](docs/configuration)
- [Deployment Guide](docs/deployment-guide)

<--->

### Features
- [Commands Reference](docs/commands)
- [How It Works](docs/how-it-works)
- [Migration Guide](docs/migration)

<--->

### Support
- [Troubleshooting](docs/troubleshooting)
- [GitHub Issues](https://github.com/ekinertac/podlift/issues)
- [Releases](https://github.com/ekinertac/podlift/releases)

{{% /columns %}}

## Quick Example

```yaml
# podlift.yml
service:
  name: myapp
  domain: myapp.com
  image: myapp:latest
  port: 3000
  replicas: 2

servers:
  - host: 1.2.3.4
    user: deploy
  - host: 5.6.7.8
    user: deploy

ssl:
  enabled: true
  email: admin@myapp.com

dependencies:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
```

Deploy:
```bash
podlift deploy
```

That's it! Your app is live with:
- Zero-downtime deployment
- Automatic load balancing
- HTTPS enabled
- PostgreSQL running
- Health checks monitoring

## Open Source

podlift is open source and available on [GitHub](https://github.com/ekinertac/podlift). Contributions welcome!

**License:** MIT

