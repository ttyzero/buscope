#!/bin/sh
# buscope installer.
#
#   curl -fsSL https://raw.githubusercontent.com/ttyzero/buscope/main/install.sh | sh
#
# Or the whole ttyzero kit (gitwing, peek, buscope, ttythemer):
#   curl -fsSL https://raw.githubusercontent.com/ttyzero/ttybus/main/install-kit.sh | sh
#
# Optional env:
#   VERSION   release tag (default: latest GitHub release, else go install @main)
#   PREFIX    install prefix (default: $HOME/.local) — binary goes in $PREFIX/bin
#   BINDIR    override install directory
#   REPO      GitHub owner/name (default: ttyzero/buscope)
#   SOURCE    "release" (default) or "go"
set -eu

BIN=buscope
REPO="${REPO:-ttyzero/buscope}"
PREFIX="${PREFIX:-${HOME}/.local}"
BINDIR="${BINDIR:-${PREFIX}/bin}"
SOURCE="${SOURCE:-release}"

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf '%s-install: need %s\n' "$BIN" "$1" >&2
		exit 1
	fi
}

need curl
need uname
need mktemp

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*)
	printf '%s-install: unsupported arch %s\n' "$BIN" "$arch" >&2
	exit 1
	;;
esac
case "$os" in
linux | darwin) ;;
*)
	printf '%s-install: unsupported os %s\n' "$BIN" "$os" >&2
	exit 1
	;;
esac

latest_tag() {
	tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null |
		grep -m1 '"tag_name"' |
		sed 's/.*"tag_name": *"\([^"]*\)".*/\1/') || true
	if [ -n "$tag" ] && [ "$tag" != "null" ]; then
		printf '%s\n' "$tag"
		return
	fi
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
		"https://github.com/${REPO}/releases/latest") || true
	base=${url##*/}
	case "$base" in
	v[0-9]*) printf '%s\n' "$base" ;;
	esac
}

go_install() {
	ref=$1
	need go
	if [ -z "$ref" ] || [ "$ref" = latest ]; then
		ref=main
	fi
	printf '%s-install: go install github.com/%s/cmd/%s@%s -> %s\n' \
		"$BIN" "$REPO" "$BIN" "$ref" "$BINDIR" >&2
	if ! mkdir -p "$BINDIR" 2>/dev/null || [ ! -w "$BINDIR" ]; then
		printf '%s-install: cannot write %s (set PREFIX or BINDIR)\n' "$BIN" "$BINDIR" >&2
		exit 1
	fi
	GOBIN="$BINDIR" go install "github.com/${REPO}/cmd/${BIN}@${ref}"
	printf 'installed %s to %s\n' "$("$BINDIR/$BIN" --version)" "$BINDIR/$BIN" >&2
}

if [ "$SOURCE" = go ]; then
	go_install "${VERSION:-main}"
	exit 0
fi

need tar

VERSION="${VERSION:-$(latest_tag)}"
case "$VERSION" in
v[0-9]*) ;;
*)
	printf '%s-install: no GitHub release; building with go\n' "$BIN" >&2
	go_install main
	exit 0
	;;
esac

asset="${BIN}_${VERSION}_${os}_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/download/${VERSION}"
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT

printf '%s-install: %s %s/%s -> %s\n' "$BIN" "$VERSION" "$os" "$arch" "$BINDIR" >&2

if ! curl -fsSL -o "$workdir/$asset" "${base}/${asset}" 2>/dev/null; then
	printf '%s-install: no asset %s; building with go\n' "$BIN" "$asset" >&2
	go_install "$VERSION"
	exit 0
fi
if ! curl -fsSL -o "$workdir/checksums.txt" "${base}/checksums.txt" 2>/dev/null; then
	printf '%s-install: no checksums.txt; building with go\n' "$BIN" >&2
	go_install "$VERSION"
	exit 0
fi

want=$(awk -v f="$asset" '$2 == f { print $1 }' "$workdir/checksums.txt")
if [ -z "$want" ]; then
	printf '%s-install: %s not listed in checksums.txt\n' "$BIN" "$asset" >&2
	exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
	got=$(sha256sum "$workdir/$asset" | awk '{ print $1 }')
else
	need shasum
	got=$(shasum -a 256 "$workdir/$asset" | awk '{ print $1 }')
fi
if [ "$want" != "$got" ]; then
	printf '%s-install: checksum mismatch for %s\n' "$BIN" "$asset" >&2
	exit 1
fi

tar -C "$workdir" -xzf "$workdir/$asset"
if [ ! -f "$workdir/$BIN" ]; then
	printf '%s-install: archive missing %s\n' "$BIN" "$BIN" >&2
	exit 1
fi
chmod 755 "$workdir/$BIN"

if ! mkdir -p "$BINDIR" 2>/dev/null || [ ! -w "$BINDIR" ]; then
	printf '%s-install: cannot write %s (set PREFIX or BINDIR)\n' "$BIN" "$BINDIR" >&2
	exit 1
fi
mv "$workdir/$BIN" "$BINDIR/$BIN"

printf 'installed %s to %s\n' "$("$BINDIR/$BIN" --version)" "$BINDIR/$BIN" >&2
case ":$PATH:" in
*":$BINDIR:"*) ;;
*)
	printf '%s-install: add %s to PATH\n' "$BIN" "$BINDIR" >&2
	;;
esac
