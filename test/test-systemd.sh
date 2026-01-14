#!/bin/bash
# Helper script to test ghost-backup systemd service in Docker

set -e

# Change to the test directory
cd "$(dirname "$0")"

echo "Building Docker image..."
docker compose -f docker-compose.test.yml build

echo "Starting container with systemd..."
docker compose -f docker-compose.test.yml up -d

echo "Waiting for systemd to initialize..."
docker compose -f docker-compose.test.yml exec ghost-backup-test bash -c "while ! systemctl is-system-running --wait 2>/dev/null; do sleep 1; done" || true
sleep 2

echo ""
echo "=== Testing ghost-backup service ==="
echo ""

# Execute commands as testuser
echo "1. Installing service..."
docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c "ghost-backup service install"

echo ""
echo "2. Checking generated systemd unit file..."
docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c "systemctl --user cat ghost-backup.service"

echo ""
echo "3. Checking for WantedBy target..."
docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c "systemctl --user cat ghost-backup.service | grep WantedBy"

echo ""
echo "4. Enabling service..."
docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c "systemctl --user enable ghost-backup.service"

echo ""
echo "5. Starting service..."
docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c "systemctl --user start ghost-backup.service"

echo ""
echo "6. Checking service status..."
docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c "systemctl --user status ghost-backup.service"

echo ""
echo "=== Test complete ==="
echo ""
echo "To access the container interactively:"
echo "  docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser"
echo ""
echo "To view logs:"
echo "  docker compose -f docker-compose.test.yml exec ghost-backup-test su - testuser -c 'journalctl --user -u ghost-backup.service'"
echo ""
echo "To stop and remove:"
echo "  docker compose -f docker-compose.test.yml down -v"
