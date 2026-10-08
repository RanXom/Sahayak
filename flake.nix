{
  description = "Sahayak — development environment (Go + SvelteKit + PostgreSQL)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f (import nixpkgs { inherit system; }));
    in
    {
      devShells = forAllSystems (pkgs:
        let
          # Version-pinned where it matters; fall back to the default attr if the
          # pinned name is not present in the channel.
          go = pkgs.go_1_26 or pkgs.go;
          node = pkgs.nodejs_22 or pkgs.nodejs;
          postgres = pkgs.postgresql_16 or pkgs.postgresql;
        in
        {
          default = pkgs.mkShell {
            packages = [
              go
              node
              pkgs.pnpm
              postgres
              pkgs.golangci-lint
              pkgs.ffmpeg
              pkgs.gnumake
              pkgs.git
              pkgs.curl
              pkgs.jq
              pkgs.openssl
              pkgs.python3
              pkgs.direnv
            ];

            shellHook = ''
              export GOTOOLCHAIN=local
              # Keep the data directory OUTSIDE the repo: Nix cannot copy unix
              # sockets into the store, and a live PG socket under the flake
              # source breaks `nix develop`/direnv evaluation.
              export PGDATA="''${XDG_DATA_HOME:-$HOME/.local/share}/sahayak/pgdata"
              export PGPORT="55432"
              export PGHOST="$PGDATA"
              export PGDATABASE="sahayak"
              export PGUSER="sahayak"
              mkdir -p "$(dirname "$PGDATA")"

              # Cross-shell helpers (bash/zsh/fish) live in ./bin.
              export PATH="${./bin}:$PATH"

              echo ""
              echo "  Sahayak dev shell"
              echo "  ├─ go       $(go version | cut -d' ' -f3)"
              echo "  ├─ node     $(node --version)   pnpm $(pnpm --version 2>/dev/null || echo '-')"
              echo "  ├─ postgres $(postgres --version | awk '{print $3}')"
              echo "  └─ psql, golangci-lint, ffmpeg, jq, openssl, python3"
              echo ""
              echo "  Database:  pg-start && pg-create   (socket on $PGHOST, port $PGPORT)"
              echo "  App:       make migrate && make seed && make run"
              echo "  Frontend:  cd frontend && npm install && npm run dev"
              echo ""
            '';
          };
        });
    };
}
