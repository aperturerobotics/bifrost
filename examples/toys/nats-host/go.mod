module github.com/aperturerobotics/bifrost/examples/toys/nats-host

go 1.14

replace github.com/aperturerobotics/bifrost => ../../../

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200706003739-05fb54d407a9 // aperture-1.3.x
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.7.1-0.20200728021714-7db2bdfa8cd7 // aperture-protobuf-1.3.x
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

// aperture: use aperture-2.0 branch of fork
replace github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.0.0-20200728011410-af6fb29263b2 // aperture-2.0

require (
	github.com/aperturerobotics/bifrost v0.0.0-20200728210917-e78147c3a9dd
	github.com/aperturerobotics/controllerbus v0.7.1-0.20200729041112-99c388e5c35a
	github.com/aperturerobotics/hydra v0.0.0-20200729004213-b1dda6b62e4b
	github.com/aperturerobotics/network-sim v0.0.0-20200729022251-c98cf1113722
	github.com/blang/semver v3.5.1+incompatible
	github.com/libp2p/go-libp2p-core v0.6.0
	github.com/sirupsen/logrus v1.6.0
	github.com/urfave/cli v1.22.4
)
