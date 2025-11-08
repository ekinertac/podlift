#!/bin/bash
# Interactive test runner for podlift E2E tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
NC='\033[0m'

echo ""
echo -e "${PURPLE}╔════════════════════════════════════════╗${NC}"
echo -e "${PURPLE}║  podlift E2E Test Runner              ║${NC}"
echo -e "${PURPLE}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Choose a test to run:${NC}"
echo ""
echo "  1) Comprehensive Test Suite (all features)"
echo "     • Single server, multi-server, load balancing"
echo "     • Zero-downtime, rollback, all commands"
echo "     • Hooks, dependencies, volumes"
echo "     • Duration: ~10-15 minutes"
echo ""
echo "  2) Quick FastAPI Test (single server)"
echo "     • Basic deployment and health checks"
echo "     • Duration: ~3-5 minutes"
echo ""
echo "  3) Custom: Single Server Only"
echo "     • Fast single server test"
echo "     • Duration: ~2-3 minutes"
echo ""
echo "  4) Custom: Multi-Server Only"
echo "     • Multi-server deployment test"
echo "     • Duration: ~5-7 minutes"
echo ""
echo "  5) Custom: Zero-Downtime Test"
echo "     • Test zero-downtime deployments"
echo "     • Duration: ~3-4 minutes"
echo ""

read -p "Enter choice (1-5): " choice

case $choice in
    1)
        echo -e "${GREEN}Running comprehensive test suite...${NC}"
        exec "$SCRIPT_DIR/comprehensive-test.sh"
        ;;
    2)
        echo -e "${GREEN}Running FastAPI test...${NC}"
        exec "$SCRIPT_DIR/test-fastapi.sh"
        ;;
    3)
        echo -e "${YELLOW}Custom single-server test not yet implemented${NC}"
        echo "Use: ./comprehensive-test.sh (it includes single-server tests)"
        exit 1
        ;;
    4)
        echo -e "${YELLOW}Custom multi-server test not yet implemented${NC}"
        echo "Use: ./comprehensive-test.sh (it includes multi-server tests)"
        exit 1
        ;;
    5)
        echo -e "${YELLOW}Custom zero-downtime test not yet implemented${NC}"
        echo "Use: ./comprehensive-test.sh (it includes zero-downtime tests)"
        exit 1
        ;;
    *)
        echo -e "${YELLOW}Invalid choice${NC}"
        exit 1
        ;;
esac

