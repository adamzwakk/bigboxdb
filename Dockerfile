# --- Web Builder --- #
FROM node:22-alpine AS frontend-build
WORKDIR /app
COPY web/ ./web
COPY .env ./.env
WORKDIR /app/web
RUN npm ci
RUN npm run build

# --- Go Builder --- #
FROM golang:1.25 AS backend-build
WORKDIR /app
COPY bbdb ./src
COPY .env ./.env
WORKDIR /app/src
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server-dist ./server

# ---- Run stage ----
FROM nginx:1.29.5-alpine-slim
RUN echo "@testing https://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache vips vips-tools ktx@testing supervisor
WORKDIR /app
COPY --from=frontend-build /app/web/dist /usr/share/nginx/html
COPY --from=backend-build /app/server-dist /app/bin/server
COPY --from=backend-build /app/.env /app/.env

EXPOSE 80
CMD ["supervisord", "-c", "/etc/supervisord.conf"]