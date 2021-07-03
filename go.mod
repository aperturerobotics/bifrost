module github.com/aperturerobotics/bifrost

go 1.16

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.7.1-0.20210518124640-25c39ec20d1d // aperture-protobuf-1.3.x
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

// aperture: use aperture-2.0 branch of fork
replace (
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831101324-59acc8fe7f74 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
)

require (
	github.com/aperturerobotics/controllerbus v0.8.2
	github.com/aperturerobotics/entitygraph v0.1.4-0.20210530040557-f19da9c2be6d
	github.com/aperturerobotics/timestamp v0.2.4-0.20210530040952-1422410fbd4a
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v1.1.1-0.20190506075156-2146c9339422
	github.com/djherbis/buffer v1.1.0
	github.com/frankban/quicktest v1.10.2 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.3
	github.com/gopherjs/gopherjs v0.0.0-20210621113107-84c6004145de
	github.com/gopherjs/websocket v0.0.0-20191103002815-9a42957e2b3a
	github.com/gorilla/websocket v1.4.3-0.20210424162022-e8629af678b7
	github.com/hashicorp/yamux v0.0.0-20210316155119-a95892c5f864
	github.com/klauspost/compress v1.13.2-0.20210622115932-09f13c9b23f1
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-tls v0.1.3
	github.com/lucas-clemente/quic-go v0.21.1
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multiaddr-net v0.2.0
	github.com/nats-io/nats-server/v2 v2.3.1
	github.com/nats-io/nats.go v1.11.0
	github.com/nats-io/nkeys v0.3.0
	github.com/paralin/kcp-go-lite v1.0.2-0.20190927004254-2be397fe467b
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4 v2.6.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b
	github.com/tjfoc/gmsm v1.4.0
	github.com/urfave/cli v1.22.5
	github.com/xtaci/smux v1.5.15
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	google.golang.org/grpc v1.39.0
	gortc.io/stun v1.23.0
)
