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
}

