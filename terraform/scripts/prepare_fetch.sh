#!/bin/sh

# Exit if any step fails
set -e

eval "$(jq -r '@sh "GITREF=\(.gitref)"')"

# Useful for debugging the script. Comment out the eval if running for debugging purposes
# GITREF="refs/heads/<branch-name>"

# Portable construct so this will work everywhere
# https://unix.stackexchange.com/a/84980
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



# Grab a copy of the zip file for the specified ref
# curl -s -L "https://github.com/jadudm/eight/archive/${GITREF}.zip" --output local.zip
# branch=$(echo "$GITREF" | cut -f3 -d"/")




# Zip up just the FAC-main/ subdirectory for pushing
# Before zip stage, run [ npm ci --production | npm run build ] in /backend/ to get the compiled assets for the site in /static/compiled/
# unzip -q -u local.zip \*"eight-$branch/backend/*"\*
# cd "${tmpdir}/FAC-$branch/backend/" &&
# npm ci --production --silent &&
# npm run build > '/dev/null' 2>&1 &&
# zip -r -o -X "${popdir}/app.zip" ./ > /dev/null
# zip -q -j -r ${popdir}/app.zip fac-*/backend

# Tell Terraform where to find it
cat << EOF
{ "path": "zips/app.zip" }
EOF