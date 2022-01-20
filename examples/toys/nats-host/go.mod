module github.com/aperturerobotics/bifrost/examples/toys/nats-host

go 1.16

replace github.com/aperturerobotics/bifrost => ../../../

require github.com/aperturerobotics/network-sim v0.0.0-20220108104750-72cf864be658

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
	github.com/aperturerobotics/bifrost v0.0.0-20220107035548-249f4b1b999f
	github.com/aperturerobotics/controllerbus v0.8.7-0.20220120083918-73c808e9d611
	github.com/blang/semver v3.5.1+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
)
