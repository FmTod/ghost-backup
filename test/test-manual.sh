#!/bin/bash
# Manual test script - simpler approach without systemd in Docker

set -e

# Save project root directory
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Change to project root
cd "$PROJECT_ROOT"

echo "=== Building ghost-backup binary ==="
go build -o ghost-backup .

echo ""
echo "=== Testing service installation locally (requires systemd user instance) ==="
echo "This will test if the WantedBy directive is correctly set."
echo ""

# Create a temporary directory for testing
TEST_DIR=$(mktemp -d)
TEST_REPO="$TEST_DIR/test-repo"

echo "Test directory: $TEST_DIR"

# Create a test repository
mkdir -p "$TEST_REPO"
cd "$TEST_REPO"
git init
git config user.email "test@example.com"
git config user.name "Test User"
echo "Test file" > README.md
git add README.md
git commit -m "Initial commit"

# Return to project root
cd "$PROJECT_ROOT"

echo ""
echo "=== Installing service (use --skip-service to avoid actually installing) ==="
if [ -f "./ghost-backup" ]; then
    ./ghost-backup service install || true
else
    echo "Error: ghost-backup binary not found"
    exit 1
fi

echo ""
echo "=== Checking generated systemd unit file ==="
UNIT_FILE="$HOME/.config/systemd/user/ghost-backup.service"

if [ -f "$UNIT_FILE" ]; then
    echo "Unit file location: $UNIT_FILE"
    echo ""
    echo "--- Full unit file ---"
    cat "$UNIT_FILE"
    echo ""
    echo "--- Checking WantedBy directive ---"
    grep "WantedBy=" "$UNIT_FILE" || echo "No WantedBy directive found!"
else
    echo "Unit file not found at $UNIT_FILE"
fi

echo ""
echo "=== Cleanup ===="
echo "To uninstall: ./ghost-backup service uninstall"
echo "Test repository: $TEST_REPO"
