# FastAPI Example - Production-Ready Application

A complete FastAPI application for testing podlift deployments.

## Features

- ✅ FastAPI with async endpoints
- ✅ Health check endpoint (`/health`)
- ✅ Multiple API endpoints (`/api/users`, `/api/info`)
- ✅ Environment variable configuration
- ✅ PostgreSQL database (optional)
- ✅ Non-root user in container
- ✅ Docker health check
- ✅ Production-ready Dockerfile

## Quick Start

### Local Development

```bash
cd examples/fastapi

# Install dependencies
pip install -r requirements.txt

# Run locally
python main.py

# Test
curl http://localhost:8000/health
```

### With Docker Compose

```bash
cd examples/fastapi

# Start services
docker-compose up -d

# Test
curl http://localhost:8000/health

# Stop
docker-compose down
```

### With podlift (Production Deployment)

```bash
cd examples/fastapi

# Initialize git (if not already)
git init
git add .
git commit -m "Initial commit"

# Configure
cp .env.example .env
# Edit .env with your values

# Edit podlift.yml
# - Change YOUR_SERVER_IP to your server
# - Change domain

# Deploy
podlift setup      # Install Docker on server
podlift validate   # Check configuration
podlift deploy     # Deploy to production

# Verify
podlift ps         # Show running containers
podlift logs web   # View logs
```

## Endpoints

### GET /
Root endpoint with welcome message.

```bash
curl http://your-server:8000/
```

Response:
```json
{
  "message": "Welcome to podlift FastAPI example",
  "status": "running",
  "version": "abc123"
}
```

### GET /health
Health check endpoint for podlift monitoring.

```bash
curl http://your-server:8000/health
```

Response:
```json
{
  "status": "healthy",
  "uptime": 3600,
  "version": "abc123"
}
```

### GET /api/users
Example API endpoint returning user list.

```bash
curl http://your-server:8000/api/users
```

Response:
```json
{
  "users": [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"}
  ]
}
```

### GET /api/info
Returns environment information.

```bash
curl http://your-server:8000/api/info
```

Response:
```json
{
  "environment": "production",
  "secret_key_set": true,
  "database_url_set": true,
  "commit": "abc123"
}
```

## Environment Variables

- `ENVIRONMENT` - Environment name (default: production)
- `SECRET_KEY` - Application secret key (required)
- `DATABASE_URL` - PostgreSQL connection string (optional)
- `APP_VERSION` - Git commit hash (set by podlift)

## File Structure

```
examples/fastapi/
├── main.py              # FastAPI application
├── requirements.txt     # Python dependencies
├── Dockerfile           # Production Dockerfile
├── docker-compose.yml   # For local development
├── podlift.yml          # podlift configuration
├── .env.example         # Environment variables template
└── README.md            # This file
```

## Migrating from docker-compose

This example includes both `docker-compose.yml` and `podlift.yml` to show the migration path.

### docker-compose.yml → podlift.yml

| docker-compose | podlift |
|----------------|---------|
| `services.web.build` | Automatic (uses Dockerfile) |
| `services.web.ports` | `services.web.port` |
| `services.web.environment` | `services.web.env` |
| `services.db` | `dependencies.postgres` |
| `volumes` | `dependencies.postgres.volume` |
| `depends_on` | Automatic (dependencies start first) |

podlift handles:
- Multi-server deployment
- Zero-downtime updates
- SSL/TLS with Let's Encrypt
- Git-based versioning
- Rollbacks

## Testing podlift

This example is used for E2E testing:

```bash
cd examples/fastapi

# Run comprehensive E2E test
../../tests/e2e/test-fastapi.sh
```

Tests:
- ✅ Initial deployment
- ✅ Health check verification
- ✅ API endpoint testing
- ✅ Redeployment (update)
- ✅ Multiple replicas
- ✅ Dependency management (PostgreSQL)
- ✅ Environment variables
- ✅ Log viewing
- ✅ Container listing

## Production Deployment Checklist

- [ ] Update `podlift.yml` with your server IP
- [ ] Set `domain` in `podlift.yml`
- [ ] Copy `.env.example` to `.env`
- [ ] Set `SECRET_KEY` in `.env`
- [ ] Set `DB_PASSWORD` in `.env`
- [ ] Run `podlift setup` (first time only)
- [ ] Run `podlift validate`
- [ ] Run `podlift deploy`
- [ ] Test endpoints
- [ ] Configure DNS to point to server
- [ ] SSL will auto-configure with Let's Encrypt

## Troubleshooting

### Health check fails

```bash
# Check if app is running
podlift ps

# View logs
podlift logs web

# Check health endpoint manually
curl http://your-server:8000/health
```

### Database connection fails

```bash
# Check if postgres is running
ssh root@your-server 'docker ps | grep postgres'

# Check DATABASE_URL is set correctly
podlift logs web | grep DATABASE
```

## License

MIT - Free to use as example for your own applications.

