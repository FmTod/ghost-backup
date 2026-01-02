# Nix Usage Guide for Ghost Backup

This document provides detailed instructions for using Ghost Backup with Nix and NixOS.

## Quick Start

### Run the application

```bash
# From the repository
nix run .# -- --help

# From GitHub (without cloning)
nix run github:neoscode/ghost-backup -- --help
```

### Install to your profile

```bash
# From the repository
nix profile install .#

# From GitHub
nix profile install github:neoscode/ghost-backup
```

### Development shell

```bash
# Enter development environment
nix develop

# What you get:
# - Go 1.25
# - gopls (Go language server)
# - gotools and go-tools
# - git
```

## NixOS Module

The flake includes a NixOS module for running Ghost Backup as a systemd service.

### Basic Configuration

Add to your `flake.nix`:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    ghost-backup.url = "github:neoscode/ghost-backup";
  };

  outputs = { nixpkgs, ghost-backup, ... }: {
    nixosConfigurations.your-hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
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

### Advanced Configuration

```nix
{
  services.ghost-backup = {
    enable = true;
    
    # Use a specific package version
    package = inputs.ghost-backup.packages.${system}.ghost-backup;
    
    # Customize service user (optional)
    user = "ghost-backup";
    group = "ghost-backup";
  };
}
```

### Module Options

- `enable`: Enable the Ghost Backup service (default: false)
- `package`: The ghost-backup package to use (default: latest from flake)
- `user`: User account under which the service runs (default: "ghost-backup")
- `group`: Group under which the service runs (default: "ghost-backup")

### What the Module Does

When enabled, the module:

1. **Creates a system user and group** for running the service
2. **Configures a systemd service** that:
   - Runs `ghost-backup service run`
   - Starts after network is available
   - Restarts on failure
   - Uses security hardening (PrivateTmp, ProtectSystem, etc.)
3. **Sets up directories**:
   - Working directory: `/var/lib/ghost-backup`
   - State directory: `/var/lib/ghost-backup` (automatically created)
   - Logs: Managed by systemd journal

### Service Management

Once enabled in your NixOS configuration:

```bash
# Check service status
systemctl status ghost-backup

# View logs
journalctl -u ghost-backup -f

# Restart service
systemctl restart ghost-backup
```

## Flake Outputs

The flake provides the following outputs for each supported system:

### Packages

- `packages.default`: The ghost-backup binary
- `packages.ghost-backup`: The ghost-backup binary (same as default)

### Apps

- `apps.default`: Run ghost-backup directly with `nix run`

### NixOS Modules

- `nixosModules.default`: The systemd service module

### Dev Shells

- `devShells.default`: Development environment with Go and tools

## Building from Source

```bash
# Clone the repository
git clone https://github.com/neoscode/ghost-backup.git
cd ghost-backup

# Build with Nix
nix build

# The binary will be in ./result/bin/ghost-backup
./result/bin/ghost-backup --help
```

## Vendored Dependencies

The flake uses `buildGoModule` which automatically handles Go dependencies via `vendorHash`. If you update Go dependencies:

1. Update `go.mod` and `go.sum`
2. Change `vendorHash` in `flake.nix` to `pkgs.lib.fakeHash`
3. Run `nix build` - it will fail and show the correct hash
4. Update `vendorHash` with the correct value
5. Build again

## Security Features

The systemd service includes several security hardening options:

- `NoNewPrivileges = true`: Prevents privilege escalation
- `PrivateTmp = true`: Isolated /tmp directory
- `ProtectSystem = "strict"`: Read-only system directories
- `ProtectHome = true`: Home directories not accessible
- `ReadWritePaths = [ "/var/lib/ghost-backup" ]`: Only writable path

## Troubleshooting

### Service fails to start

Check the journal logs:
```bash
journalctl -u ghost-backup -n 50
```

### Permission issues

The service runs as the `ghost-backup` user by default. Ensure repositories are accessible:
```bash
# Add ghost-backup user to a group if needed
usermod -aG yourgroup ghost-backup
```

### Configuration not found

The service looks for configuration in:
- `~/.config/ghost-backup/registry.json` (for the service user)
- Each repository's `.ghost-backup.json`

Make sure the registry is created for the service user.

## Contributing

When modifying the flake:

1. Test with `nix flake check`
2. Verify all outputs with `nix flake show`
3. Test the build with `nix build`
4. Test the app with `nix run .# -- --help`
5. If on NixOS, test the module in a VM

