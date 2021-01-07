#!/bin/bash

echo "Check if Tuist is installed"
if command -v tuist; then
  echo "Tuist already installed."
  tuist update
  (cd / && tuist version)
else
  bash <(curl -Ls https://install.tuist.io)
  (cd / && tuist version)
fi

# Only fail script starting from here since 
# `tuist update` exists with an error code although
# it actually succeeds.
set -e

echo ""
echo "Check if local Tuist needs to be updated"
GLOBAL_TUIST=$(cd / && tuist version)
echo "Global tuist version: $GLOBAL_TUIST"

LOCAL_TUIST=$(cat .tuist-version)
echo "Local tuist version: $LOCAL_TUIST"

if [[ $GLOBAL_TUIST != $LOCAL_TUIST ]]; then
  echo ""
  echo "Tuist needs updating!"
  echo "$GLOBAL_TUIST" > .tuist-version
  echo "Tuist updated to: $GLOBAL_TUIST."
else 
  echo ""
  echo "Using latest tuist."
fi