{
  description = "Ghost Backup - Automated Git backup service for uncommitted changes";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      # Per-system outputs
      perSystemOutputs = flake-utils.lib.eachDefaultSystem (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          ghost-backup = pkgs.callPackage ./nix/package.nix { inherit self; };
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

          devShells.default = pkgs.callPackage ./nix/shell.nix { };
        });
    in
    perSystemOutputs // {
      # NixOS module (defined once, not per-system)
      nixosModules.default = import ./nix/module.nix { inherit perSystemOutputs; };
    };
}

