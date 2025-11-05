# Troubleshooting

Common errors and how to fix them.

## Deployment Errors

### Error: Working tree has uncommitted changes

```
ERROR: Working tree has uncommitted changes

Modified files:
  M  app/views.py

Commit or stash your changes before deploying.
```

**Cause**: You have uncommitted changes in your git repository.

**Solution**: 
```bash
# Option 1: Commit changes
git add -A
git commit -m "Your message"

# Option 2: Stash changes
git stash

# Then deploy
podlift deploy
```

**Why this happens**: podlift enforces clean git state to ensure deployments are reproducible.

---

### Error: SSH connection failed

```
ERROR: SSH connection failed to 192.168.1.10

Connection timeout after 10s
```

**Causes**:
1. Server is down
2. Firewall blocking SSH
3. Wrong IP address
4. SSH not running on server

**Solutions**:

Test SSH manually:
```bash
ssh root@192.168.1.10
```

Check if server is reachable:
```bash
ping 192.168.1.10
```

Verify SSH is running on server:
```bash
# On the server
systemctl status sshd
```

Check firewall:
```bash
# On the server
ufw status
ufw allow 22/tcp
```

---

### Error: SSH authentication failed

```
ERROR: SSH authentication failed

Permission denied (publickey)
```

**Cause**: SSH key not authorized on server.

**Solution**:

Copy your SSH key to server:
```bash
ssh-copy-id root@192.168.1.10
```

Or manually add key:
```bash
# On your machine
cat ~/.ssh/id_rsa.pub

# On the server
echo "your-public-key" >> ~/.ssh/authorized_keys
```

Verify SSH config in `podlift.yml`:
```yaml
servers:
  - host: 192.168.1.10
    user: root
    ssh_key: ~/.ssh/id_rsa  # Correct path?
```

---

### Error: Docker not installed

```
ERROR: Docker not installed on 192.168.1.10

Command 'docker' not found
```

**Solution**:

Install Docker on server:
```bash
ssh root@192.168.1.10

# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# Verify
docker --version
```

---

### Error: Port already in use

```
ERROR: Port 5432 in use on 192.168.1.10

Required for: postgres
```

**Cause**: Another service is using the port.

**Solutions**:

Check what's using the port:
```bash
ssh root@192.168.1.10 'lsof -i :5432'
```

Option 1: Stop the conflicting service:
```bash
ssh root@192.168.1.10 'systemctl stop postgresql'
```

Option 2: Use a different port in `podlift.yml`:
```yaml
dependencies:
  postgres:
    port: 5433  # Different port
```

---

### Error: Health check failed

```
ERROR: Service 'web' failed healthcheck after 30s

Healthcheck: GET http://localhost:8000/health
Expected: [200]
Actual: 500 Internal Server Error
```

**Causes**:
1. Application not starting correctly
2. Database connection failed
3. Missing environment variables
4. Wrong healthcheck path

**Solutions**:

Check container logs:
```bash
podlift logs web
```

Common issues:

**Missing environment variables:**
```bash
# Check .env file exists and has required variables
cat .env

# Verify environment variables are set
podlift config --show-secrets
```

**Database not ready:**
```yaml
# Increase healthcheck timeout
services:
  web:
    healthcheck:
      timeout: 60s  # Give DB more time to start
```

**Wrong path:**
```yaml
# Update healthcheck path
services:
  web:
    healthcheck:
      path: /up  # Try different endpoint
```

**Debug interactively:**
```bash
# SSH to server and check container
ssh root@192.168.1.10
docker ps
docker logs myapp-web-xxx
docker exec -it myapp-web-xxx bash
```

---

### Error: Insufficient disk space

```
ERROR: Insufficient disk space on 192.168.1.10

Available: 2GB
Required: 10GB
```

**Solution**:

Clean up Docker:
```bash
ssh root@192.168.1.10

# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Remove unused containers
docker container prune
```

Check disk usage:
```bash
df -h
du -sh /var/lib/docker
```

---

## Registry Errors

### Error: Registry authentication failed

```
ERROR: Registry authentication failed

denied: denied
```

**Causes**:
1. Wrong username/password
2. Token expired
3. Missing scopes

**Solutions**:

Test login manually:
```bash
echo "$REGISTRY_PASSWORD" | docker login ghcr.io -u "$REGISTRY_USER" --password-stdin
```

For GitHub Container Registry:
1. Go to https://github.com/settings/tokens
2. Create new token
3. Select scopes: `read:packages`, `write:packages`
4. Copy token to `.env`:
```bash
REGISTRY_PASSWORD=ghp_xxxxxxxxxxxxx
```

Verify environment variables:
```bash
echo $REGISTRY_USER
echo $REGISTRY_PASSWORD
```

---

## SSL Errors

### Error: Domain not pointing to server

```
ERROR: SSL setup failed

Domain myapp.com does not resolve to 192.168.1.10
Current: 203.0.113.5
```

**Solution**:

Update DNS records to point to your server:
```
A record: myapp.com -> 192.168.1.10
```

Wait for DNS propagation (can take up to 48 hours, usually faster).

Verify DNS:
```bash
dig myapp.com
nslookup myapp.com
```

---

### Error: Ports 80/443 not open

```
ERROR: Cannot obtain SSL certificate

Port 80 or 443 not accessible from internet
```

**Solution**:

Open ports on firewall:
```bash
ssh root@192.168.1.10

# UFW
ufw allow 80/tcp
ufw allow 443/tcp

# iptables
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

Check if ports are listening:
```bash
netstat -tuln | grep ':80\|:443'
```

Test from outside:
```bash
curl http://your-server-ip
```

---

## Rollback Errors

### Error: No previous version found

```
ERROR: Cannot rollback

No previous deployment found
```

**Cause**: This is the first deployment, nothing to rollback to.

**Solution**: Deploy a working version first, then you can rollback to it.

---

### Error: Previous version image not found

```
ERROR: Image not found for version x9y8z7w

Image may have been manually deleted
```

**Solution**:

Specify a version that exists:
```bash
# List available versions
ssh root@192.168.1.10 'docker images | grep myapp'

# Rollback to specific version
podlift rollback --to abc123
```

Or redeploy from that git commit:
```bash
git checkout x9y8z7w
podlift deploy
```

---

## Configuration Errors

### Error: Invalid YAML syntax

```
ERROR: Configuration invalid

yaml: line 12: mapping values are not allowed in this context
```

**Cause**: YAML syntax error.

**Solution**:

Common YAML mistakes:

**Missing space after colon:**
```yaml
# Wrong
servers:
  -host: 192.168.1.10

# Correct
servers:
  - host: 192.168.1.10
```

**Inconsistent indentation:**
```yaml
# Wrong (mixing spaces and tabs)
services:
  web:
      port: 8000

# Correct (2 spaces)
services:
  web:
    port: 8000
```

Validate YAML online: https://www.yamllint.com/

---

### Error: Required field missing

```
ERROR: Configuration invalid

Missing required field: service
```

**Solution**:

Add required field to `podlift.yml`:
```yaml
service: myapp  # Required
```

Required fields:
- `service` - Service name
- `image` - Image name
- `servers` - At least one server

---

## Network Errors

### Error: Cannot connect to Docker daemon

```
ERROR: Cannot connect to Docker daemon on 192.168.1.10

Is Docker running?
```

**Solution**:

Start Docker on server:
```bash
ssh root@192.168.1.10
systemctl start docker
systemctl enable docker
```

Check Docker status:
```bash
systemctl status docker
```

Verify user has Docker permissions:
```bash
# Add user to docker group
usermod -aG docker your-user

# Logout and login again
```

---

## Performance Issues

### Deployment is slow

**Causes**:
1. Large image size
2. Slow network
3. No Docker layer caching

**Solutions**:

**Optimize Dockerfile:**
```dockerfile
# Use specific base image
FROM python:3.11-slim

# Order commands by change frequency
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
```

**Use .dockerignore:**
```
.git
node_modules
*.log
.env
```

**Use registry instead of SCP for multi-server:**
```yaml
registry:
  server: ghcr.io
  username: ${REGISTRY_USER}
  password: ${REGISTRY_PASSWORD}
```

Registry is faster when deploying to multiple servers (push once, pull many).

---

## Debug Mode

Enable verbose logging:

```bash
podlift deploy --verbose
```

Shows all commands executed:
```
[DEBUG] Executing: ssh root@192.168.1.10 'docker ps'
[DEBUG] Output: CONTAINER ID   IMAGE
[DEBUG] Executing: docker build -t myapp:abc123 .
```

This helps identify exactly where failures occur.

---

## Getting Help

If you're still stuck:

1. **Check logs**: `podlift logs web`
2. **Check server state**: SSH and inspect manually
3. **Validate config**: `podlift validate --verbose`
4. **Search issues**: https://github.com/yourusername/podlift/issues
5. **Ask for help**: Create new issue with:
   - podlift version (`podlift version`)
   - Configuration (sanitize secrets)
   - Full error message
   - Output of `podlift validate --verbose`

## Common Patterns

### Fresh Server Setup

Complete setup on a new Ubuntu server:

```bash
# On server
apt update
apt install -y docker.io
systemctl enable docker
systemctl start docker
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable

# On your machine
ssh-copy-id root@192.168.1.10
podlift validate
podlift deploy
podlift ssl setup --email admin@myapp.com
```

### Deployment Failed, How to Recover

```bash
# 1. Check what's running
podlift ps

# 2. Check logs for errors
podlift logs web

# 3. Fix the issue (code, config, etc.)

# 4. Redeploy
podlift deploy
```

Old containers keep running if deployment fails, so no downtime.

### Manual Container Management

If podlift fails and you need to manage manually:

```bash
# SSH to server
ssh root@192.168.1.10

# Check containers
docker ps -a

# View logs
docker logs myapp-web-abc123-1

# Stop container
docker stop myapp-web-abc123-1

# Start container
docker start myapp-web-abc123-1

# Remove container
docker rm myapp-web-abc123-1
```

Everything is standard Docker. You have full control.

