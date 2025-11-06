# E2E Testing Guide

End-to-end testing for podlift using real VMs.

## Quick Start

### Option 1: Multipass (Recommended)

**Install Multipass:**
```bash
brew install multipass
```

**Run example test:**
```bash
./tests/e2e/multipass-example.sh
```

This will:
1. Launch an Ubuntu VM
2. Install Docker
3. Setup SSH access
4. Give you the VM IP to test with podlift

### Option 2: Vagrant

**Install Vagrant:**
```bash
brew install vagrant
brew install --cask virtualbox
```

**Create Vagrantfile** (see `docs/notes/VM_TESTING_OPTIONS.md`)

**Start VMs:**
```bash
vagrant up
```

## Manual Testing Flow

### 1. Launch Test VM

```bash
# With Multipass
multipass launch --name test-server --cpus 2 --memory 2G

# Install Docker
multipass exec test-server -- bash -c "curl -fsSL https://get.docker.com | sh"

# Get IP
multipass info test-server | grep IPv4
```

### 2. Create Test App

```bash
# Create simple test app
mkdir test-app
cd test-app

# Create Dockerfile
cat > Dockerfile <<EOF
FROM nginx:alpine
COPY index.html /usr/share/nginx/html/
EOF

# Create index.html
echo "<h1>podlift test deployment</h1>" > index.html

# Create podlift.yml
cat > podlift.yml <<EOF
service: testapp
image: testapp

servers:
  - host: YOUR_VM_IP  # From multipass info
    user: ubuntu
    ssh_key: ~/.ssh/id_rsa

services:
  web:
    port: 80
EOF
```

### 3. Test Deployment

```bash
# Validate config
podlift validate

# Deploy
podlift deploy

# Check status
podlift ps

# View logs
podlift logs web

# Test
curl http://YOUR_VM_IP
```

### 4. Test Rollback

```bash
# Make a change
echo "<h1>Version 2</h1>" > index.html
git commit -am "Version 2"

# Deploy new version
podlift deploy

# Rollback
podlift rollback

# Verify
curl http://YOUR_VM_IP  # Should show original version
```

### 5. Cleanup

```bash
multipass delete test-server
multipass purge
```

## Multi-Server Testing

### Launch Multiple VMs

```bash
# Web servers
multipass launch --name web1 --cpus 2 --memory 2G
multipass launch --name web2 --cpus 2 --memory 2G

# Database server
multipass launch --name db1 --cpus 2 --memory 4G

# Install Docker on all
for vm in web1 web2 db1; do
  multipass exec $vm -- bash -c "curl -fsSL https://get.docker.com | sh"
done

# Get IPs
multipass list
```

### Create Multi-Server Config

```yaml
service: myapp
image: myapp

servers:
  web:
    - host: WEB1_IP
      user: ubuntu
      labels: [primary]
    - host: WEB2_IP
      user: ubuntu

  db:
    - host: DB1_IP
      user: ubuntu

dependencies:
  postgres:
    image: postgres:16
    role: db  # Deploy to db server
    port: 5432

services:
  web:
    port: 8000
    replicas: 2
```

### Test Deployment

```bash
podlift deploy

# Verify on all servers
for vm in web1 web2 db1; do
  echo "=== $vm ==="
  multipass exec $vm -- docker ps
done
```

## Automated E2E Tests

See `tests/e2e/multipass-example.sh` for a complete automated test script.

**Run automated test:**
```bash
./tests/e2e/multipass-example.sh
```

## Test Scenarios

### Scenario 1: Basic Deployment
- [x] Single server
- [x] Docker image deployment
- [x] Container starts
- [x] Health check passes

### Scenario 2: Zero-Downtime Deployment
- [ ] Deploy v1
- [ ] Make change, deploy v2
- [ ] Verify no downtime (continuous curl)
- [ ] Old containers stopped

### Scenario 3: Rollback
- [ ] Deploy v1
- [ ] Deploy v2
- [ ] Rollback to v1
- [ ] Verify v1 running

### Scenario 4: Multi-Server
- [ ] Deploy to 2+ servers
- [ ] Verify all servers updated
- [ ] Load balancing works

### Scenario 5: Dependencies
- [ ] Deploy with postgres
- [ ] App connects to postgres
- [ ] Dependencies persist across redeploys

### Scenario 6: SSL
- [ ] Setup domain
- [ ] Run `podlift ssl setup`
- [ ] HTTPS works
- [ ] Auto-renewal configured

## Troubleshooting

### VM won't start
```bash
multipass list
multipass delete <name>
multipass purge
```

### Can't SSH to VM
```bash
# Check VM is running
multipass list

# Get shell directly
multipass shell <name>

# Check SSH key
cat ~/.ssh/id_rsa.pub
```

### Docker not installed
```bash
multipass exec <name> -- bash -c "curl -fsSL https://get.docker.com | sh"
```

### Can't connect from podlift
```bash
# Test SSH manually
ssh ubuntu@VM_IP

# Check firewall
multipass exec <name> -- sudo ufw status
```

## CI/CD Integration

For GitHub Actions, use cloud VMs instead of local VMs:

```yaml
# .github/workflows/e2e.yml
- name: Create test server
  run: |
    # Use DigitalOcean/Hetzner API
    # Create droplet
    # Run tests
    # Destroy droplet
```

## Cost

**Local (Free):**
- Multipass: Free
- Vagrant + VirtualBox: Free

**Cloud (Per test run):**
- DigitalOcean: ~$0.01
- Hetzner: ~$0.01

## Next Steps

1. Implement `podlift deploy` command (Phase 1)
2. Add E2E tests for each scenario
3. Automate in CI/CD
4. Add performance benchmarks

## Resources

- [Multipass Documentation](https://multipass.run/docs)
- [Vagrant Documentation](https://www.vagrantup.com/docs)
- [VM Testing Options](../docs/notes/VM_TESTING_OPTIONS.md)

