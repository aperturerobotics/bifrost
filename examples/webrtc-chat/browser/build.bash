set -e

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

GOOS=js GOARCH=wasm go build -o example.wasm ./main.go

echo "Build completed successfully"
