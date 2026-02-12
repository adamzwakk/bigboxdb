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