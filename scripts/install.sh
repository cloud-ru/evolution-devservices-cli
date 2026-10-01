#!/usr/bin/env bash
# Self-installer for the Evolution DevServices CLI (eds).
#
# For Windows, use install.ps1 instead:
#   iwr -useb https://raw.githubusercontent.com/cloud-ru/evolution-devservices-cli/main/scripts/install.ps1 | iex
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/cloud-ru/evolution-devservices-cli/main/scripts/install.sh | bash
#   EDS_CLI_VERSION=v0.2.0 bash install.sh
#   ./scripts/install.sh --version v0.2.0 --repo cloud-ru/evolution-devservices-cli
#
# The script:
#   1. Detects OS / architecture.
#   2. Downloads the matching static binary from a GitHub Release (default)
#      or a custom mirror if EDS_CLI_BASE_URL/--base-url is set.
#   3. Installs it into $INSTALL_DIR (default ~/.local/bin).
#   4. Verifies the install with `eds version`.
#
# Environment variables (overridable via flags):
#   EDS_CLI_VERSION   - tag/version to download (default: latest)
#   EDS_CLI_REPO      - GitHub "org/repo" to fetch releases from
#                       (default: cloud-ru/evolution-devservices-cli)
#   EDS_CLI_BASE_URL  - base URL of a custom artifact mirror; when set, this
#                       is used instead of GitHub Releases (e.g. an internal
#                       S3-compatible bucket published via `make upload`)
#   EDS_CLI_DIR       - install directory (default ~/.local/bin)
#   EDS_CLI_BIN       - binary name (default: eds)
#   EDS_CLI_QUIET     - if set to 1, suppress progress output

set -euo pipefail

VERSION="${EDS_CLI_VERSION:-latest}"
REPO="${EDS_CLI_REPO:-cloud-ru/evolution-devservices-cli}"
BASE_URL="${EDS_CLI_BASE_URL:-}"
INSTALL_DIR="${EDS_CLI_DIR:-$HOME/.local/bin}"
BIN_NAME="${EDS_CLI_BIN:-eds}"
QUIET="${EDS_CLI_QUIET:-}"

usage() {
  cat <<EOF
Usage: install.sh [--version <ver>] [--repo <org/repo>] [--base-url <url>] [--dir <path>] [--bin <name>]

Defaults can also be passed via env vars: EDS_CLI_VERSION, EDS_CLI_REPO,
EDS_CLI_BASE_URL, EDS_CLI_DIR, EDS_CLI_BIN.

By default, binaries are downloaded from GitHub Releases at
https://github.com/\${EDS_CLI_REPO}/releases. Pass --base-url (or set
EDS_CLI_BASE_URL) to fetch from a custom mirror instead.

Examples:
  bash install.sh
  EDS_CLI_VERSION=v0.2.0 bash install.sh
  EDS_CLI_BASE_URL=https://storage.cloud.ru/evolution-devservices-cli bash install.sh
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --repo) REPO="$2"; shift 2 ;;
    --base-url) BASE_URL="$2"; shift 2 ;;
    --dir) INSTALL_DIR="$2"; shift 2 ;;
    --bin) BIN_NAME="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown flag: $1" >&2; usage; exit 2 ;;
  esac
done

log() {
  [[ -n "$QUIET" ]] || echo "[eds-cli installer] $*" >&2
}

# -- detect platform ---------------------------------------------------------

OS_RAW="$(uname -s)"
ARCH_RAW="$(uname -m)"

case "$OS_RAW" in
  Linux)   OS=linux ;;
  Darwin)  OS=darwin ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows ;;
  *)
    echo "Unsupported OS: $OS_RAW" >&2
    exit 1
    ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *)
    echo "Unsupported architecture: $ARCH_RAW" >&2
    exit 1
    ;;
esac

EXT=""
[[ "$OS" == "windows" ]] && EXT=".exe"

ASSET="eds-${OS}-${ARCH}${EXT}"

# -- resolve download URL -----------------------------------------------------
#
# Two modes:
#   - GitHub Releases (default): resolves "latest" via the GitHub API and
#     downloads from the release's asset URL.
#   - Custom mirror (--base-url/EDS_CLI_BASE_URL): same bucket layout as
#     `make upload` publishes (<base>/<version>/<asset>, <base>/latest).

if [[ -n "$BASE_URL" ]]; then
  BASE_URL="${BASE_URL%/}"
  if [[ "$VERSION" == "latest" ]]; then
    VERSION="$(curl -fsSL "${BASE_URL}/latest" 2>/dev/null || true)"
    if [[ -z "$VERSION" ]]; then
      echo "Could not resolve 'latest' version from ${BASE_URL}/latest" >&2
      exit 1
    fi
  fi
  URL="${BASE_URL}/${VERSION}/${ASSET}"
else
  if [[ "$VERSION" == "latest" ]]; then
    URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
  else
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
  fi
fi

log "Downloading $URL"

mkdir -p "$INSTALL_DIR"
TMP="$(mktemp -t eds-cli.XXXXXX)"
trap 'rm -f "$TMP"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -fsSL --retry 3 -o "$TMP" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$TMP" "$URL"
else
  echo "Neither curl nor wget is available" >&2
  exit 1
fi

chmod +x "$TMP"

DEST="${INSTALL_DIR}/${BIN_NAME}${EXT}"
mv "$TMP" "$DEST"

# -- verify ------------------------------------------------------------------

log "Installed to ${DEST}"
if ! "$DEST" version >/dev/null 2>&1; then
  echo "Installed binary failed to report version" >&2
  exit 1
fi
log "Verified: $($DEST version 2>/dev/null || echo unknown)"

# -- PATH hint ---------------------------------------------------------------

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    log "Add to PATH: export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac

echo "${DEST}"
