#!/bin/bash
# Example E2E test using Multipass
# This demonstrates how to test podlift with real VMs

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 podlift E2E Test with Multipass${NC}"
echo "=================================="
echo ""

# Check if multipass is installed
if ! command -v multipass &> /dev/null; then
    echo "❌ Multipass not installed"
    echo "Install with: brew install multipass"
    exit 1
fi

VM_NAME="podlift-test-$(date +%s)"

# Cleanup function
cleanup() {
    echo ""
    echo -e "${BLUE}🧹 Cleaning up...${NC}"
    multipass delete $VM_NAME 2>/dev/null || true
    multipass purge 2>/dev/null || true
    rm -f /tmp/podlift-test.yml
    echo -e "${GREEN}✓ Cleaned up${NC}"
}

trap cleanup EXIT

# 1. Launch VM (completely fresh, no Docker)
echo -e "${BLUE}📦 Launching fresh Ubuntu VM...${NC}"
echo "  (No Docker installed - podlift setup will handle it)"
multipass launch --name $VM_NAME --cpus 2 --memory 2G

echo -e "${GREEN}✓ VM launched${NC}"
echo ""

# 2. Wait for VM to be ready
echo -e "${BLUE}⏳ Waiting for VM to be ready...${NC}"
sleep 10

# 3. Get VM information
echo -e "${BLUE}📍 Getting VM information...${NC}"
IP=$(multipass info $VM_NAME | grep IPv4 | awk '{print $2}')
echo "  IP Address: $IP"
echo -e "${GREEN}✓ VM ready${NC}"
echo ""

# 4. Verify it's a fresh server (no Docker)
echo -e "${BLUE}🔍 Checking server state...${NC}"
if multipass exec $VM_NAME -- which docker &>/dev/null; then
    echo "  Docker: already installed"
else
    echo "  Docker: not installed ✓ (podlift setup will install it)"
fi
echo -e "${GREEN}✓ Fresh server ready${NC}"
echo ""

# NOTE: In production use, you would run:
# podlift setup   # Installs Docker, configures firewall, security
# But for now, let's install Docker manually for testing:
echo -e "${BLUE}🐳 Installing Docker (manual - podlift setup will automate this)...${NC}"
multipass exec $VM_NAME -- bash -c "curl -fsSL https://get.docker.com | sudo sh"
multipass exec $VM_NAME -- sudo usermod -aG docker ubuntu
echo -e "${GREEN}✓ Docker installed${NC}"
echo ""

# 5. Setup SSH access (copy public key)
echo -e "${BLUE}🔑 Setting up SSH access...${NC}"
if [ -f ~/.ssh/id_rsa.pub ]; then
    PUB_KEY=$(cat ~/.ssh/id_rsa.pub)
    multipass exec $VM_NAME -- bash -c "mkdir -p ~/.ssh && echo '$PUB_KEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
    echo -e "${GREEN}✓ SSH key copied${NC}"
else
    echo "⚠️  No SSH key found at ~/.ssh/id_rsa.pub"
    echo "   Generate one with: ssh-keygen -t rsa"
fi
echo ""

# 6. Test SSH connection
echo -e "${BLUE}🔗 Testing SSH connection...${NC}"
if ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ubuntu@$IP "echo 'SSH works'" 2>/dev/null; then
    echo -e "${GREEN}✓ SSH connection successful${NC}"
else
    echo "⚠️  SSH connection failed (might need manual key setup)"
fi
echo ""

# 7. Show what podlift config would look like
echo -e "${BLUE}📝 Example podlift.yml for this VM:${NC}"
cat <<EOF

service: myapp
image: myapp

servers:
  - host: $IP
    user: ubuntu
    ssh_key: ~/.ssh/id_rsa

# To test deployment, create the above config and run:
# podlift deploy

EOF

# 8. Show VM status
echo -e "${BLUE}📊 VM Status:${NC}"
multipass list

echo ""
echo -e "${GREEN}✅ E2E test environment ready!${NC}"
echo ""
echo "Next steps:"
echo "  1. SSH into VM: multipass shell $VM_NAME"
echo "  2. Get IP: multipass info $VM_NAME"
echo "  3. Test deploy: podlift deploy (with config pointing to $IP)"
echo "  4. Cleanup: multipass delete $VM_NAME && multipass purge"
echo ""
echo "VM will be automatically cleaned up when this script exits."
echo "Press Ctrl+C to cleanup and exit."
echo ""

# Keep VM running for manual testing
read -p "Press Enter to cleanup and exit..."

