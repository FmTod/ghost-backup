# Ghost Backup

A production-ready, multi-platform CLI tool that provides automated git backup functionality for multiple repositories. Ghost Backup runs as a background service, continuously monitoring your repositories and pushing "invisible" git snapshots (work-in-progress) to a backup server.

## Features

- **Multi-Repository Support**: Monitor multiple git repositories simultaneously
- **Per-Repository Configuration**: Each repository has its own settings (interval, security options)
- **Hot Reloading**: Change settings without restarting the service
- **Secret Scanning**: Optional integration with gitleaks to prevent backing up sensitive data
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Namespace Isolation**: Backups are organized by user email and branch name
- **Background Service**: Runs as a system service with minimal overhead

## Installation

### Prerequisites

- Go 1.21 or higher
- Git installed and configured
- (Optional) [gitleaks](https://github.com/gitleaks/gitleaks) for secret scanning

### Building from Source

```bash
git clone https://github.com/neoscode/ghost-backup.git
cd ghost-backup
go build -o ghost-backup
sudo mv ghost-backup /usr/local/bin/
```

## Quick Start

### 1. Initialize Ghost Backup for a Repository

Navigate to your git repository and run:

```bash
cd /path/to/your/repo
ghost-backup init
```

This will:
- Create `.ghost-backup.json` with default settings
- Add the repository to the global registry
- Install and start the system service (may require sudo)

### 2. Custom Initialization

Specify custom settings during initialization:

```bash
ghost-backup init --interval 300 --scan-secrets=false
```

Options:
- `--path, -p`: Path to the repository (default: current directory)
- `--interval, -i`: Backup interval in seconds (default: 60)
- `--scan-secrets, -s`: Enable secret scanning with gitleaks (default: true)

### 3. View Backups

List all available backups for the current repository:

```bash
cd /path/to/your/repo
ghost-backup list
```

### 4. Restore a Backup

Restore a specific backup by hash:

```bash
ghost-backup restore <hash>
```

Or use the cherry-pick method:

```bash
ghost-backup restore <hash> --method cherry-pick
```

### 5. Generate Pruning Workflow (Optional)

Automatically clean up old backups using GitHub Actions:

```bash
cd /path/to/your/repo
ghost-backup workflow
```

This creates a GitHub Actions workflow that periodically deletes old backup refs. Customize with:

```bash
ghost-backup workflow --cron "0 2 * * *" --retention 14
```

Options:
- `--cron, -c`: Cron schedule (default: "0 2 * * 0" – weekly on Sunday at 2am)
- `--retention, -r`: Days to keep backups (default: 30)

### 6. Validate Configuration

Check your repository's configuration and setup:

```bash
cd /path/to/your/repo
ghost-backup check
```

This validates:
- Git repository status
- Configuration file existence and validity
- Global registry inclusion
- Git user email and remote configuration
- Service status
- Gitleaks availability (for secret scanning)

The command provides detailed diagnostics and suggestions for fixing any issues.

## Configuration

### Global Registry

The global registry is stored at `~/.config/ghost-backup/registry.json` and contains the list of all monitored repositories.

```json
{
  "repositories": [
    "/home/user/project1",
    "/home/user/project2"
  ]
}
```

### Local Repository Configuration

Each repository has its own configuration file at `.ghost-backup.json`:

```json
{
  "interval": 60,
  "scan_secrets": true
}
```

**To change settings**, edit this file. The service will automatically detect the changes on the next backup cycle.

**Note**: You can commit `.ghost-backup.json` to your repository to share backup settings with your team, or add it to `.gitignore` to keep settings local.

### Configuration Options

- **interval**: Backup interval in seconds (minimum recommended: 60)
- **scan_secrets**: Whether to scan for secrets using gitleaks before backing up

## Service Management

### Check Service Status

```bash
ghost-backup service status
```

### Start/Stop/Restart Service

```bash
ghost-backup service start
ghost-backup service stop
ghost-backup service restart
```

### Install/Uninstall Service

```bash
sudo ghost-backup service install
sudo ghost-backup service uninstall
```

### Run Service in Foreground (for debugging)

```bash
ghost-backup service run
```

## Viewing Logs

Logs are stored at `~/.config/ghost-backup/ghost-backup.log`.

To view logs in real-time:

```bash
tail -f ~/.config/ghost-backup/ghost-backup.log
```

## How It Works

### Backup Process

1. **Change Detection**: The worker checks if there are uncommitted changes in the repository
2. **Snapshot Creation**: Creates a git stash without modifying the working directory
3. **Secret Scanning** (if enabled): Scans the diff for secrets using gitleaks with a 60-second timeout
4. **Push to Remote**: Pushes the snapshot to `refs/backups/<user_email>/<branch_name>`

### Architecture

- **Single Global Service**: One background service manages all repositories
- **Worker Goroutines**: Each repository runs in its own goroutine
- **Hot Reloading**: Workers periodically check for configuration changes
- **Error Isolation**: If one repository fails, others continue running

### Backup Reference Namespace

Backups are organized using the following reference pattern:

```
refs/backups/<user_email>/<branch_name>
```

Example:
```
refs/backups/user_at_example.com/main
refs/backups/user_at_example.com/feature_new-api
```

## Uninstalling

To remove a repository from monitoring:

```bash
cd /path/to/your/repo
ghost-backup uninstall
```

This will:
- Remove the repository from the global registry
- Delete `.ghost-backup.json`
- Restart the service to stop monitoring

## Troubleshooting

### Quick Diagnostics

Before troubleshooting specific issues, run the diagnostic command:

```bash
ghost-backup check
```

This will identify common issues and provide specific fix suggestions.

### Service Won't Start

**Problem**: Service fails to start, possibly due to permissions.

**Solution**: System services typically require root privileges. Try:

```bash
sudo ghost-backup init
```

Or manually install and start the service:

```bash
sudo ghost-backup service install
sudo ghost-backup service start
```

### Gitleaks Timeout

**Problem**: Gitleaks scan times out on large diffs.

**Solution**: The timeout is set to 60 seconds. For repositories with large changes, consider disabling secret scanning:

```bash
# Edit .ghost-backup.json
{
  "interval": 60,
  "scan_secrets": false
}
```

### Repository Not Being Backed Up

**Problem**: No backups are being created for a repository.

**Solution**: 
1. Check service status: `ghost-backup service status`
2. View logs: `tail -f ~/.config/ghost-backup/ghost-backup.log`
3. Verify repository is in registry: `cat ~/.config/ghost-backup/registry.json`
4. Ensure remote is configured: `git remote -v`

### Secrets Detected

**Problem**: Backup is blocked because secrets were detected.

**Solution**: Review the log output to see what was detected. Remove the secrets from your changes or disable secret scanning if it's a false positive:

```json
{
  "interval": 60,
  "scan_secrets": false
}
```

## Security Considerations

### Secret Scanning

When `scan_secrets` is enabled, ghost-backup uses gitleaks to scan diffs before backing up. This prevents accidentally pushing sensitive information like:

- API keys
- Passwords
- Private keys
- Tokens
- Credentials

**Important**: Secret scanning is only as good as gitleaks' detection rules. Always review your code before committing.

### Backup Storage

Backups are pushed to the configured git remote. Ensure your remote is:
- Secured with proper authentication
- Using HTTPS or SSH with key-based authentication
- Hosted on a trusted server

### Network Considerations

Ghost Backup pushes to git remotes over the network. Ensure:
- Your network connection is secure
- The remote server is trusted
- Credentials are properly managed (SSH keys, credential helpers)

## Advanced Usage

### Automated Backup Pruning

The generated GitHub Actions workflow provides automated cleanup of old backups:

**Features:**
- Runs on a schedule (customizable with cron)
- Deletes backup refs older than retention period
- Can be manually triggered from the GitHub Actions tab
- Provides a summary of deletions

**Example schedules:**
- `"0 2 * * 0"` - Weekly on Sunday at 2am (default)
- `"0 2 * * *"` - Daily at 2am
- `"0 0 1 * *"` - Monthly on the 1st
- `"0 */6 * * *"` - Every 6 hours

**Manual trigger:**
You can also run the workflow manually from the GitHub Actions tab and specify a custom retention period.

### Multiple Remotes

If your repository has multiple remotes, ghost-backup will use the first one found (preferring "origin").

### Custom Intervals per Repository

Different repositories can have different backup intervals:

```bash
# Project with frequent changes - backup every 30 seconds
cd ~/project1
echo '{"interval": 30, "scan_secrets": true}' > .ghost-backup.json

# Stable project - backup every 5 minutes
cd ~/project2
echo '{"interval": 300, "scan_secrets": true}' > .ghost-backup.json
```

### Restoring to a Different Branch

To restore a backup to a different branch:

```bash
git checkout other-branch
ghost-backup restore <hash>
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License – see LICENSE file for details

## Author

- [Victor R (viicslen)](https://github.com/viicslen)

## Acknowledgments

- [kardianos/service](https://github.com/kardianos/service) - Cross-platform service management
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [gitleaks](https://github.com/gitleaks/gitleaks) - Secret scanning
