module github.com/aperturerobotics/bifrost/examples/nats-host

go 1.16

replace github.com/aperturerobotics/bifrost => ../../

require (
	github.com/aperturerobotics/bifrost v0.7.9-0.20221017070424-77ff84bf6130
	github.com/aperturerobotics/controllerbus v0.18.2-0.20221019080900-fe522f5faeb9
)

// aperture: use compatibility forks
replace (
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.28.2-0.20220816034953-16dc6b89a8f8 // aperture
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831101324-59acc8fe7f74 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	github.com/paralin/kcp-go-lite => github.com/paralin/kcp-go-lite v1.0.2-0.20210907043027-271505668bd0 // aperture
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.8.2-0.20220322010420-77ab346a2cf8 // aperture
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.28.2-0.20221007002036-6510dd3bc392 // aperture
	nhooyr.io/websocket => github.com/paralin/nhooyr-websocket v1.8.8-0.20220321125022-7defdf942f07 // aperture
	storj.io/drpc => github.com/paralin/drpc v0.0.31-0.20220527065730-0e2a1370bccb // aperture
)

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/sirupsen/logrus v1.9.0
	github.com/urfave/cli/v2 v2.17.1
)
