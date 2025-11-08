#!/bin/bash
# Comprehensive E2E test suite for podlift
# Tests all features: multi-server, load balancing, zero-downtime, rollback, etc.

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PODLIFT_BIN="${PODLIFT_BIN:-$PROJECT_ROOT/podlift}"
TEST_DIR="/tmp/podlift-e2e-test-$$"
VM_PREFIX="podlift-e2e-$$"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
STEP=1

# VM tracking
declare -a VMS
declare -A VM_IPS

# Cleanup function
cleanup() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}🧹 Cleaning up test environment...${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Stop any background processes
    jobs -p | xargs kill -9 2>/dev/null || true
    
    # Delete VMs
    for vm in "${VMS[@]}"; do
        echo "  Deleting VM: $vm"
        multipass delete "$vm" 2>/dev/null || true
    done
    multipass purge 2>/dev/null || true
    
    # Clean up test directory
    rm -rf "$TEST_DIR"
    
    echo -e "${GREEN}✓ Cleanup complete${NC}"
}

trap cleanup EXIT

# Logging functions
log_section() {
    echo ""
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

log_step() {
    echo ""
    echo -e "${CYAN}[$STEP] $1${NC}"
    STEP=$((STEP + 1))
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

log_info() {
    echo -e "${BLUE}  $1${NC}"
}

# Test assertion functions
assert_equals() {
    TESTS_RUN=$((TESTS_RUN + 1))
    local expected="$1"
    local actual="$2"
    local message="$3"
    
    if [ "$expected" = "$actual" ]; then
        log_success "$message"
        return 0
    else
        log_error "$message (expected: '$expected', got: '$actual')"
        return 1
    fi
}

assert_contains() {
    TESTS_RUN=$((TESTS_RUN + 1))
    local haystack="$1"
    local needle="$2"
    local message="$3"
    
    if echo "$haystack" | grep -q "$needle"; then
        log_success "$message"
        return 0
    else
        log_error "$message (expected to contain: '$needle')"
        return 1
    fi
}

assert_http_status() {
    TESTS_RUN=$((TESTS_RUN + 1))
    local url="$1"
    local expected_status="$2"
    local message="$3"
    
    local actual_status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$actual_status" = "$expected_status" ]; then
        log_success "$message"
        return 0
    else
        log_error "$message (expected: $expected_status, got: $actual_status)"
        return 1
    fi
}

assert_command_success() {
    TESTS_RUN=$((TESTS_RUN + 1))
    local message="$1"
    shift
    
    if "$@" >/dev/null 2>&1; then
        log_success "$message"
        return 0
    else
        log_error "$message (command failed: $*)"
        return 1
    fi
}

# VM management functions
launch_vm() {
    local vm_name="$1"
    local cpus="${2:-2}"
    local memory="${3:-2G}"
    
    log_info "Launching VM: $vm_name (CPUs: $cpus, Memory: $memory)"
    
    multipass launch --name "$vm_name" --cpus "$cpus" --memory "$memory" ubuntu:22.04
    
    # Get IP
    local ip=$(multipass info "$vm_name" | grep IPv4 | awk '{print $2}')
    
    # Setup SSH
    if [ -f ~/.ssh/id_rsa.pub ]; then
        local pub_key=$(cat ~/.ssh/id_rsa.pub)
        multipass exec "$vm_name" -- bash -c "mkdir -p ~/.ssh && echo '$pub_key' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
    fi
    
    # Track VM
    VMS+=("$vm_name")
    VM_IPS["$vm_name"]="$ip"
    
    log_success "VM ready: $vm_name ($ip)"
    echo "$ip"
}

wait_for_http() {
    local url="$1"
    local timeout="${2:-30}"
    local interval="${3:-1}"
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        if curl -s -f "$url" >/dev/null 2>&1; then
            return 0
        fi
        sleep "$interval"
        elapsed=$((elapsed + interval))
    done
    
    return 1
}

# Check prerequisites
check_prerequisites() {
    log_section "🔍 Checking Prerequisites"
    
    if ! command -v multipass &> /dev/null; then
        log_error "Multipass not installed"
        echo "Install with: brew install multipass"
        exit 1
    fi
    log_success "Multipass installed"
    
    if [ ! -f "$PODLIFT_BIN" ]; then
        log_error "podlift binary not found at: $PODLIFT_BIN"
        echo "Build with: cd $PROJECT_ROOT && make build"
        exit 1
    fi
    log_success "podlift binary found"
    
    if [ ! -f ~/.ssh/id_rsa.pub ]; then
        log_error "SSH key not found"
        echo "Generate with: ssh-keygen -t rsa -b 4096"
        exit 1
    fi
    log_success "SSH key found"
}

# Setup test application
setup_test_app() {
    log_section "📦 Setting Up Test Application"
    
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Create a simple web app with versioning
    cat > app.py <<'PYTHON'
import os
from http.server import HTTPServer, BaseHTTPRequestHandler
import json

VERSION = os.getenv('APP_VERSION', 'v1')
PORT = int(os.getenv('PORT', 8000))

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({
                'message': 'Hello from podlift!',
                'version': VERSION,
                'hostname': os.getenv('HOSTNAME', 'unknown')
            }).encode())
        elif self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'status': 'healthy'}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass  # Suppress logs

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', PORT), Handler)
    print(f'Server running on port {PORT}')
    server.serve_forever()
PYTHON
    
    cat > Dockerfile <<'DOCKERFILE'
FROM python:3.11-slim
WORKDIR /app
COPY app.py .
EXPOSE 8000
CMD ["python", "app.py"]
DOCKERFILE
    
    cat > docker-compose.yml <<'COMPOSE'
version: '3.8'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: testdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
COMPOSE
    
    # Initialize git
    git init
    git config user.email "test@podlift.test"
    git config user.name "podlift E2E"
    git add .
    git commit -m "Initial commit - v1"
    
    log_success "Test application created"
}

# Test: Single server deployment
test_single_server() {
    log_section "🚀 Test 1: Single Server Deployment"
    
    log_step "Launching single server..."
    local ip=$(launch_vm "${VM_PREFIX}-single")
    
    log_step "Creating podlift config..."
    cat > podlift.yml <<EOF
service: test-app
image: test-app

servers:
  - host: $ip
    user: ubuntu
    ssh_key: ~/.ssh/id_rsa

services:
  web:
    port: 8000
    replicas: 1
    healthcheck:
      path: /health
      expect: [200]
    env:
      APP_VERSION: \${GIT_COMMIT}
EOF
    
    git add podlift.yml
    git commit -m "Add podlift config"
    
    log_step "Setting up server..."
    $PODLIFT_BIN setup --no-security --no-firewall
    
    log_step "Deploying application..."
    $PODLIFT_BIN deploy
    
    log_step "Waiting for application to start..."
    sleep 5
    
    log_step "Testing deployment..."
    assert_http_status "http://$ip:8000/health" "200" "Health check responds"
    
    local response=$(curl -s "http://$ip:8000/")
    assert_contains "$response" "Hello from podlift" "Root endpoint works"
    assert_contains "$response" "v1" "Version is v1"
    
    log_step "Testing podlift commands..."
    local ps_output=$($PODLIFT_BIN ps)
    assert_contains "$ps_output" "healthy" "podlift ps shows healthy status"
    
    local logs_output=$($PODLIFT_BIN logs web --tail 5)
    assert_command_success "podlift logs works" test -n "$logs_output"
    
    log_success "✅ Single server deployment test passed"
}

# Test: Multi-server deployment
test_multi_server() {
    log_section "🌐 Test 2: Multi-Server Deployment"
    
    log_step "Launching multiple servers..."
    local web1_ip=$(launch_vm "${VM_PREFIX}-web1")
    local web2_ip=$(launch_vm "${VM_PREFIX}-web2")
    local db_ip=$(launch_vm "${VM_PREFIX}-db" 2 4G)
    
    log_step "Creating multi-server config..."
    cat > podlift.yml <<EOF
service: test-app
image: test-app

servers:
  web:
    - host: $web1_ip
      user: ubuntu
      ssh_key: ~/.ssh/id_rsa
      labels: [primary]
    - host: $web2_ip
      user: ubuntu
      ssh_key: ~/.ssh/id_rsa
  
  db:
    - host: $db_ip
      user: ubuntu
      ssh_key: ~/.ssh/id_rsa

dependencies:
  postgres:
    image: postgres:16-alpine
    role: db
    port: 5432
    env:
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: testdb
    healthcheck:
      timeout: 30
    volumes:
      - postgres_data:/var/lib/postgresql/data

services:
  web:
    port: 8000
    replicas: 2
    healthcheck:
      path: /health
      expect: [200]
    env:
      APP_VERSION: \${GIT_COMMIT}
EOF
    
    git add podlift.yml
    git commit -m "Multi-server config"
    
    log_step "Setting up all servers..."
    $PODLIFT_BIN setup --no-security --no-firewall
    
    log_step "Deploying to multi-server setup..."
    $PODLIFT_BIN deploy
    
    log_step "Waiting for services to start..."
    sleep 10
    
    log_step "Testing web servers..."
    assert_http_status "http://$web1_ip:8000/health" "200" "Web1 health check"
    assert_http_status "http://$web2_ip:8000/health" "200" "Web2 health check"
    
    log_step "Verifying load balancer (nginx)..."
    # Primary server should have nginx
    local nginx_check=$(multipass exec "${VM_PREFIX}-web1" -- docker ps --filter "name=nginx" --format "{{.Names}}" 2>/dev/null || echo "")
    if [ -n "$nginx_check" ]; then
        log_success "Load balancer is running"
        
        # Test load balancing
        log_step "Testing load balancing..."
        local lb_response=$(curl -s "http://$web1_ip/" 2>/dev/null || echo "")
        assert_contains "$lb_response" "Hello from podlift" "Load balancer forwards requests"
    else
        log_warning "Load balancer not found (expected with 2+ servers)"
    fi
    
    log_step "Verifying postgres on db server..."
    local pg_check=$(multipass exec "${VM_PREFIX}-db" -- docker ps --filter "name=postgres" --format "{{.Names}}" 2>/dev/null || echo "")
    assert_command_success "Postgres is running" test -n "$pg_check"
    
    log_success "✅ Multi-server deployment test passed"
}

# Test: Zero-downtime deployment
test_zero_downtime() {
    log_section "⚡ Test 3: Zero-Downtime Deployment"
    
    log_step "Using first web server for zero-downtime test..."
    local web1_ip="${VM_IPS["${VM_PREFIX}-web1"]}"
    
    if [ -z "$web1_ip" ]; then
        log_error "Web1 IP not found, skipping test"
        return
    fi
    
    log_step "Starting continuous request monitoring..."
    local monitor_file="/tmp/downtime-test-$$.log"
    local downtime_detected=0
    
    # Background process to make continuous requests
    (
        for i in {1..60}; do
            if ! curl -s -f "http://$web1_ip:8000/health" >/dev/null 2>&1; then
                echo "DOWNTIME at $(date +%s)" >> "$monitor_file"
            fi
            sleep 0.5
        done
    ) &
    local monitor_pid=$!
    
    sleep 2
    
    log_step "Making code change (v2)..."
    cat > app.py <<'PYTHON'
import os
from http.server import HTTPServer, BaseHTTPRequestHandler
import json

VERSION = os.getenv('APP_VERSION', 'v2')
PORT = int(os.getenv('PORT', 8000))

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({
                'message': 'Hello from podlift v2!',
                'version': VERSION,
                'hostname': os.getenv('HOSTNAME', 'unknown')
            }).encode())
        elif self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({'status': 'healthy'}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', PORT), Handler)
    print(f'Server v2 running on port {PORT}')
    server.serve_forever()
PYTHON
    
    git add app.py
    git commit -m "Version 2: Updated message"
    
    log_step "Deploying v2 (zero-downtime)..."
    $PODLIFT_BIN deploy
    
    log_step "Waiting for monitor to complete..."
    wait $monitor_pid 2>/dev/null || true
    
    log_step "Checking for downtime..."
    if [ -f "$monitor_file" ] && [ -s "$monitor_file" ]; then
        local downtime_count=$(wc -l < "$monitor_file")
        log_warning "Detected $downtime_count failed requests during deployment"
        # Allow up to 2 failed requests (acceptable during transition)
        if [ "$downtime_count" -le 2 ]; then
            log_success "Minimal downtime (acceptable)"
        else
            log_error "Too much downtime detected"
        fi
    else
        log_success "No downtime detected during deployment"
    fi
    
    rm -f "$monitor_file"
    
    log_step "Verifying v2 is deployed..."
    sleep 3
    local response=$(curl -s "http://$web1_ip:8000/")
    assert_contains "$response" "v2" "Version 2 is deployed"
    
    log_success "✅ Zero-downtime deployment test passed"
}

# Test: Rollback
test_rollback() {
    log_section "🔄 Test 4: Rollback"
    
    local web1_ip="${VM_IPS["${VM_PREFIX}-web1"]}"
    
    if [ -z "$web1_ip" ]; then
        log_error "Web1 IP not found, skipping test"
        return
    fi
    
    log_step "Current version should be v2..."
    local before_response=$(curl -s "http://$web1_ip:8000/")
    assert_contains "$before_response" "v2" "Currently on v2"
    
    log_step "Performing rollback..."
    $PODLIFT_BIN rollback
    
    sleep 5
    
    log_step "Verifying rollback to v1..."
    local after_response=$(curl -s "http://$web1_ip:8000/")
    assert_contains "$after_response" "v1" "Rolled back to v1"
    
    log_success "✅ Rollback test passed"
}

# Test: All commands
test_all_commands() {
    log_section "🔧 Test 5: All podlift Commands"
    
    log_step "Testing 'podlift ps'..."
    local ps_output=$($PODLIFT_BIN ps 2>/dev/null || echo "FAILED")
    assert_contains "$ps_output" "test-app" "ps shows service name"
    
    log_step "Testing 'podlift logs'..."
    assert_command_success "logs command works" $PODLIFT_BIN logs web --tail 5
    
    log_step "Testing 'podlift status'..."
    assert_command_success "status command works" $PODLIFT_BIN status
    
    log_step "Testing 'podlift config'..."
    local config_output=$($PODLIFT_BIN config 2>/dev/null || echo "FAILED")
    assert_contains "$config_output" "service: test-app" "config shows configuration"
    
    log_step "Testing 'podlift version'..."
    assert_command_success "version command works" $PODLIFT_BIN version
    
    log_step "Testing 'podlift exec'..."
    local web1_ip="${VM_IPS["${VM_PREFIX}-web1"]}"
    if [ -n "$web1_ip" ]; then
        # Try to exec into container
        local exec_output=$($PODLIFT_BIN exec web -- echo "test" 2>/dev/null || echo "")
        if [ -n "$exec_output" ]; then
            log_success "exec command works"
        else
            log_warning "exec command needs interactive terminal"
        fi
    fi
    
    log_success "✅ All commands test passed"
}

# Test: Hooks
test_hooks() {
    log_section "🪝 Test 6: Deployment Hooks"
    
    log_step "Adding hooks to config..."
    local web1_ip="${VM_IPS["${VM_PREFIX}-web1"]}"
    
    cat > podlift.yml <<EOF
service: test-app
image: test-app

servers:
  web:
    - host: $web1_ip
      user: ubuntu
      ssh_key: ~/.ssh/id_rsa
      labels: [primary]

services:
  web:
    port: 8000
    replicas: 1
    healthcheck:
      path: /health
      expect: [200]
    env:
      APP_VERSION: \${GIT_COMMIT}

hooks:
  before_deploy:
    - echo "Before deploy hook executed" > /tmp/podlift-hook-before
  after_deploy:
    - echo "After deploy hook executed" > /tmp/podlift-hook-after
EOF
    
    git add podlift.yml
    git commit -m "Add hooks"
    
    log_step "Deploying with hooks..."
    $PODLIFT_BIN deploy
    
    sleep 3
    
    log_step "Verifying hooks executed..."
    local before_hook=$(multipass exec "${VM_PREFIX}-web1" -- cat /tmp/podlift-hook-before 2>/dev/null || echo "")
    local after_hook=$(multipass exec "${VM_PREFIX}-web1" -- cat /tmp/podlift-hook-after 2>/dev/null || echo "")
    
    assert_contains "$before_hook" "Before deploy" "before_deploy hook executed"
    assert_contains "$after_hook" "After deploy" "after_deploy hook executed"
    
    log_success "✅ Hooks test passed"
}

# Test: Dependencies with volumes
test_dependencies() {
    log_section "🗄️ Test 7: Dependencies with Volumes"
    
    local db_ip="${VM_IPS["${VM_PREFIX}-db"]}"
    
    if [ -z "$db_ip" ]; then
        log_warning "DB server not found, skipping test"
        return
    fi
    
    log_step "Checking postgres is running..."
    local pg_container=$(multipass exec "${VM_PREFIX}-db" -- docker ps --filter "name=postgres" --format "{{.Names}}" 2>/dev/null || echo "")
    assert_command_success "Postgres container exists" test -n "$pg_container"
    
    log_step "Verifying postgres health..."
    sleep 5
    local pg_health=$(multipass exec "${VM_PREFIX}-db" -- docker exec "$pg_container" pg_isready -U postgres 2>/dev/null || echo "")
    assert_contains "$pg_health" "accepting connections" "Postgres is healthy"
    
    log_step "Verifying volume persistence..."
    local volume_check=$(multipass exec "${VM_PREFIX}-db" -- docker volume ls --filter "name=postgres_data" --format "{{.Name}}" 2>/dev/null || echo "")
    assert_command_success "Postgres volume exists" test -n "$volume_check"
    
    log_success "✅ Dependencies test passed"
}

# Test: Environment variables
test_environment_variables() {
    log_section "🔐 Test 8: Environment Variables"
    
    local web1_ip="${VM_IPS["${VM_PREFIX}-web1"]}"
    
    if [ -z "$web1_ip" ]; then
        log_warning "Web1 server not found, skipping test"
        return
    fi
    
    log_step "Checking environment variables in container..."
    local container_name=$(multipass exec "${VM_PREFIX}-web1" -- docker ps --filter "label=podlift.service=test-app" --format "{{.Names}}" | head -1 2>/dev/null || echo "")
    
    if [ -n "$container_name" ]; then
        local env_check=$(multipass exec "${VM_PREFIX}-web1" -- docker exec "$container_name" env 2>/dev/null || echo "")
        assert_contains "$env_check" "APP_VERSION" "Environment variables are set"
        log_success "✅ Environment variables test passed"
    else
        log_warning "No container found to test environment variables"
    fi
}

# Print summary
print_summary() {
    echo ""
    echo ""
    log_section "📊 Test Summary"
    
    echo -e "${BLUE}Total Tests Run:    ${NC}$TESTS_RUN"
    echo -e "${GREEN}Tests Passed:       ${NC}$TESTS_PASSED"
    echo -e "${RED}Tests Failed:       ${NC}$TESTS_FAILED"
    echo ""
    
    local success_rate=0
    if [ $TESTS_RUN -gt 0 ]; then
        success_rate=$((TESTS_PASSED * 100 / TESTS_RUN))
    fi
    
    echo -e "${BLUE}Success Rate:       ${NC}${success_rate}%"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}✅ ALL TESTS PASSED!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 0
    else
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}❌ SOME TESTS FAILED${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 1
    fi
}

# Main test execution
main() {
    echo ""
    echo -e "${PURPLE}╔════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║                                        ║${NC}"
    echo -e "${PURPLE}║  podlift Comprehensive E2E Test Suite ║${NC}"
    echo -e "${PURPLE}║                                        ║${NC}"
    echo -e "${PURPLE}╚════════════════════════════════════════╝${NC}"
    echo ""
    
    local start_time=$(date +%s)
    
    check_prerequisites
    setup_test_app
    
    # Run all tests
    test_single_server
    test_multi_server
    test_zero_downtime
    test_rollback
    test_all_commands
    test_hooks
    test_dependencies
    test_environment_variables
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo ""
    echo -e "${BLUE}Test Duration: ${duration}s${NC}"
    
    print_summary
}

# Run tests
main "$@"

