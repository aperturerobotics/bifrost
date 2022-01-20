module github.com/aperturerobotics/bifrost

go 1.16

// aperture: use compatibility forks
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/libp2p/go-libp2p-core => github.com/paralin/go-libp2p-core v0.12.1-0.20211209071220-3b91008fd2c4 // aperture
	github.com/libp2p/go-libp2p-tls => github.com/paralin/go-libp2p-tls v0.3.1-0.20211020072724-21716cf18549 // js-compat
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.23.1-0.20210907061838-0a0338bd72f0 // aperture
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831101324-59acc8fe7f74 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	github.com/paralin/kcp-go-lite => github.com/paralin/kcp-go-lite v1.0.2-0.20210907043027-271505668bd0 // aperture
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => github.com/paralin/grpc-go v1.30.1-0.20210804030014-1587a7c16b66 // aperture
	storj.io/drpc => github.com/paralin/drpc v0.0.27-0.20220104045627-466c7ca18e92 // aperture
)

require (
	github.com/aperturerobotics/controllerbus v0.8.7-0.20220120083918-73c808e9d611
	github.com/aperturerobotics/entitygraph v0.1.4-0.20210530040557-f19da9c2be6d
	github.com/aperturerobotics/timestamp v0.3.4
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/djherbis/buffer v1.2.0
	github.com/frankban/quicktest v1.14.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/hashicorp/yamux v0.0.0-20211028200310-0bc27b27de87
	github.com/klauspost/compress v1.13.6
	github.com/klauspost/reedsolomon v1.9.15 // indirect
	github.com/libp2p/go-libp2p-core v0.12.0
	github.com/libp2p/go-libp2p-tls v0.0.0-00010101000000-000000000000
	github.com/lucas-clemente/quic-go v0.0.0-00010101000000-000000000000
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.4.1
	github.com/nats-io/nats-server/v2 v2.0.0-00010101000000-000000000000
	github.com/nats-io/nats.go v0.0.0-00010101000000-000000000000
	github.com/nats-io/nkeys v0.3.0
	github.com/paralin/kcp-go-lite v1.0.2-0.20210907043027-271505668bd0 // aperture
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b
	github.com/tjfoc/gmsm v1.4.1
	github.com/urfave/cli v1.22.5
	github.com/xtaci/smux v1.5.16
	github.com/zeebo/blake3 v0.2.1
	github.com/zeebo/xxh3 v1.0.2-0.20211113223132-d10cc761c3ac // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	google.golang.org/grpc v1.31.0
	google.golang.org/protobuf v1.27.1 // indirect
	nhooyr.io/websocket v1.8.7
	storj.io/drpc v0.0.26
)
