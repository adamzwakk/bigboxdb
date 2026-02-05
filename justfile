set dotenv-load := true

up-services:
    podman compose up -d mariadb redis meilisearch

down-services:
    podman compose down mariadb redis meilisearch

lsql:
    lazysql "mariadb://${MYSQL_USER}:${MYSQL_PASSWORD}@${MYSQL_HOST}:${MYSQL_PORT}"

reset-db:
    podman compose down mariadb && podman volume rm bigboxdb-server_db-data && podman compose up -d mariadb

build:
    cd server && go build -o ../dist/bigboxdb_server

run-server:
    cd server && go run . host

migrate:
    cd server && go run . migrate

build-release:
    cd server && go build -ldflags="-s -w" -o ../dist/bigboxdb_server_release

get-admin-key:
    podman compose exec mariadb mariadb -D "${MYSQL_DATABASE}" -u ${MYSQL_USER} -p${MYSQL_PASSWORD} -N -s -e "select api_key from users where id = 1;"

web-install:
    cd web && npm install

web-dev:
    cd web && npm run dev

web-build:
    cd web && npm run build

prod-build:
    just web-build && podman compose build server && podman compose up -d