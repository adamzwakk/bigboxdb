set dotenv-load := true

up:
    podman compose up -d

down:
    podman compose down

lsql:
    lazysql "mariadb://${MYSQL_USER}:${MYSQL_PASSWORD}@${MYSQL_HOST}:${MYSQL_PORT}"

reset-db:
    podman compose down mariadb && podman volume rm bigboxdb-server_db-data && podman compose up -d mariadb

build:
    cd src && go build -o ../dist/bigboxdb_server

run:
    cd src && go run . host

migrate:
    cd src && go run . migrate

build-release:
    cd src && go build -ldflags="-s -w" -o ../dist/bigboxdb_server_release

get-admin-key:
    podman compose exec mariadb mariadb -D "${MYSQL_DATABASE}" -u ${MYSQL_USER} -p${MYSQL_PASSWORD} -N -s -e "select api_key from users where id = 1;"