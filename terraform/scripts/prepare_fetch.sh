#!/bin/sh
# Exit if any step fails
set -e

TMP_DIR=$(mktemp -d -t 'build_temp.XXX')
TOP_LEVEL_DIR="${PWD}"

# This temporary directory will be in /tmp on a Linux-type machine.
pushd "$TMP_DIR"
  # Get the current repo
  wget https://github.com/jadudm/eight/archive/refs/heads/main.zip
  unzip main.zip
  pushd eight-main/cmd/fetch
    # Head into the `fetch` directory
    make build
    # leaves fetch.exe in eight/cmd/fetch
    mkdir -p /app
    mv eight/cmd/fetch.exe /app
    zip -r -o -X "${TOP_LEVEL_DIR}/zips/app.zip" /app > /dev/null
  popd
popd

# Tell Terraform where to find it
cat << EOF
{ "path": "zips/app.zip" }
EOF