#!/usr/bin/env bash
# scripts/deploy.sh
#
# Convenience wrapper around docker compose for production deployments.
# Run this on your server (or call it from a CI step).
#
# Usage:
#   ./scripts/deploy.sh           # pull latest image, rebuild, restart
#   ./scripts/deploy.sh logs      # follow container logs
#   ./scripts/deploy.sh stop      # stop the stack
#   ./scripts/deploy.sh status    # show running containers

set -euo pipefail

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"

# Resolve the repo root (wherever this script lives)
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

case "${1:-up}" in
  up|deploy)
    echo "▶ Pulling latest code..."
    git pull origin main

    echo "▶ Building and starting containers..."
    docker compose $COMPOSE_FILES up --build -d --remove-orphans

    echo "▶ Cleaning up dangling images..."
    docker image prune -f

    echo "✅ Deployed. Access the app at http://$(hostname):${PORT:-8080}"
    ;;
  logs)
    docker compose $COMPOSE_FILES logs -f
    ;;
  stop)
    docker compose $COMPOSE_FILES down
    ;;
  status)
    docker compose $COMPOSE_FILES ps
    ;;
  *)
    echo "Usage: $0 [up|logs|stop|status]"
    exit 1
    ;;
esac
