module github.com/aperturerobotics/bifrost

go 1.13

replace github.com/multiformats/go-multihash => github.com/paralin/go-multihash v0.0.0-20190927235035-24ce17a9c4f3 // gopherjs-compat

require (
	github.com/aperturerobotics/controllerbus v0.1.6-0.20191002033011-c0e6f06edefd
	github.com/aperturerobotics/entitygraph v0.1.2-0.20190927211258-1d6a1c008f98
	github.com/aperturerobotics/timestamp v0.2.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.0.0-20191010011042-988181ef23fa // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/djherbis/buffer v1.1.0
	github.com/gogo/protobuf v1.3.1-0.20191006172112-69adf3ecd52d // indirect
	github.com/golang/protobuf v1.3.3-0.20190920234318-1680a479a2cf
	github.com/golang/snappy v0.0.2-0.20190904063534-ff6b7dc882cf
	github.com/gopherjs/gopherjs v0.0.0-20190915194858-d3ddacdb130f
	github.com/gopherjs/websocket v0.0.0-20170522004412-87ee47603f13
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d
	github.com/klauspost/compress v1.8.6
	github.com/klauspost/cpuid v1.2.2-0.20190713094507-cf2ded4aa833 // indirect
	github.com/klauspost/reedsolomon v1.9.3 // indirect
	github.com/libp2p/go-libp2p-core v0.2.4-0.20190930170843-2f75277a8d7b
	github.com/libp2p/go-libp2p-tls v0.1.1
	github.com/lucas-clemente/quic-go v0.12.1
	github.com/mr-tron/base58 v1.1.2
	github.com/multiformats/go-multiaddr v0.1.1
	github.com/multiformats/go-multiaddr-net v0.1.1
	github.com/paralin/kcp-go-lite v1.0.2-0.20190909213738-b58bf160d159
	github.com/patrickmn/go-cache v2.1.1-0.20191004192108-46f407853014+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4 v2.0.5+incompatible
	github.com/pkg/errors v0.8.2-0.20190227000051-27936f6d90f9
	github.com/sirupsen/logrus v1.4.2
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/xor v0.0.0-20181023030647-4e92f724b73b
	github.com/tjfoc/gmsm v1.0.2-0.20190417070453-18fd8096dc8a
	github.com/urfave/cli v1.21.0
	github.com/xtaci/smux v1.1.1
	golang.org/x/crypto v0.0.0-20190923035154-9ee001bba392
	google.golang.org/grpc v1.24.0
	gortc.io/stun v1.21.1
)
