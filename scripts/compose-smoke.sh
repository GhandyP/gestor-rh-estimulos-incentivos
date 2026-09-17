#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT/compose.yaml"
PROJECT_NAME="${SMOKE_PROJECT_NAME:-estimulos-smoke-$$}"
HOST_PORT="${SMOKE_HOST_PORT:-${HOST_PORT:-0}}"
COOKIE_SECURE="${COOKIE_SECURE:-false}"
CONTAINER_NAME="${PROJECT_NAME}-app"
COMPOSE=(docker compose --project-name "$PROJECT_NAME" --file "$COMPOSE_FILE")

cleanup() {
    local status=$?
    set +e
    "${COMPOSE[@]}" down --volumes --remove-orphans >/dev/null 2>&1
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if ! docker compose version >/dev/null 2>&1; then
    echo "docker compose is required for the Compose smoke test" >&2
    exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required for the Compose smoke test" >&2
    exit 1
fi

HOST_PORT="$HOST_PORT" \
COOKIE_SECURE="$COOKIE_SECURE" \
COMPOSE_CONTAINER_NAME="$CONTAINER_NAME" \
"${COMPOSE[@]}" up -d --build

published_port="$("${COMPOSE[@]}" port app 8080 | awk -F: 'NF { print $NF; exit }')"
if [[ -z "$published_port" ]]; then
    echo "unable to determine the published Compose port" >&2
    exit 1
fi
base_url="http://127.0.0.1:$published_port"

for attempt in {1..60}; do
    if curl --fail --silent --show-error --max-time 3 "$base_url/readyz" >/dev/null; then
        break
    fi
    if (( attempt == 60 )); then
        echo "Compose service did not become ready" >&2
        "${COMPOSE[@]}" ps >&2 || true
        exit 1
    fi
    sleep 1
done

for endpoint in /healthz /login; do
    if ! curl --fail --silent --show-error --max-time 5 "$base_url$endpoint" >/dev/null; then
        echo "Compose smoke check failed for $endpoint" >&2
        exit 1
    fi
done

echo "Compose smoke passed at $base_url"
