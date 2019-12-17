module github.com/aperturerobotics/bifrost

go 1.13

replace github.com/multiformats/go-multihash => github.com/paralin/go-multihash v0.0.0-20190927235035-24ce17a9c4f3 // gopherjs-compat

require (
	github.com/akutz/memconn v0.1.0 // indirect
	github.com/aperturerobotics/controllerbus v0.1.7-0.20191216071838-7fbb24a1ac6d
	github.com/aperturerobotics/entitygraph v0.1.2
	github.com/aperturerobotics/timestamp v0.2.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/djherbis/buffer v1.1.0
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/golang/protobuf v1.3.3-0.20191022195553-ed6926b37a63
	github.com/golang/snappy v0.0.2-0.20190904063534-ff6b7dc882cf
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de
	github.com/gopherjs/websocket v0.0.0-20191103002815-9a42957e2b3a
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d
	github.com/klauspost/compress v1.9.1
	github.com/klauspost/cpuid v1.2.2-0.20191106130511-5a626f7029c9 // indirect
	github.com/klauspost/reedsolomon v1.9.3 // indirect
	github.com/libp2p/go-libp2p-core v0.3.1-0.20191214080825-6f2516674ace
	github.com/libp2p/go-libp2p-tls v0.1.2
	github.com/lucas-clemente/quic-go v0.14.1
	github.com/mr-tron/base58 v1.1.3
	github.com/multiformats/go-multiaddr v0.2.0
	github.com/multiformats/go-multiaddr-net v0.1.1
	github.com/paralin/kcp-go-lite v1.0.2-0.20190927004254-2be397fe467b
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4 v2.3.0+incompatible
	github.com/pkg/errors v0.8.2-0.20191109102316-7f95ac13edff
	github.com/sirupsen/logrus v1.4.2
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/xor v0.0.0-20181023030647-4e92f724b73b
	github.com/tjfoc/gmsm v1.0.2-0.20191016014117-71e91645d4e3
	github.com/urfave/cli v1.22.2-0.20191024042601-850de854cda0
	github.com/xtaci/smux v1.4.6
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	google.golang.org/grpc v1.24.0
	gortc.io/stun v1.21.1
)
