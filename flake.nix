{
  description = "Chirpy - Web Server in Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        # Project configuration
        projectName = "chirpy";
        dbName = projectName;
        pgDataDir = ".postgres";
        migrationsDir = "./sql/schema";

        # PostgreSQL helper scripts
        pgstart = pkgs.writeShellScriptBin "pgstart" ''
          # Initialise PostgreSQL if needed
          if [ ! -d "$PGDATA" ]; then
            echo "Initialising PostgreSQL..."
            initdb --auth=trust --no-locale --encoding=UTF8
          fi

          if pg_ctl status >/dev/null 2>&1; then
            echo "PostgreSQL is already running"
          else
            pg_ctl -o "-k $PGHOST" -l "$PGDATA/logfile" start
            
            # Wait for PostgreSQL to be ready
            for i in {1..10}; do
              if pg_isready -h "$PGHOST" >/dev/null 2>&1; then
                break
              fi
              sleep 0.5
            done
            
            # Auto-create database if it doesn't exist
            if ! psql -lqt | cut -d \| -f 1 | grep -qw ${dbName} 2>/dev/null; then
              createdb ${dbName} && echo "Created database: ${dbName}"
            fi
            
            echo "PostgreSQL started"
          fi
        '';

        pgstop = pkgs.writeShellScriptBin "pgstop" ''
          if pg_ctl status >/dev/null 2>&1; then
            pg_ctl stop -m fast
            echo "PostgreSQL stopped"
          else
            echo "PostgreSQL is not running"
          fi
        '';

        pgstatus = pkgs.writeShellScriptBin "pgstatus" ''
          pg_ctl status
        '';

        pglogs = pkgs.writeShellScriptBin "pglogs" ''
          tail -f "$PGDATA/logfile"
        '';

      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go toolchain
            go
            golangci-lint
            gotools
            delve

            # Database
            postgresql
            goose # Go database migration tool
            sqlc # SQL -> Go code generation tool

            # Development tools
            air # Live reload for Go
            jq # JSON processing for API testing

            # HTTP testing
            httpie # Perform HTTP requests in the CLI

            # Security
            govulncheck # Vulnerability scanner for Go dependencies

            # PostgreSQL helpers
            pgstart
            pgstop
            pgstatus
            pglogs
          ];

          shellHook = ''
            export PGDATA="$PWD/${pgDataDir}"
            export PGHOST="$PGDATA"
            export PGDATABASE="${dbName}"
            export DATABASE_URL="postgres:///$PGDATABASE?host=$PGHOST"

            # Goose environment variables
            export GOOSE_DRIVER="postgres"
            export GOOSE_DBSTRING="$DATABASE_URL"
            export GOOSE_MIGRATION_DIR="${migrationsDir}"

            # Check if PostgreSQL is running
            if pg_ctl status >/dev/null 2>&1; then
              echo "✓ ${projectName} dev environment ready (PostgreSQL running)"
            else
              echo "✓ ${projectName} dev environment ready"
              echo "  Run 'pgstart' to start PostgreSQL"
            fi

            echo "  Commands: pgstart, pgstop, pgstatus, pglogs"
          '';
        };
      }
    );
}
