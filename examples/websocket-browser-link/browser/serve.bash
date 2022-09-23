#!/bin/bash
set -eo pipefail

# cd to project root
if [ ! -d ./hack ]; then
    if [ -d ../../../hack ]; then
        cd ../../../
    else
        cd $(git rev-parse --show-toplevel)
    fi
fi

# serve the example
echo "Listening on port :8080"
make serve-example
