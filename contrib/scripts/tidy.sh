#!/usr/bin/env bash

set -e

modules=("./core" "./parse" ".")

for mod in "${modules[@]}"; do
    pushd "$mod" > /dev/null || {
        echo "Could not change to directory: $mod"
        exit 1
    }

    go mod tidy

    STATUS=$(git status go.mod go.sum --porcelain)

    if [[ -n "$STATUS" ]]; then
        echo "Changes detected in 'go.mod' or 'go.sum' after running 'go mod tidy' in module \"${mod}\"!"
        popd > /dev/null
        exit 1
    else
        echo "No changes detected in 'go.mod' or 'go.sum' in module \"${mod}\"."
    fi

    popd > /dev/null
done

exit 0
