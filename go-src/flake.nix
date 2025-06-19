{
  description = "Coolify Go Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            air # Live reload for Go
            docker
            docker-compose
            git
            curl
            jq
          ];

          shellHook = ''
            echo "🚀 Coolify Go Development Environment"
            echo "Go version: $(go version)"
            echo "Docker: $(docker --version)"
            echo ""
            echo "Available commands:"
            echo "  docker-compose up -d   - Start PostgreSQL & Redis containers"
            echo "  go run main.go         - Run the server"
            echo "  air                    - Run with live reload"
            echo ""
            
            # Set environment variables
            export POSTGRES_URL="postgres://coolify:password@localhost:5432/coolify?sslmode=disable"
            export REDIS_URL="redis://localhost:6379"
            export PORT="8080"
            export GO_ENV="development"
          '';
        };
      });
}
