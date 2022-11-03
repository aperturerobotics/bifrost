module github.com/aperturerobotics/bifrost

go 1.18

require (
	github.com/aperturerobotics/controllerbus v0.18.2-0.20221019082518-c9323cca3baa
	github.com/aperturerobotics/entitygraph v0.3.1
	github.com/aperturerobotics/starpc v0.13.2-0.20221102012118-4a398033e50b
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
	github.com/aperturerobotics/ts-proto-common-types v0.2.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/djherbis/buffer v1.2.0
	github.com/golang/snappy v0.0.4
	github.com/hashicorp/yamux v0.1.1
	github.com/klauspost/compress v1.15.11
	github.com/libp2p/go-libp2p v0.23.3-0.20221025171119-7465a509d011
	github.com/lucas-clemente/quic-go v0.30.0
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.7.0
	github.com/nats-io/nats-server/v2 v2.7.4
	github.com/nats-io/nats.go v1.13.0
	github.com/nats-io/nkeys v0.3.0
	github.com/paralin/kcp-go-lite v4.3.4+incompatible
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/planetscale/vtprotobuf v0.3.0
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/sirupsen/logrus v1.9.0
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b
	github.com/tjfoc/gmsm v1.4.1
	github.com/urfave/cli/v2 v2.23.0
	github.com/xtaci/smux v1.5.16
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/crypto v0.1.1-0.20221024173537-a3485e174077
	gonum.org/v1/gonum v0.12.0
	google.golang.org/protobuf v1.28.1
	nhooyr.io/websocket v1.8.8-0.20210410000328-8dee580a7f74
	storj.io/drpc v0.0.30
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/frankban/quicktest v1.14.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/ipfs/go-cid v0.3.2 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/klauspost/cpuid/v2 v2.1.1 // indirect
	github.com/klauspost/reedsolomon v1.9.2 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-openssl v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.1-0.20220919134236-1c09f2ab3ec1 // indirect
	github.com/marten-seemann/qtls-go1-18 v0.1.3 // indirect
	github.com/marten-seemann/qtls-go1-19 v0.1.1 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/minio/highwayhash v1.0.0 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multicodec v0.6.0 // indirect
	github.com/multiformats/go-multihash v0.2.1 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/nats-io/jwt/v2 v2.0.0-20200820224411-1e751ff168ab // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/templexxx/cpu v0.0.1 // indirect
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xorsimd v0.4.1 // indirect
	github.com/valyala/fastjson v1.6.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/zeebo/errs v1.2.2 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/exp v0.0.0-20221031165847-c99f073a8326 // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/net v0.1.1-0.20221101210122-c87783997501 // indirect
	golang.org/x/sys v0.1.1-0.20221025174404-572649812736 // indirect
	golang.org/x/time v0.1.1-0.20221020023724-80b9fac54d29 // indirect
	golang.org/x/tools v0.2.1-0.20221101210316-32e1cb7aeda1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)
