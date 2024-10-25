#!/bin/env bash

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$THIS_SCRIPT_DIR" || exit
path="$HOME/.bitrise/toolkits/go/cache/path-._e2e_step-manage-ios-code-signing"
go build -o "$path"
$path

