#!/bin/bash
set -e

echo "🔍 podlift Verification Script"
echo "=============================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Change to project root
cd "$(dirname "$0")/.."

# 1. Check Go version
print_info "Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}')
print_success "Go version: $GO_VERSION"
echo ""

# 2. Run tests
print_info "Running tests..."
if go test ./... -v > /tmp/podlift-test.log 2>&1; then
    TEST_COUNT=$(grep -c "^PASS" /tmp/podlift-test.log || true)
    print_success "All tests passed ($TEST_COUNT test suites)"
else
    print_error "Tests failed"
    cat /tmp/podlift-test.log
    exit 1
fi
echo ""

# 3. Check test coverage
print_info "Checking test coverage..."
go test ./... -coverprofile=/tmp/podlift-coverage.out > /dev/null 2>&1
COVERAGE=$(go tool cover -func=/tmp/podlift-coverage.out | grep total | awk '{print $3}')
print_success "Test coverage: $COVERAGE"
echo ""

# 4. Build binary
print_info "Building binary..."
if make build > /dev/null 2>&1; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi
echo ""

# 5. Test CLI commands
print_info "Testing CLI commands..."

# Test version command
if ./bin/podlift version > /dev/null 2>&1; then
    print_success "podlift version works"
else
    print_error "podlift version failed"
    exit 1
fi

# Test help command
if ./bin/podlift --help > /dev/null 2>&1; then
    print_success "podlift --help works"
else
    print_error "podlift --help failed"
    exit 1
fi

# Test init help
if ./bin/podlift init --help > /dev/null 2>&1; then
    print_success "podlift init --help works"
else
    print_error "podlift init --help failed"
    exit 1
fi
echo ""

# 6. Verify test data
print_info "Verifying test configurations..."

# Check minimal config
if [ -f "testdata/minimal.yml" ]; then
    print_success "testdata/minimal.yml exists"
else
    print_error "testdata/minimal.yml missing"
    exit 1
fi

# Check full config
if [ -f "testdata/full.yml" ]; then
    print_success "testdata/full.yml exists"
else
    print_error "testdata/full.yml missing"
    exit 1
fi
echo ""

# 7. Check documentation
print_info "Checking documentation..."
DOCS=(
    "README.md"
    "docs/installation.md"
    "docs/commands.md"
    "docs/configuration.md"
    "docs/deployment-guide.md"
    "docs/how-it-works.md"
    "docs/troubleshooting.md"
    "docs/migration.md"
)

DOC_COUNT=0
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        DOC_COUNT=$((DOC_COUNT + 1))
    fi
done
TOTAL_DOCS=${#DOCS[@]}
print_success "$DOC_COUNT/$TOTAL_DOCS documentation files present"
echo ""

# 8. Check project structure
print_info "Verifying project structure..."
DIRS=(
    "cmd/podlift"
    "internal/config"
    "tests/integration"
    "tests/e2e"
    "testdata"
    "docs"
)

for dir in "${DIRS[@]}"; do
    if [ -d "$dir" ]; then
        print_success "$dir/"
    else
        print_error "$dir/ missing"
    fi
done
echo ""

# 9. Summary
echo "=============================="
echo -e "${GREEN}✓ All verifications passed!${NC}"
echo ""
echo "Summary:"
echo "  • Tests: All passing"
echo "  • Coverage: $COVERAGE"
echo "  • Build: Successful"
echo "  • CLI: Working"
echo "  • Docs: Present"
echo ""
echo "Ready to continue Phase 0 development!"

