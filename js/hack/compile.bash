#!/bin/bash
set -eo pipefail

# compile.bash
# Execute in the Bifrost root as PWD.
if [ ! -d ./handshake ]; then
    echo "Execute in bifrost root directory."
    exit 1
fi

if [ ! -d ../js-go-compiler ]; then
    echo "Execute in the js repository"
    exit 1
fi

# Compile node
if [ -d ./vendor ]; then
    echo "Skipping vendor restore"
else
    echo "Restoring deps..."
    dep ensure -v -vendor-only
fi

echo "Compiling gopherjs..."
pushd ../js-go-compiler
./hack/build_gopherjs.bash
popd

echo "Setting up environment..."
source ../js-go-compiler/hack/env.bash
mkdir -p ./gopath/src/github.com/aperturerobotics
ln -fs $(pwd) ./gopath/src/github.com/aperturerobotics/bifrost
export GOPATH=$(pwd)/gopath
pushd ./gopath/src/github.com/aperturerobotics/bifrost

echo "Compiling with GopherJS..."
pushd ./handshake/identity/s2s/e2e
gopherjs build ./
popd

popd
echo "Done!"
