# Test Files

This directory contains test scripts and configurations for ghost-backup.

## Files

- **test-manual.sh** - Manual test script that builds and installs the service locally to verify systemd unit configuration
- **test-systemd.sh** - Docker-based systemd test (experimental, systemd in Docker is complex)
- **docker-compose.test.yml** - Docker Compose configuration for systemd testing
- **Dockerfile.test** - Multi-stage build with systemd support
- **docker-entrypoint.sh** - Docker entrypoint script for systemd container

## Usage

### Manual Testing (Recommended)

Test the service installation locally:

```bash
./test-manual.sh
```

This will:
1. Build the ghost-backup binary
2. Create a test repository
3. Install the service
4. Display the generated systemd unit file
5. Verify the `WantedBy=` directive is set to `default.target`

**Cleanup:**
```bash
../ghost-backup service uninstall
```

### Docker Testing (Experimental)

Test with systemd in Docker (requires privileged mode):

```bash
./test-systemd.sh
```

Note: Systemd in Docker is complex and may not work on all systems.

**Cleanup:**
```bash
docker compose -f docker-compose.test.yml down -v
```
