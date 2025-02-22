#!/bin/bash

set -e

function main {
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  REPO_DIR="$SCRIPT_DIR/../.."
  BINARY_PATH="${REPO_DIR}/lod2"

  #
  echo "1. Starting update process..."

  cd "$REPO_DIR"

  ARCHIVE_PATH="$REPO_DIR/_archive_bin/$(git rev-parse HEAD)"

  #
  echo "2. Pulling latest changes from Git..."

  if ! git diff-index --quiet HEAD --; then
    echo "hey"
    #echo "! Git repository '${REPO_DIR}' is not clean. Please commit or stash your changes."
    #exit 1
  fi

  #git pull origin main || echo "Offline or an error occurred. Skipping 'git pull'."

  #
  echo "3. Rebuilding the binary..."

  # Archive the existing binary to archives/<git commit hash>, if it exists.
  mkdir -p "$REPO_DIR/_archive_bin"
  [ -f "$BINARY_PATH" ] && cp "$BINARY_PATH" "$ARCHIVE_PATH"

  # If anything goes wrong, restore the archived binary.
  trap "mv '$ARCHIVE_PATH' '$BINARY_PATH'" EXIT
  go build -o "$BINARY_PATH"

  #
  echo "4. Terminating any running instances..."
  pkill -SIGTERM -f "$BINARY_PATH" || echo "   No running instances found; nothing to terminate"

  #
  echo "5. Starting the updated application..."
  exec "$BINARY_PATH" "$@"
}

main "$@"
