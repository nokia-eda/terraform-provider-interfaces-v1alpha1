#!/usr/bin/env bash

# This script downloads tfplugindocs binary for the specified OS/ARCH

set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <url> <zipname> <build-dir>" >&2
  exit 2
fi

URL="$1"
ZIPNAME="$2"
BUILD_DIR="$3"

mkdir -p "$BUILD_DIR"
TMPDIR="$(mktemp -d)"
ZIPFILE="$TMPDIR/$ZIPNAME"

cleanup() {
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

echo "Downloading $URL"
if command -v curl >/dev/null 2>&1; then
  curl -sfL --output "$ZIPFILE" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -O "$ZIPFILE" "$URL"
else
  echo "Neither curl nor wget is installed; please install one to download tfplugindocs" >&2
  exit 1
fi

echo "Extracting $ZIPFILE"
if command -v unzip >/dev/null 2>&1; then
  unzip -o "$ZIPFILE" -d "$TMPDIR" >/dev/null
else
  python3 - <<PY
import sys, zipfile
zipfile.ZipFile(sys.argv[1]).extractall(sys.argv[2])
PY
fi

# Locate the tfplugindocs binary inside the extracted tree
BINPATH=""
BINPATH=$(find "$TMPDIR" -maxdepth 3 -type f -name 'tfplugindocs*' -perm /u=x,g=x,o=x 2>/dev/null | head -n1 || true)
if [ -z "$BINPATH" ]; then
  BINPATH=$(find "$TMPDIR" -maxdepth 3 -type f -name 'tfplugindocs*' 2>/dev/null | head -n1 || true)
fi

if [ -z "$BINPATH" ]; then
  echo "tfplugindocs binary not found inside archive" >&2
  exit 1
fi

mv "$BINPATH" "$BUILD_DIR/tfplugindocs"
chmod +x "$BUILD_DIR/tfplugindocs"
echo "tfplugindocs saved to $BUILD_DIR/tfplugindocs"
