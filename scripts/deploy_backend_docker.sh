#!/usr/bin/env bash
set -Eeuo pipefail

BRANCH="${1:-}"

if [[ -z "$BRANCH" ]]; then
  echo "Usage: $0 <branch>"
  echo "Example: $0 main"
  exit 1
fi

REPO_DIR="${REPO_DIR:-/opt/InterVUZ}"
ENV_FILE="${ENV_FILE:-$REPO_DIR/.env}"
ENV_TEMPLATE="${ENV_TEMPLATE:-$REPO_DIR/scripts/backend.env.template}"
COMPOSE_FILE="${COMPOSE_FILE:-$REPO_DIR/docker-compose.yml}"

log() {
  printf '[backend-deploy] %s\n' "$1"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: command '$1' is not installed"
    exit 1
  fi
}

require_cmd git
require_cmd docker

if [[ ! -d "$REPO_DIR/.git" ]]; then
  echo "Error: '$REPO_DIR' is not a git repository"
  exit 1
fi

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "Error: docker compose file not found: $COMPOSE_FILE"
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  if [[ -f "$ENV_TEMPLATE" ]]; then
    cp "$ENV_TEMPLATE" "$ENV_FILE"
    echo "Created $ENV_FILE from template. Edit it and re-run deploy."
  else
    echo "Error: env file not found and template missing: $ENV_TEMPLATE"
  fi
  exit 1
fi

log "Fetching origin..."
git -C "$REPO_DIR" fetch --prune origin

if ! git -C "$REPO_DIR" show-ref --verify --quiet "refs/remotes/origin/$BRANCH"; then
  echo "Error: branch 'origin/$BRANCH' not found"
  exit 1
fi

if git -C "$REPO_DIR" rev-parse --verify --quiet "$BRANCH" >/dev/null; then
  log "Checking out existing local branch '$BRANCH'..."
  git -C "$REPO_DIR" checkout "$BRANCH"
else
  log "Creating local tracking branch '$BRANCH'..."
  git -C "$REPO_DIR" checkout -b "$BRANCH" --track "origin/$BRANCH"
fi

log "Resetting local branch to origin/$BRANCH..."
git -C "$REPO_DIR" reset --hard "origin/$BRANCH"

log "Validating required env values..."
if ! grep -Eq '^ACADEMIC_WEEK1_START_DATE=.+' "$ENV_FILE"; then
  echo "Error: ACADEMIC_WEEK1_START_DATE is missing in $ENV_FILE"
  exit 1
fi

if ! grep -Eq '^ASSISTANT_API_KEY=.+' "$ENV_FILE"; then
  echo "Warning: ASSISTANT_API_KEY is empty in $ENV_FILE"
  echo "Assistant endpoint /assistant/chat will return unavailable."
fi

cd "$REPO_DIR"

log "Pulling latest images..."
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull || true

log "Building and starting containers..."
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up --build -d

log "Current compose status:"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps

log "Health check: GET /health"
if command -v curl >/dev/null 2>&1; then
  APP_PORT_VALUE="$(grep -E '^APP_PORT=' "$ENV_FILE" | tail -n1 | cut -d= -f2- || true)"
  APP_PORT_VALUE="${APP_PORT_VALUE:-8000}"
  sleep 2
  curl -fsS "http://127.0.0.1:${APP_PORT_VALUE}/health" || {
    echo "Warning: health check failed. Check logs:"
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs --tail=200 app
    exit 1
  }
fi

log "Deploy completed successfully."
