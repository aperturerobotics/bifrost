PROTOWRAP=hack/bin/protowrap
PROTOC_GEN_GO=hack/bin/protoc-gen-go
PROTOC_GEN_GO_DRPC=hack/bin/protoc-gen-go-drpc
GOIMPORTS=hack/bin/goimports
GOLANGCI_LINT=hack/bin/golangci-lint
GOLIST=go list -f "{{ .Dir }}" -m
export GO111MODULE=on

all:

vendor:
	go mod vendor

$(PROTOC_GEN_GO):
	cd ./hack; \
	go build -v \
		-o ./bin/protoc-gen-go \
		github.com/golang/protobuf/protoc-gen-go

$(GOIMPORTS):
	cd ./hack; \
	go build -v \
		-o ./bin/goimports \
		golang.org/x/tools/cmd/goimports

$(PROTOC_GEN_GO_DRPC):
	cd ./hack; \
	go build -v \
		-o ./bin/protoc-gen-go-drpc \
		storj.io/drpc/cmd/protoc-gen-go-drpc

$(PROTOWRAP):
	cd ./hack; \
	go build -v \
		-o ./bin/protowrap \
		github.com/square/goprotowrap/cmd/protowrap

$(GOLANGCI_LINT):
	cd ./hack; \
	go build -v \
		-o ./bin/golangci-lint \
		github.com/golangci/golangci-lint/cmd/golangci-lint

gengo: $(GOIMPORTS) $(PROTOWRAP) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_DRPC) vendor
	shopt -s globstar; \
	set -eo pipefail; \
	export PROJECT=$$(go list -m); \
	export PATH=$$(pwd)/hack/bin:$${PATH}; \
	mkdir -p $$(pwd)/vendor/$$(dirname $${PROJECT}); \
	rm $$(pwd)/vendor/$${PROJECT} || true; \
	ln -s $$(pwd) $$(pwd)/vendor/$${PROJECT} ; \
	$(PROTOWRAP) \
		-I $$(pwd)/vendor \
		--go_out=plugins=grpc:$$(pwd)/vendor \
		--go-drpc_out=$$(pwd)/vendor \
    --go-drpc_opt=json=false \
    --go-drpc_opt=protolib=github.com/golang/protobuf/proto \
		--proto_path $$(pwd)/vendor \
		--print_structure \
		--only_specified_files \
		$$(\
			git \
				ls-files "*.proto" |\
				xargs printf -- \
				"$$(pwd)/vendor/$${PROJECT}/%s "); \
	rm $$(pwd)/vendor/$${PROJECT} || true
	go mod vendor
	$(GOIMPORTS) -w ./

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...


