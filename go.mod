module github.com/aperturerobotics/bifrost

go 1.13

replace github.com/multiformats/go-multihash => github.com/paralin/go-multihash v0.0.0-20190831070958-91cde46649b8 // gopherjs-compat

require (
	github.com/aperturerobotics/controllerbus v0.1.5
	github.com/aperturerobotics/entitygraph v0.1.1-0.20190909222015-b58513aa9083
	github.com/aperturerobotics/timestamp v0.2.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.0.0-20190824003749-130ea5bddde3 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/djherbis/buffer v1.1.0
	github.com/gogo/protobuf v1.3.1-0.20190908201246-8a5ed79f6888 // indirect
	github.com/golang/protobuf v1.3.3-0.20190920234318-1680a479a2cf
	github.com/golang/snappy v0.0.1
	github.com/gopherjs/gopherjs v0.0.0-20190915194858-d3ddacdb130f
	github.com/gopherjs/websocket v0.0.0-20170522004412-87ee47603f13
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d
	github.com/klauspost/compress v1.8.1
	github.com/klauspost/cpuid v1.2.2-0.20190713094507-cf2ded4aa833 // indirect
	github.com/klauspost/reedsolomon v1.9.3-0.20190625143044-789068412913 // indirect
	github.com/libp2p/go-libp2p-core v0.2.3-0.20190828160545-b74f60b9cc2b
	github.com/libp2p/go-libp2p-tls v0.1.1
	github.com/lucas-clemente/quic-go v0.12.0
	github.com/mr-tron/base58 v1.1.2
	github.com/multiformats/go-multiaddr v0.1.0
	github.com/multiformats/go-multiaddr-net v0.1.0
	github.com/paralin/kcp-go-lite v1.0.2-0.20190909213738-b58bf160d159
	github.com/patrickmn/go-cache v2.1.1-0.20180815053127-5633e0862627+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4 v2.0.5+incompatible
	github.com/pkg/errors v0.8.2-0.20190227000051-27936f6d90f9
	github.com/sirupsen/logrus v1.4.2
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/tjfoc/gmsm v1.0.2-0.20190220015903-f915c2cebf58 // indirect
	github.com/urfave/cli v1.21.0
	github.com/xtaci/smux v1.1.1
	golang.org/x/crypto v0.0.0-20190909091759-094676da4a83
	google.golang.org/grpc v1.24.0
	gortc.io/stun v1.21.1
)
