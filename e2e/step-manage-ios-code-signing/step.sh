#/bin/env bash

path="$HOME/.bitrise/toolkits/go/cache/path-._e2e_step-manage-ios-code-signing"
go build -o "$path"
$path

