#!/bin/sh
# doctree installer
#   curl -fsSL https://raw.githubusercontent.com/foreverfl/docktree/main/install.sh | sh
#
# Env overrides:
#   DOCTREE_INSTALL_DIR  install destination (default: $HOME/.local/bin)
#   DOCTREE_VERSION      tag to install (default: latest non-draft release)

set -eu

OWNER=foreverfl
REPO=docktree
BIN=doctree
INSTALL_DIR=${DOCTREE_INSTALL_DIR:-"$HOME/.local/bin"}

err() { printf 'install.sh: %s\n' "$*" >&2; exit 1; }
info() { printf '%s\n' "$*"; }

detect_os() {
  case "$(uname -s)" in
    Darwin) echo darwin ;;
    Linux)  echo linux ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    arm64|aarch64) echo arm64 ;;
    *) err "unsupported arch: $(uname -m)" ;;
  esac
}

resolve_version() {
  if [ -n "${DOCTREE_VERSION:-}" ]; then
    echo "$DOCTREE_VERSION"
    return
  fi
  curl -fsSL "https://api.github.com/repos/$OWNER/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | head -1 \
    | cut -d'"' -f4
}

main() {
  command -v curl >/dev/null 2>&1 || err "curl is required"
  command -v tar  >/dev/null 2>&1 || err "tar is required"

  os=$(detect_os)
  arch=$(detect_arch)

  info "resolving doctree release..."
  tag=$(resolve_version)
  [ -n "$tag" ] || err "failed to resolve release tag (set DOCTREE_VERSION to override)"

  asset="${BIN}_${os}_${arch}.tar.gz"
  url="https://github.com/$OWNER/$REPO/releases/download/$tag/$asset"

  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT

  info "downloading $asset ($tag)..."
  curl -fsSL -o "$tmp/$asset" "$url" || err "download failed: $url"

  tar -xzf "$tmp/$asset" -C "$tmp"
  [ -f "$tmp/$BIN" ] || err "tarball did not contain $BIN"

  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$tmp/$BIN" "$INSTALL_DIR/$BIN"

  info "installed $INSTALL_DIR/$BIN ($tag)"

  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) info "warning: $INSTALL_DIR is not in PATH — add it to your shell rc:"
       info "  export PATH=\"$INSTALL_DIR:\$PATH\""
       ;;
  esac
}

main "$@"
