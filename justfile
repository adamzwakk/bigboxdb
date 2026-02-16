set dotenv-load := true

## Fun Notes
## for dir in /mnt/Projects/PcBoxes/games/*/webfiles-gltf/; do go run . import "$dir"; done

up-services:
    podman compose up -d mariadb redis meilisearch

down-services:
    podman compose down mariadb redis meilisearch

lsql:
    lazysql "mariadb://${MYSQL_USER}:${MYSQL_PASSWORD}@${MYSQL_HOST}:${MYSQL_PORT}"

reset-db:
    podman compose down mariadb && sudo rm -rf ./data-sql && podman compose up -d mariadb

build:
    cd bbdb/server && go build -o ../dist/bigboxdb_server

run-server:
    cd bbdb && go run ./server host

migrate:
    cd bbdb/server && go run . migrate

build-release:
    cd bbdb/server && go build -ldflags="-s -w" -o ../dist/bigboxdb_server_release

get-admin-key:
    podman compose exec mariadb mariadb -D "${MYSQL_DATABASE}" -u ${MYSQL_USER} -p${MYSQL_PASSWORD} -N -s -e "select api_key from users where id = 1;"

get-meilisearch-key:
    cd bbdb/server && go run . init-meilisearch

web-install:
    cd web && npm install

web-dev:
    cd web && npm run dev

web-build:
    cd web && npm run build

prod-up:
    podman compose -f compose.prod.yml up -d

prod-down:
    podman compose -f compose.prod.yml down

prod-build:
    podman compose -f compose.prod.yml build server && podman compose -f compose.prod.yml up -d

prod-build-nc:
    podman compose -f compose.prod.yml build server --no-cache && podman compose -f compose.prod.yml up -d

prod-migrate:
    podman compose -f compose.prod.yml exec server /app/bin/server migrate 

prod-get-admin-key:
    podman compose -f compose.prod.yml exec mariadb mariadb -D "${MYSQL_DATABASE}" -u ${MYSQL_USER} -p${MYSQL_PASSWORD} -N -s -e "select api_key from users where id = 1;"

prod-get-meilisearch-key:
    podman compose -f compose.prod.yml exec server /app/bin/server init-meilisearch