module github.com/aperturerobotics/bifrost/examples/nats-host

go 1.16

replace github.com/aperturerobotics/bifrost => ../../

require github.com/aperturerobotics/controllerbus v0.11.2-0.20220709003231-eddf61b76ad2

// aperture: use compatibility forks
replace (
	github.com/libp2p/go-libp2p-core => github.com/paralin/go-libp2p-core v0.19.1-0.20220702014232-392fe0626689 // aperture
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.28.1-0.20220710071647-c608532b9a6e // aperture
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831101324-59acc8fe7f74 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	github.com/paralin/kcp-go-lite => github.com/paralin/kcp-go-lite v1.0.2-0.20210907043027-271505668bd0 // aperture
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.8.2-0.20220322010420-77ab346a2cf8 // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.27.2-0.20220609075637-a1d116b0035f // aperture
	nhooyr.io/websocket => github.com/paralin/nhooyr-websocket v1.8.8-0.20220321125022-7defdf942f07 // aperture
	storj.io/drpc => github.com/paralin/drpc v0.0.31-0.20220527065730-0e2a1370bccb // aperture
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/sirupsen/logrus v1.8.2-0.20220112234510-85981c045988
	github.com/urfave/cli v1.22.9
)

require github.com/aperturerobotics/bifrost v0.0.0-00010101000000-000000000000
