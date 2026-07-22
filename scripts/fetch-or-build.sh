#!/bin/sh
# Install-time build step for macOS/Linux (run by Herdr at `plugin install`,
# cwd = plugin root).
#
# Fast path: download the prebuilt binary matching this plugin's version and
# platform from GitHub Releases and verify its SHA-256. Fallback: build from
# source with `go` (only needed when no matching prebuilt exists).
set -eu

REPO="ShankyJS/herdr-space-scoped-agents"
BIN="herdr-space-scoped-agents"

root="${HERDR_PLUGIN_ROOT:-$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)}"
manifest="$root/herdr-plugin.toml"
out="$root/bin/$BIN"

version="$(sed -n 's/^version *= *"\(.*\)"/\1/p' "$manifest" | head -n1)"
[ -n "$version" ] || { echo "cannot read version from $manifest" >&2; exit 1; }

case "$(uname -s)" in
  Darwin) goos=darwin ;;
  Linux)  goos=linux ;;
  *)      goos="" ;;
esac
case "$(uname -m)" in
  x86_64|amd64)  goarch=amd64 ;;
  arm64|aarch64) goarch=arm64 ;;
  *)             goarch="" ;;
esac

mkdir -p "$root/bin"

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
  else shasum -a 256 "$1" | awk '{print $1}'; fi
}

download() {
  [ -n "$goos" ] && [ -n "$goarch" ] || return 1
  command -v curl >/dev/null 2>&1 || return 1
  asset="${BIN}-${goos}-${goarch}"
  base="https://github.com/${REPO}/releases/download/v${version}"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  echo "fetching ${asset} (v${version})..." >&2
  curl -fsSL "$base/$asset" -o "$tmp/bin" || return 1
  if curl -fsSL "$base/$asset.sha256" -o "$tmp/sum" 2>/dev/null; then
    want="$(awk '{print $1}' "$tmp/sum")"
    got="$(sha256 "$tmp/bin")"
    [ "$want" = "$got" ] || { echo "checksum mismatch (want $want got $got)" >&2; return 1; }
  else
    echo "warning: no published checksum; skipping verification" >&2
  fi
  mv "$tmp/bin" "$out"
  chmod +x "$out"
}

build() {
  command -v go >/dev/null 2>&1 || return 1
  echo "building from source with go..." >&2
  ( cd "$root" && go build -ldflags "-s -w -X main.version=${version}" -o "$out" . )
}

if download; then
  echo "installed prebuilt -> $out" >&2
elif build; then
  echo "built from source -> $out" >&2
else
  echo "ERROR: no prebuilt binary for ${goos:-?}/${goarch:-?} v${version} and no Go toolchain to build from source" >&2
  exit 1
fi
