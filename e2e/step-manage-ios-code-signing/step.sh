#!/bin/env bash

set -ex
cd "$BITRISE_SOURCE_DIR/e2e/step-manage-ios-code-signing" || exit
path="$HOME/.bitrise/toolkits/go/cache/path-._e2e_step-manage-ios-code-signing"
go build -o "$path"

cd "$BITRISE_SOURCE_DIR" || exit
$path