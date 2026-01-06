{ perSystemOutputs }:

{ config, lib, pkgs, ... }:

let
  cfg = config.services.ghost-backup;
in
{
  options.services.ghost-backup = {
    enable = lib.mkEnableOption "Ghost Backup service";

    package = lib.mkOption {
      type = lib.types.package;
      default = perSystemOutputs.packages.${pkgs.system}.default;
      defaultText = lib.literalExpression "pkgs.ghost-backup";
      description = "The ghost-backup package to use";
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = with pkgs; [
        git
        gitleaks
        cfg.package
    ];

    systemd.user.services.ghost-backup = {
      description = "Ghost Backup - Automated Git backup service";
      after = [ "network.target" ];
      wantedBy = [ "default.target" ];

      path = with pkgs; [
        git
        gitleaks
      ];

      serviceConfig = {
        Type = "simple";
        ExecStart = "${lib.getExe cfg.package} service run";
        Restart = "on-failure";
        RestartSec = "10s";

        # State directory for logs (under ~/.local/state/ghost-backup)
        StateDirectory = "ghost-backup";

        # Security hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
      };
    };
  };
}

