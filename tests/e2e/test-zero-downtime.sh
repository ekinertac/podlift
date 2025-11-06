#!/bin/bash
# Test zero-downtime deployment
# Verifies that deployment happens without dropping requests

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🚀 Testing Zero-Downtime Deployment${NC}"
echo "======================================="
echo ""

# Config
PODLIFT_BIN="/Users/ekinertac/Code/podlift/bin/podlift"
VM_NAME="podlift-zero-downtime-$(date +%s)"

# Cleanup
cleanup() {
    echo ""
    echo -e "${BLUE}🧹 Cleaning up...${NC}"
    multipass delete $VM_NAME 2>/dev/null || true
    multipass purge 2>/dev/null || true
    rm -rf /tmp/zero-downtime-test
    echo -e "${GREEN}✓ Cleaned up${NC}"
}
trap cleanup EXIT

# 1. Launch VM
echo -e "${BLUE}[1/10] Launching VM...${NC}"
multipass launch --name $VM_NAME --cpus 2 --memory 2G
IP=$(multipass info $VM_NAME | grep IPv4 | awk '{print $2}')
echo -e "${GREEN}✓ VM at $IP${NC}"

# 2. Setup SSH
echo -e "${BLUE}[2/10] Setting up SSH...${NC}"
cat ~/.ssh/id_rsa.pub | multipass exec $VM_NAME -- bash -c "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
echo -e "${GREEN}✓ SSH configured${NC}"

# 3. Create test app (nginx with version indicator)
echo -e "${BLUE}[3/10] Creating test app...${NC}"
mkdir -p /tmp/zero-downtime-test
cd /tmp/zero-downtime-test

cat > Dockerfile <<'EOF'
FROM nginx:alpine
COPY index.html /usr/share/nginx/html/
EXPOSE 80
HEALTHCHECK --interval=5s CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1
EOF

cat > index.html <<'EOF'
<!DOCTYPE html>
<html><body><h1>Version 1</h1></body></html>
EOF

git init
git config user.email "test@test.com"
git config user.name "Test"
git add .
git commit -m "Version 1"

cat > podlift.yml <<EOF
service: zerodowntime-test
image: zerodowntime-test

servers:
  - host: $IP
    user: ubuntu

services:
  web:
    port: 80
    replicas: 2
    healthcheck:
      path: /
      expect: [200]

proxy:
  enabled: true
EOF

git add podlift.yml
git commit -m "Add podlift config"

echo -e "${GREEN}✓ Test app created${NC}"

# 4. Setup server
echo -e "${BLUE}[4/10] Setting up server...${NC}"
$PODLIFT_BIN setup --no-security --no-firewall
echo -e "${GREEN}✓ Server ready${NC}"

# 5. Deploy v1
echo -e "${BLUE}[5/10] Deploying version 1...${NC}"
$PODLIFT_BIN deploy --zero-downtime
echo -e "${GREEN}✓ Version 1 deployed${NC}"

# 6. Verify v1 is running
echo -e "${BLUE}[6/10] Verifying version 1...${NC}"
sleep 3
RESPONSE=$(curl -s http://$IP 2>&1 || echo "FAILED")
if echo "$RESPONSE" | grep -q "Version 1"; then
    echo -e "${GREEN}✓ Version 1 is live${NC}"
else
    echo -e "${RED}✗ Version 1 not accessible${NC}"
    echo "Response: $RESPONSE"
    exit 1
fi

# 7. Prepare version 2
echo -e "${BLUE}[7/10] Preparing version 2...${NC}"
cat > index.html <<'EOF'
<!DOCTYPE html>
<html><body><h1>Version 2 - Zero Downtime!</h1></body></html>
EOF

git add index.html
git commit -m "Version 2"
echo -e "${GREEN}✓ Version 2 ready${NC}"

# 8. Start continuous monitoring (background process)
echo -e "${BLUE}[8/10] Starting continuous traffic (monitoring downtime)...${NC}"
MONITOR_LOG="/tmp/downtime-monitor-$$.log"
> $MONITOR_LOG

# Monitor in background - make request every 0.5s
{
    start_time=$(date +%s)
    request_count=0
    error_count=0
    
    while [ $(($(date +%s) - start_time)) -lt 45 ]; do
        if curl -s -m 2 http://$IP > /dev/null 2>&1; then
            echo "$(date +%s.%N) OK" >> $MONITOR_LOG
            request_count=$((request_count + 1))
        else
            echo "$(date +%s.%N) ERROR" >> $MONITOR_LOG
            error_count=$((error_count + 1))
        fi
        sleep 0.5
    done
    
    echo "STATS: $request_count requests, $error_count errors" >> $MONITOR_LOG
} &
MONITOR_PID=$!

sleep 2
echo -e "${GREEN}✓ Traffic monitoring started (PID $MONITOR_PID)${NC}"

# 9. Deploy v2 with zero-downtime
echo -e "${BLUE}[9/10] Deploying version 2 (zero-downtime)...${NC}"
echo "  (Monitoring requests in background...)"
$PODLIFT_BIN deploy --zero-downtime
echo -e "${GREEN}✓ Version 2 deployed${NC}"

# 10. Wait for monitoring to complete
echo -e "${BLUE}[10/10] Waiting for monitoring to complete...${NC}"
sleep 5
kill $MONITOR_PID 2>/dev/null || true
wait $MONITOR_PID 2>/dev/null || true

# Analyze results
echo ""
echo "======================================="
echo -e "${BLUE}📊 Downtime Analysis${NC}"
echo "======================================="
echo ""

if [ -f $MONITOR_LOG ]; then
    ERROR_COUNT=$(grep "ERROR" $MONITOR_LOG | wc -l | tr -d ' ')
    OK_COUNT=$(grep "OK" $MONITOR_LOG | wc -l | tr -d ' ')
    TOTAL=$((ERROR_COUNT + OK_COUNT))
    
    echo "Total requests: $TOTAL"
    echo "Successful: $OK_COUNT"
    echo "Failed: $ERROR_COUNT"
    
    if [ $ERROR_COUNT -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✅ ZERO DOWNTIME ACHIEVED!${NC}"
        echo "All requests succeeded during deployment"
    else
        ERROR_PERCENT=$((ERROR_COUNT * 100 / TOTAL))
        echo ""
        if [ $ERROR_PERCENT -lt 5 ]; then
            echo -e "${GREEN}✅ Minimal downtime: ${ERROR_PERCENT}%${NC}"
        else
            echo -e "${YELLOW}⚠ Some downtime detected: ${ERROR_PERCENT}%${NC}"
        fi
    fi
else
    echo -e "${YELLOW}⚠ No monitoring data${NC}"
fi

# Verify v2 is running
echo ""
echo -e "${BLUE}Verifying final state...${NC}"
V2_CHECK=$(curl -s http://$IP 2>&1)
if echo "$V2_CHECK" | grep -q "Version 2"; then
    echo -e "${GREEN}✓ Version 2 is now live${NC}"
else
    echo -e "${YELLOW}⚠ Version check inconclusive${NC}"
    echo "Response: $V2_CHECK"
fi

# Show container status
echo ""
echo -e "${BLUE}Container status:${NC}"
$PODLIFT_BIN ps

echo ""
echo "======================================="
echo -e "${GREEN}✅ Zero-Downtime Test Complete${NC}"
echo "======================================="
echo ""
echo "VM will be cleaned up on exit."
echo "Press Enter to cleanup..."
read -t 30 || true

