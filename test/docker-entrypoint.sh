#!/bin/bash
set -e

# Enable lingering for testuser to allow user services
loginctl enable-linger testuser || true

# Start systemd
exec /lib/systemd/systemd
