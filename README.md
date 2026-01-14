# Ghost Backup

A production-ready, multi-platform CLI tool that provides automated git backup functionality for multiple repositories. Ghost Backup runs as a background service, continuously monitoring your repositories and pushing "invisible" git snapshots (work-in-progress) to a backup server.

## Features

- **Multi-Repository Support**: Monitor multiple git repositories simultaneously
- **Per-Repository Configuration**: Each repository has its own settings (interval, security options)
- **Hot Reloading**: Change settings without restarting the service
- **Secret Scanning**: Optional integration with gitleaks to prevent backing up sensitive data
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Namespace Isolation**: Backups are organized by user email and branch name
- **Background Service**: Runs as a user service with minimal overhead
- **Git Worktree Support**: Full support for git worktrees - backup and monitor worktree directories

## Installation

### Prerequisites

- Go 1.21 or higher
- Git installed and configured
- (Optional) [gitleaks](https://github.com/gitleaks/gitleaks) for secret scanning

### Quick Install (Linux/macOS/WSL)

Install the latest release with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/FmTod/ghost-backup/main/install.sh | bash
```

This script will:

- Detect your operating system and architecture
- Download the appropriate binary from the latest GitHub release
- Install it to `/usr/local/bin/ghost-backup`
- Verify the installation

To install to a custom location:

```bash
curl -fsSL https://raw.githubusercontent.com/FmTod/ghost-backup/main/install.sh | bash -s -- --prefix ~/.local
```

<details>
<summary><h3>Using Nix</h3></summary>

#### Run without installing

```bash
nix run github:FmTod/ghost-backup -- --help
```

#### Install to your profile

```bash
nix profile install github:FmTod/ghost-backup
```

#### NixOS Module

Add to your NixOS configuration:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    ghost-backup.url = "github:FmTod/ghost-backup";
  };

  outputs = { nixpkgs, ghost-backup, ... }: {
    nixosConfigurations.your-hostname = nixpkgs.lib.nixosSystem {
      modules = [
        ghost-backup.nixosModules.default
        {
          services.ghost-backup.enable = true;
        }
      ];
    };
  };
}
```

This configures a systemd user service that runs as your user account.

Manage the service with:

```bash
# Check status
systemctl --user status ghost-backup

# View logs
journalctl --user -u ghost-backup -f

# Restart
systemctl --user restart ghost-backup
```

#### Development Shell

Enter a development environment with all dependencies:

```bash
nix develop github:FmTod/ghost-backup
```

Or locally:

```bash
cd ghost-backup
nix develop
```

</details>

<details>
<summary><h3>Manual Installation</h3></summary>

1. Download the latest release for your platform from the [releases page](https://github.com/FmTod/ghost-backup/releases/latest):

   **Linux (amd64)**:

   ```bash
   curl -L -o ghost-backup https://github.com/FmTod/ghost-backup/releases/latest/download/ghost-backup-linux-amd64
   chmod +x ghost-backup
   sudo mv ghost-backup /usr/local/bin/
   ```

   **Linux (arm64)**:

   ```bash
   curl -L -o ghost-backup https://github.com/FmTod/ghost-backup/releases/latest/download/ghost-backup-linux-arm64
   chmod +x ghost-backup
   sudo mv ghost-backup /usr/local/bin/
   ```

   **macOS (Intel)**:

   ```bash
   curl -L -o ghost-backup https://github.com/FmTod/ghost-backup/releases/latest/download/ghost-backup-darwin-amd64
   chmod +x ghost-backup
   sudo mv ghost-backup /usr/local/bin/
   ```

   **macOS (Apple Silicon)**:

   ```bash
   curl -L -o ghost-backup https://github.com/FmTod/ghost-backup/releases/latest/download/ghost-backup-darwin-arm64
   chmod +x ghost-backup
   sudo mv ghost-backup /usr/local/bin/
   ```

   **Windows**:
   Download [ghost-backup-windows-amd64.exe](https://github.com/FmTod/ghost-backup/releases/latest/download/ghost-backup-windows-amd64.exe) and add it to your PATH.

2. Verify the installation:

   ```bash
   ghost-backup --version
   ```

</details>

### Building from Source

```bash
git clone https://github.com/FmTod/ghost-backup.git
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
- Install and start the user service

### 2. Custom Initialization

Specify custom settings during initialization:

```bash
ghost-backup init --interval 5 --scan-secrets=false
```

Options:

- `--path, -p`: Path to the repository (default: current directory)
- `--interval, -i`: Backup interval in minutes (default: 60)
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

### 5. Create a Backup Now

To create a backup immediately without waiting for the scheduled interval:

```bash
cd /path/to/your/repo
ghost-backup backup
```

This will:

- Check for uncommitted changes
- Create a stash if changes exist
- Scan for secrets (if enabled)
- Push the backup to the remote

You can also specify a path:

```bash
ghost-backup backup --path /path/to/repo
```

### 6. Generate Pruning Workflow (Optional)

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

### 7. Validate Configuration

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

### Global Configuration

The global configuration is stored at `~/.config/ghost-backup/config.json` and contains settings that apply to all repositories.

```json
{
  "git_user": "myusername",
  "git_token": "ghp_xxxxxxxxxxxx"
}
```

#### Git Authentication Token

For non-interactive authentication (required when running as a service), you can configure a Git username and personal access token:

```bash
# Set credentials interactively (secure, hidden input)
ghost-backup config set-token

# Set credentials with username and token
ghost-backup config set-token --username myuser --token ghp_xxxxxxxxxxxx

# Set only token (username will default to token for auth)
ghost-backup config set-token --token ghp_xxxxxxxxxxxx

# View configured credentials (token masked)
ghost-backup config get-token

# Clear credentials
ghost-backup config clear-token
```

**Creating a GitHub Personal Access Token:**

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name (e.g., "ghost-backup")
4. Select scopes: `repo` (Full control of private repositories)
5. Generate and copy the token
6. Run `ghost-backup config set-token` and paste the token (username is optional)

**Note:** After setting credentials, restart the service:

```bash
ghost-backup service restart
```

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

- **interval**: Backup interval in minutes (minimum recommended: 1)
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
ghost-backup service install
ghost-backup service uninstall
```

### Run Service in Foreground (for debugging)

```bash
ghost-backup service run
```

## Viewing Logs

### Log Location

Logs are stored at `~/.local/state/ghost-backup/ghost-backup.log` (following XDG Base Directory specification).

**For systemd services (NixOS module)**:

- View with: `journalctl --user -u ghost-backup -f` (recommended)
- Or: `tail -f ~/.local/state/ghost-backup/ghost-backup.log`

**For manual installations**:

- View with: `tail -f ~/.local/state/ghost-backup/ghost-backup.log`

### Viewing Logs in Real-Time

For systemd services (NixOS module):

```bash
journalctl --user -u ghost-backup -f
```

For manual installations:

```bash
tail -f ~/.local/state/ghost-backup/ghost-backup.log
```

## How It Works

### Backup Process

1. **Change Detection**: The worker checks if there are uncommitted changes in the repository
2. **Snapshot Creation**: Creates a git stash without modifying the working directory
3. **Secret Scanning** (if enabled): Scans the diff for secrets using gitleaks with a 60-second timeout
4. **Push to Remote**: Pushes the snapshot to `refs/backups/<user_identifier>/<branch_name>`

### User Identifier System

Ghost Backup uses a flexible user identifier system to organize backups in a way that balances privacy with team visibility.

**Priority order for user identifiers:**

1. **`git_user` from global config** - User-configured identifier (recommended for teams)
2. **Git username** - From `git config user.name`, sanitized for ref compatibility
3. **Sanitized email** - Last resort fallback

**Why this matters:**

- **Team leads can identify backups**: No hashed identifiers - clear usernames like "johndoe"
- **Privacy by default**: Emails are sanitized (e.g., `john_at_example.com`)
- **User control**: Team members can set their own identifier

**Setting your identifier:**

```bash
# Recommended: Set a custom identifier
ghost-backup config set-token --username johndoe --token ghp_xxxx

# Check what identifier will be used
ghost-backup check
```

This ensures consistent, identifiable backup refs like `refs/backups/johndoe/main` that team leads can easily recognize.

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

**Problem**: Service fails to start.

**Solution**: Try running the service in the foreground to see errors:

```bash
ghost-backup service run
```

Or check if the service is properly installed:

```bash
ghost-backup service status
```

If not installed, install it:

```bash
ghost-backup service install
ghost-backup service start
```

### Read-Only File System Error (NixOS)

**Problem**: Service fails with error: `failed to open log file: read-only file system`

**Cause**: This occurs when using the NixOS module with systemd hardening settings that restrict write access.

**Solution**:

1. Ensure you're using the latest version of the ghost-backup flake/module
2. Rebuild your NixOS configuration:

   ```bash
   sudo nixos-rebuild switch
   ```

3. Restart the user service:

   ```bash
   systemctl --user restart ghost-backup
   ```

**Alternative**: If you're not using the NixOS module and installed manually, ensure the config directory is writable:

```bash
mkdir -p ~/.config/ghost-backup
chmod 755 ~/.config/ghost-backup
```

### Git Authentication Failures

**Problem**: Service fails to push backups with authentication errors like:

- "fatal: could not read Username"
- "fatal: could not read Password"
- "Authentication failed"

**Cause**: Git is trying to prompt for credentials in non-interactive mode (service cannot prompt).

**Solution**: Configure Git credentials (username and token):

1. Create a GitHub personal access token (see Configuration section above)
2. Set the credentials:

   ```bash
   ghost-backup config set-token --username myuser --token ghp_xxxx
   # Or interactively:
   ghost-backup config set-token
   ```

3. Restart the service:

   ```bash
   ghost-backup service restart
   ```

**Verify credentials are configured:**

```bash
ghost-backup config get-token
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
2. View logs: `tail -f ~/.local/state/ghost-backup/ghost-backup.log`
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
# Project with frequent changes - backup every minute
cd ~/project1
echo '{"interval": 1, "scan_secrets": true}' > .ghost-backup.json

# Stable project - backup every 5 minutes
cd ~/project2
echo '{"interval": 5, "scan_secrets": true}' > .ghost-backup.json
```

### Restoring to a Different Branch

To restore a backup to a different branch:

```bash
git checkout other-branch
ghost-backup restore <hash>
```

### Working with Git Worktrees

Ghost Backup fully supports [git worktrees](https://git-scm.com/docs/git-worktree), which allow you to have multiple working directories from a single repository. You can initialize and monitor worktrees just like regular repositories:

```bash
# Create a worktree
cd ~/my-project
git worktree add ../my-project-feature feature-branch

# Initialize ghost-backup in the worktree
cd ../my-project-feature
ghost-backup init

# The worktree will be monitored independently
# Backups will be organized by branch: refs/backups/<user>/feature-branch
```

**Key features with worktrees:**

- Each worktree can have its own `.ghost-backup.json` configuration
- Backups are organized by the worktree's current branch
- All git operations (stash, diff, remote) work seamlessly
- Worktrees share the same remote configuration as the main repository

**Example workflow:**

```bash
# Main repository on 'main' branch
cd ~/project
ghost-backup init --interval 60

# Create worktree for feature development
git worktree add ../project-feature feature-branch
cd ../project-feature
ghost-backup init --interval 60  # Same default interval

# Both locations are monitored independently with their own settings
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
