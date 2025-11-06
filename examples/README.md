# podlift Examples

Production-ready example applications for testing and learning podlift.

## Available Examples

### FastAPI (Python)

Complete FastAPI application with:
- Multiple API endpoints
- Health checks
- PostgreSQL integration
- Environment variables
- Docker Compose comparison

**Use for:**
- Learning podlift deployment
- E2E testing
- Migration from docker-compose

**Location:** `examples/fastapi/`

[View README](./fastapi/README.md)

## Running E2E Tests

Each example includes comprehensive E2E tests:

```bash
# FastAPI E2E test
./tests/e2e/test-fastapi.sh
```

Tests cover:
- Initial deployment
- Health checks
- API functionality
- Redeployment
- Multiple replicas
- Dependency management
- All podlift commands

## Using Examples for Your Project

1. **Copy the example:**
```bash
cp -r examples/fastapi my-project
cd my-project
```

2. **Customize:**
- Update `main.py` with your application logic
- Modify `requirements.txt` for your dependencies
- Update `podlift.yml` with your server
- Set environment variables in `.env`

3. **Deploy:**
```bash
podlift setup
podlift deploy
```

## Example Structure

Each example includes:
- **Application code** - Complete working app
- **Dockerfile** - Production-ready
- **docker-compose.yml** - For local development
- **podlift.yml** - podlift configuration
- **.env.example** - Environment variables template
- **README.md** - Complete documentation

## More Examples Coming

Planned examples:
- [ ] Next.js (Node.js)
- [ ] Rails (Ruby)
- [ ] Django (Python)
- [ ] Go web app
- [ ] Static site (nginx)

## Contributing Examples

Want to add an example? It should include:
- Working application with health check endpoint
- Production-ready Dockerfile
- podlift.yml configuration
- E2E test script
- README with setup instructions

See `examples/fastapi/` as template.

