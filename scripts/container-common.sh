#!/usr/bin/env bash
set -euo pipefail

ensure_container_running() {
  local container_name="$1"

  if ! command -v docker >/dev/null 2>&1; then
    echo "docker CLI is not installed or not available in PATH."
    echo "Install Docker Desktop, then retry."
    exit 1
  fi

  if [[ "$(docker inspect -f '{{.State.Running}}' "$container_name" 2>/dev/null || true)" == "true" ]]; then
    return 0
  fi

  echo "Container '$container_name' is not running."
  echo
  echo "How to fix:"
  echo "  1) Start the container: make docker-run"
  echo "  2) Verify status:      docker ps --filter name=$container_name"
  echo "  3) Then retry your command"
  exit 1
}
