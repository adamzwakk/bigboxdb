build:
    cd src && go build -o ../dist/bigboxdb_server

run:
    cd src && go run .

build-release:
    cd src && go build -ldflags="-s -w" -o ../dist/bigboxdb_server_release