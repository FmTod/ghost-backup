{
  description = "Ghost Backup - Automated Git backup service for uncommitted changes";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        ghost-backup = pkgs.buildGoModule {
          pname = "ghost-backup";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-qmz0Qp5kj7AIdU47Kd/zfvquS5kB0Bnfhqq1mdEhTTQ=";

          # Add git to the build environment for tests
          nativeBuildInputs = [ pkgs.git ];

          # Make git available during tests
          preCheck = ''
            export HOME=$(mktemp -d)
            git config --global user.email "test@example.com"
            git config --global user.name "Test User"
          '';

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${self.rev or "dev"}"
          ];

          meta = with pkgs.lib; {
            description = "Automated Git backup service for uncommitted changes";
            homepage = "https://github.com/neoscode/ghost-backup";
            license = licenses.mit;
            maintainers = [ ];
          };
        };
      in
      {
        packages = {
          default = ghost-backup;
          ghost-backup = ghost-backup;
        };

        apps = {
          default = {
            type = "app";
            program = "${ghost-backup}/bin/ghost-backup";
          };
        };

        nixosModules.default = { config, lib, pkgs, ... }:
          let
            cfg = config.services.ghost-backup;
          in
          {
            options.services.ghost-backup = {
              enable = lib.mkEnableOption "Ghost Backup service";

              package = lib.mkOption {
                type = lib.types.package;
                default = ghost-backup;
                description = "The ghost-backup package to use";
              };

              user = lib.mkOption {
                type = lib.types.str;
                default = "ghost-backup";
                description = "User account under which ghost-backup runs";
              };

              group = lib.mkOption {
                type = lib.types.str;
                default = "ghost-backup";
                description = "Group under which ghost-backup runs";
              };
            };

            config = lib.mkIf cfg.enable {
              users.users.${cfg.user} = {
                isSystemUser = true;
                group = cfg.group;
                description = "Ghost Backup service user";
              };

              users.groups.${cfg.group} = {};

              systemd.services.ghost-backup = {
                description = "Ghost Backup - Automated Git backup service";
                after = [ "network.target" ];
                wantedBy = [ "multi-user.target" ];

                serviceConfig = {
                  Type = "simple";
                  User = cfg.user;
                  Group = cfg.group;
                  ExecStart = "${lib.getExe cfg.package} service run";
                  Restart = "on-failure";
                  RestartSec = "10s";

                  # Security hardening
                  NoNewPrivileges = true;
                  PrivateTmp = true;
                  ProtectSystem = "strict";
                  ProtectHome = true;
                  ReadWritePaths = [ "/var/lib/ghost-backup" ];
                  StateDirectory = "ghost-backup";
                  WorkingDirectory = "/var/lib/ghost-backup";
                };
              };
            };
          };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            git
          ];

          shellHook = ''
            echo "Ghost Backup development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go build        - Build the application"
            echo "  go test ./...   - Run all tests"
            echo "  go run .        - Run the application"
          '';
        };
      }
    );
}

