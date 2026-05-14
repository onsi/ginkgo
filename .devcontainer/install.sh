#!/usr/bin/env bash
set -e

echo "finalizing devcontainer setup.."

if [ "$(id -u)" -ne 0 ]; then
    echo -e 'script must be run as root.'
    exit 1
fi

WSCSSHARED="/workspaces/.codespaces/shared"
mkdir -p "${WSCSSHARED}"
cp .devcontainer/first-run-notice.ansi "${WSCSSHARED}/first-run-notice.txt"
cp .devcontainer/ginkgotest .devcontainer/serve-docs /usr/local/bin

echo "done."
