set dotenv-load := true

up:
    podman compose up -d

down:
    podman compose down

lsql:
    lazysql "mariadb://${MYSQL_USER}:${MYSQL_PASSWORD}@localhost:3306"

build:
    cd src && go build -o ../dist/bigboxdb_server

run:
    cd src && go run .

build-release:
    cd src && go build -ldflags="-s -w" -o ../dist/bigboxdb_server_release