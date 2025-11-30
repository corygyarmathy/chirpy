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
        queriesDir = "./sql/queries";

        # PostgreSQL helper scripts
        dbstart = pkgs.writeShellScriptBin "dbstart" ''
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

        dbstop = pkgs.writeShellScriptBin "dbstop" ''
          if pg_ctl status >/dev/null 2>&1; then
            pg_ctl stop -m fast
            echo "PostgreSQL stopped"
          else
            echo "PostgreSQL is not running"
          fi
        '';

        dbstatus = pkgs.writeShellScriptBin "dbstatus" ''
          pg_ctl status
        '';

        dblogs = pkgs.writeShellScriptBin "dblogs" ''
          tail -f "$PGDATA/logfile"
        '';

        # Setup script to create project structure and config files
        setup-project = pkgs.writeShellScriptBin "setup-project" ''
          echo "Setting up project structure..."

          # Create directories
          mkdir -p ${migrationsDir}
          mkdir -p ${queriesDir}

          # Create .air.toml if it doesn't exist
          if [ ! -f .air.toml ]; then
            cat > .air.toml << 'EOF'
          root = "."
          testdata_dir = "testdata"
          tmp_dir = "tmp"

          [build]
            args_bin = []
            bin = "./tmp/main"
            cmd = "go build -o ./tmp/main ."
            delay = 1000
            exclude_dir = ["assets", "tmp", "vendor", "testdata", ".postgres"]
            exclude_file = []
            exclude_regex = ["_test.go"]
            exclude_unchanged = false
            follow_symlink = false
            full_bin = ""
            include_dir = []
            include_ext = ["go", "tpl", "tmpl", "html"]
            include_file = []
            kill_delay = "0s"
            log = "build-errors.log"
            poll = false
            poll_interval = 0
            post_cmd = []
            pre_cmd = []
            rerun = false
            rerun_delay = 500
            send_interrupt = false
            stop_on_error = false

          [color]
            app = ""
            build = "yellow"
            main = "magenta"
            runner = "green"
            watcher = "cyan"

          [log]
            main_only = false
            time = false

          [misc]
            clean_on_exit = false

          [screen]
            clear_on_rebuild = false
            keep_scroll = true
          EOF
            echo "✓ Created .air.toml"
          fi

          # Create sqlc.yaml if it doesn't exist
          if [ ! -f sqlc.yaml ]; then
            cat > sqlc.yaml << 'EOF'
          version: "2"
          sql:
            - schema: "${migrationsDir}"
              queries: "${queriesDir}"
              engine: "postgresql"
              gen:
                go:
                  package: "database"
                  out: "internal/database"
                  emit_json_tags: true
                  emit_interface: false
                  emit_exact_table_names: false
          EOF
            echo "✓ Created sqlc.yaml"
          fi

          echo "✓ Project structure ready"
          echo "  Migrations: ${migrationsDir}"
          echo "  Queries: ${queriesDir}"
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
            # Helper scripts
            dbstart
            dbstop
            dbstatus
            dblogs
            setup-project
          ];

          shellHook = ''
            # PostgreSQL configuration (PG* vars required by PostgreSQL tools)
            export PGDATA="$PWD/${pgDataDir}"
            export PGHOST="$PGDATA"
            export PGDATABASE="${dbName}"

            # Generic database configuration (for your application)
            export DB_URL="postgres:///$PGDATABASE?host=$PGHOST"
            export DATABASE_URL="$DB_URL"  # Alias for compatibility

            # Goose environment variables
            export GOOSE_DRIVER="postgres"
            export GOOSE_DBSTRING="$DB_URL"
            export GOOSE_MIGRATION_DIR="${migrationsDir}"

            # Auto-setup project structure on first run
            if [ ! -d "${migrationsDir}" ] || [ ! -f "sqlc.yaml" ] || [ ! -f ".air.toml" ]; then
              setup-project
            fi

            # Check if PostgreSQL is running
            if pg_ctl status >/dev/null 2>&1; then
              echo "✓ ${projectName} dev environment ready (PostgreSQL running)"
            else
              echo "✓ ${projectName} dev environment ready"
              echo "  Run 'dbstart' to start PostgreSQL"
            fi

            echo "  Commands: dbstart, dbstop, dbstatus, dblogs"
          '';
        };
      }
    );
}
