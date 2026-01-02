# Nix Configuration Files

This directory contains the separated Nix configuration for the Ghost Backup project.

## File Structure

- **`package.nix`** - The package definition for building ghost-backup
  - Defines the buildGoModule derivation
  - Includes build dependencies (git for tests)
  - Sets up test environment
  - Configures ldflags and metadata

- **`module.nix`** - The NixOS module for the systemd service
  - Provides `services.ghost-backup` configuration options
  - Creates system user and group
  - Configures systemd service with security hardening
  - Manages state directory

- **`shell.nix`** - The development shell environment
  - Provides Go toolchain and development tools
  - Includes gopls, gotools, go-tools
  - Custom shellHook with helpful information

## Usage

These files are imported by the main `flake.nix` in the project root:

```nix
# flake.nix structure
{
  outputs = { self, nixpkgs, flake-utils }:
    let
      perSystemOutputs = flake-utils.lib.eachDefaultSystem (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          ghost-backup = pkgs.callPackage ./nix/package.nix { inherit self; };
        in
        {
          packages.default = ghost-backup;
          devShells.default = pkgs.callPackage ./nix/shell.nix { };
        });
    in
    perSystemOutputs // {
      nixosModules.default = import ./nix/module.nix { inherit perSystemOutputs; };
    };
}
```

## Modifying

### To update the package build:
Edit `package.nix` to change:
- Build flags
- Dependencies
- Test configuration
- Version or metadata

### To modify the NixOS service:
Edit `module.nix` to:
- Add new configuration options
- Change systemd service settings
- Adjust security hardening
- Modify user/group configuration

### To change the dev environment:
Edit `shell.nix` to:
- Add new development tools
- Change environment variables
- Customize the shell prompt
- Add pre-commit hooks

## Benefits of Separation

1. **Modularity** - Each concern is isolated in its own file
2. **Reusability** - Files can be imported in different contexts
3. **Maintainability** - Easier to find and modify specific functionality
4. **Clarity** - The main flake.nix remains concise and readable
5. **Testing** - Individual components can be tested separately

