{ pkgs, self }:

pkgs.buildGoModule {
  pname = "ghost-backup";
  version = "0.1.0";

  src = ../.;

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
    mainProgram = "ghost-backup";
    description = "Automated Git backup service for uncommitted changes";
    homepage = "https://gitlab.neoscode.com/development-tools/ghost-backup";
    license = licenses.mit;
    maintainers = [ ];
  };
}

