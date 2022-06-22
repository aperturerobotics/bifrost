PROTOWRAP=hack/bin/protowrap
PROTOC_GEN_GO=hack/bin/protoc-gen-go
PROTOC_GEN_GO_DRPC=hack/bin/protoc-gen-go-drpc
PROTOC_GEN_VTPROTO=hack/bin/protoc-gen-go-vtproto
GOIMPORTS=hack/bin/goimports
GOLANGCI_LINT=hack/bin/golangci-lint
GO_MOD_OUTDATED=hack/bin/go-mod-outdated
GOLIST=go list -f "{{ .Dir }}" -m

export GO111MODULE=on
undefine GOOS
undefine GOARCH

all:

vendor:
	go mod vendor

$(PROTOC_GEN_GO):
	cd ./hack; \
	go build -v \
		-o ./bin/protoc-gen-go \
		google.golang.org/protobuf/cmd/protoc-gen-go

$(PROTOC_GEN_VTPROTO):
	cd ./hack; \
	go build -v \
		-o ./bin/protoc-gen-go-vtproto \
		github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto

$(PROTOC_GEN_GO_DRPC):
	cd ./hack; \
	go build -v \
		-o ./bin/protoc-gen-go-drpc \
		storj.io/drpc/cmd/protoc-gen-go-drpc

$(GOIMPORTS):
	cd ./hack; \
	go build -v \
		-o ./bin/goimports \
		golang.org/x/tools/cmd/goimports

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

$(GO_MOD_OUTDATED):
	cd ./hack; \
	go build -v \
		-o ./bin/go-mod-outdated \
		github.com/psampaz/go-mod-outdated

.PHONY: gengo
gengo: $(GOIMPORTS) $(PROTOWRAP) $(PROTOC_GEN_GO) $(PROTOC_GEN_VTPROTO) $(PROTOC_GEN_GO_DRPC) vendor
	go mod vendor
	shopt -s globstar; \
	set -eo pipefail; \
	export PROJECT=$$(go list -m); \
	export PATH=$$(pwd)/hack/bin:$${PATH}; \
	mkdir -p $$(pwd)/vendor/$$(dirname $${PROJECT}); \
	rm $$(pwd)/vendor/$${PROJECT} || true; \
	ln -s $$(pwd) $$(pwd)/vendor/$${PROJECT} ; \
	$(PROTOWRAP) \
		-I $$(pwd)/vendor \
		--go_out=$$(pwd)/vendor \
		--go-vtproto_out=$$(pwd)/vendor \
		--go-vtproto_opt=features=marshal+unmarshal+size \
		--go-drpc_out=$$(pwd)/vendor \
		--go-drpc_opt=json=false \
		--go-drpc_opt=protolib=github.com/planetscale/vtprotobuf/codec/drpc \
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

.PHONY: list
list: $(GO_MOD_OUTDATED)
	go list -mod=mod -u -m -json all | $(GO_MOD_OUTDATED)

.PHONY: outdated
outdated: $(GO_MOD_OUTDATED)
	go list -mod=mod -u -m -json all | $(GO_MOD_OUTDATED) -update -direct

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

.PHONY: fix
fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

.PHONY: test
test:
	go test -v ./...
