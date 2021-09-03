module github.com/aperturerobotics/bifrost/examples/toys/nats-host

go 1.16

replace github.com/aperturerobotics/bifrost => ../../../

// aperture: use protobuf 1.3.x based fork for compatibility
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.22.1-0.20210728081144-c7bd4637cac2 // aperture-protobuf-1.3.x
	github.com/sirupsen/logrus => github.com/paralin/logrus v1.0.6-0.20201109071409-0c5b7319eff6 // gopherjs-compat
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)

// aperture: use aperture-2.0 branch of fork
replace (
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831101324-59acc8fe7f74 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
)

require (
	github.com/aperturerobotics/bifrost v0.0.0-20210518124802-ee5cb0d80eb3
	github.com/aperturerobotics/controllerbus v0.8.6-0.20210902104809-9f0fc115965e
	github.com/aperturerobotics/network-sim v0.0.0-20210518125026-4fc2522f78e6
	github.com/blang/semver v3.5.1+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
)
