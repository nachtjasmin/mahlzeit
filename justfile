default: build

TOOLS_AIR_VERSION := "v1.43.0"
TOOLS_DBMATE_VERSION := "v2.2.0"
TOOLS_SQLC_VERSION := "1.18.0"

_install-tools:
	mkdir -p bin
	go install github.com/cosmtrek/air@{{TOOLS_AIR_VERSION}}
	curl -fsSL -o ./bin/dbmate https://github.com/amacneil/dbmate/releases/download/{{TOOLS_DBMATE_VERSION}}/dbmate-{{ os() }}-amd64
	chmod +x ./bin/dbmate
	curl -fsSL -o ./bin/sqlc.tgz https://github.com/kyleconroy/sqlc/releases/download/v{{TOOLS_SQLC_VERSION}}/sqlc_{{TOOLS_SQLC_VERSION}}_{{ os() }}_amd64.tar.gz
	tar xzf ./bin/sqlc.tgz --directory ./bin/
	rm ./bin/sqlc.tgz

_install-deps:
	go mod download

# Build the application, it's saved under ./mahlzeit
build: _install-deps
	go generate ./web
	go build -o mahlzeit ./cmd/mahlzeit

tmpdir  := `mktemp -d`
version := "0.0.1"
tarfile := "mahlzeit-" + version + "-" + os() + "-" + arch()
tardir  := tmpdir / tarfile
tarball := tardir + ".tar.gz"

# Package builds the application and compresses the binary and all necessary files
# into a single .tar.gz archive.
package: build
	mkdir -p {{tardir}}
	cp README.md LICENSE.md config.toml {{tardir}}
	cp mahlzeit {{tardir}}
	cp -r web/dist/assets/ {{tardir}}
	mkdir -p {{tardir}}/web/templates/ {{tardir}}/db/migrations/
	cp -r web/templates/* {{tardir}}/web/templates/
	cp -r db/migrations/* {{tardir}}/db/migrations/
	cd {{tardir}} && tar zcvf {{tarball}} .
	cp {{tarball}} {{invocation_directory()}}
	rm -rf {{tarball}} {{tardir}}

# Apply all pending database migrations.
migrate:
    docker compose up -d
    dbmate --wait up

# Installs the dependencies and applies all database migrations.
prepare: _install-deps migrate

# Start the watch mode for local development
dev: migrate
    air
