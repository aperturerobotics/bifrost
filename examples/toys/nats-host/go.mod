module github.com/aperturerobotics/bifrost/examples/toys/nats-host

go 1.14

replace github.com/aperturerobotics/bifrost => ../../../

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.7.1-0.20200823084006-3bf6fe7f6a79 // aperture-protobuf-1.3.x
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

// aperture: use aperture-2.0 branch of fork
replace (
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831100512-248bd2ddb696 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v0.0.0-20200828220730-1cfd0cf50367 // aperture-2.0
)

require (
	github.com/aperturerobotics/bifrost v0.0.0-20200823084156-e28df1d443b1
	github.com/aperturerobotics/controllerbus v0.8.1-0.20200802060256-360612dc3698
	github.com/aperturerobotics/network-sim v0.0.0-20200729022251-c98cf1113722
	github.com/blang/semver v3.5.1+incompatible
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/sirupsen/logrus v1.6.0
	github.com/urfave/cli v1.22.4
)
