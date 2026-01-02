{ pkgs }:

pkgs.mkShell {
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
}

