#!/bin/sh

set -e

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
version=$(cat "$script_dir/../VERSION")

cd "$script_dir"
echo "Building web"
npm install
DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$version npm run build
