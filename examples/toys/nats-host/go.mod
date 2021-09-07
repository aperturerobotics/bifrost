module github.com/aperturerobotics/bifrost/examples/toys/nats-host

go 1.16

replace github.com/aperturerobotics/bifrost => ../../../

// aperture: use compatibility forks
replace (
	github.com/golang/protobuf => github.com/aperturerobotics/go-protobuf-1.3.x v0.0.0-20200726220404-fa7f51c52df0 // aperture-1.3.x
	github.com/lucas-clemente/quic-go => github.com/aperturerobotics/quic-go v0.23.1-0.20210907061838-0a0338bd72f0 // aperture
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20200831101324-59acc8fe7f74 // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => github.com/paralin/grpc-go v1.30.1-0.20210804030014-1587a7c16b66 // aperture
)

require (
	github.com/aperturerobotics/bifrost v0.0.0-20210903191425-d5a788a36744
	github.com/aperturerobotics/controllerbus v0.8.6-0.20210902104809-9f0fc115965e
	github.com/aperturerobotics/network-sim v0.0.0-20210903194329-4a8b6801f3c1
	github.com/blang/semver v3.5.1+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
)
