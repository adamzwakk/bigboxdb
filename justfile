build:
    podman compose -f dev-compose.yml up -d

run:
    cd src && go run .