#!/usr/bin/env bash
set -euo pipefail

# Simple Traefik bootstrap for Docker

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

NETWORK_NAME="${TRAEFIK_NETWORK:-traefik-network}"
TRAEFIK_VERSION="${TRAEFIK_VERSION:-v3.0}"
DASHBOARD_PORT="${TRAEFIK_DASHBOARD_PORT:-8080}"
EMAIL="${TRAEFIK_ACME_EMAIL:-your-email@example.com}"
MODE="${TRAEFIK_MODE:-dev}"

if [ "$MODE" != "dev" ] && [ "$MODE" != "prod" ]; then
  echo "Invalid TRAEFIK_MODE: $MODE (expected: dev or prod)" >&2
  exit 1
fi

compose_cmd() {
  if command -v docker >/dev/null 2>&1; then
    if docker compose version >/dev/null 2>&1; then
      echo "docker compose"
      return 0
    fi
  fi
  if command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
    return 0
  fi
  return 1
}

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is required but not installed." >&2
  exit 1
fi

if ! COMPOSE_BIN="$(compose_cmd)"; then
  echo "Docker Compose is required but not installed." >&2
  exit 1
fi

if ! docker network inspect "$NETWORK_NAME" >/dev/null 2>&1; then
  docker network create "$NETWORK_NAME"
fi

if [ ! -f traefik.yml ]; then
  if [ "$MODE" = "prod" ]; then
    cat > traefik.yml <<EOF
api:
  dashboard: true
  insecure: false

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
          permanent: true
  websecure:
    address: ":443"
    http:
      tls:
        certResolver: letsencrypt

certificatesResolvers:
  letsencrypt:
    acme:
      email: "$EMAIL"
      storage: /acme.json
      httpChallenge:
        entryPoint: web

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: $NETWORK_NAME

log:
  level: INFO

accessLog: {}
EOF
  else
    cat > traefik.yml <<EOF
api:
  dashboard: true
  insecure: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: $NETWORK_NAME

log:
  level: INFO

accessLog: {}
EOF
  fi
fi

if [ ! -f docker-compose.yml ]; then
  if [ "$MODE" = "prod" ]; then
    cat > docker-compose.yml <<EOF
version: '3'

services:
  traefik:
    image: traefik:$TRAEFIK_VERSION
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/traefik.yml:ro
      - ./acme.json:/acme.json
    networks:
      - $NETWORK_NAME

networks:
  $NETWORK_NAME:
    external: true
EOF
  else
    cat > docker-compose.yml <<EOF
version: '3'

services:
  traefik:
    image: traefik:$TRAEFIK_VERSION
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "$DASHBOARD_PORT:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/traefik.yml:ro
      - ./acme.json:/acme.json
    networks:
      - $NETWORK_NAME

networks:
  $NETWORK_NAME:
    external: true
EOF
  fi
fi

if [ ! -f acme.json ]; then
  touch acme.json
  chmod 600 acme.json
fi

# Inject ACME email if a template exists
if grep -q "your-email@example.com" traefik.yml; then
  sed -i.bak "s/your-email@example.com/${EMAIL}/" traefik.yml || true
  rm -f traefik.yml.bak
fi

$COMPOSE_BIN up -d

if [ "$MODE" = "prod" ]; then
  echo "Traefik is running in production mode (dashboard insecure endpoint disabled)."
else
  echo "Traefik is running. Dashboard: http://localhost:${DASHBOARD_PORT}"
fi
