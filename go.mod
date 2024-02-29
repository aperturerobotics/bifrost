module github.com/aperturerobotics/bifrost

go 1.21

require (
	github.com/aperturerobotics/controllerbus v0.36.0 // latest
	github.com/aperturerobotics/entitygraph v0.6.1
	github.com/aperturerobotics/starpc v0.27.2-0.20240229060046-28d111bca538 // latest
	github.com/aperturerobotics/timestamp v0.8.2
	github.com/aperturerobotics/ts-proto-common-types v0.20.2 // latest
	github.com/aperturerobotics/util v1.14.3 // master
)

// aperture: use compatibility forks
replace (
	github.com/multiformats/go-multiaddr => github.com/paralin/go-multiaddr v0.10.2-0.20230807174004-e1767541c061 // aperture
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.0.0-20200820224411-1e751ff168ab // indirect: used by bifrost-nats-server
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20221228081037-b7c2df0c151f // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	github.com/nats-io/nkeys => github.com/nats-io/nkeys v0.3.0 // indirect: used by bifrost-nats-server
	github.com/paralin/kcp-go-lite => github.com/paralin/kcp-go-lite v1.0.2-0.20240110212540-c015327ec623 // aperture
	github.com/quic-go/quic-go => github.com/aperturerobotics/quic-go v0.41.1-0.20240125035303-1093432c45e9 // aperture
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.4-0.20240119050608-13332fb58195 // aperture
	go.bug.st/serial => github.com/paralin/go-serial v0.0.0-20240220235126-6223594b0314 // fix-wasm-1
	google.golang.org/protobuf => github.com/aperturerobotics/protobuf-go v1.32.1-0.20240118233629-6aeee82c476f // aperture
	nhooyr.io/websocket => github.com/paralin/nhooyr-websocket v1.8.8-0.20220321125022-7defdf942f07 // aperture
	storj.io/drpc => github.com/paralin/drpc v0.0.31-0.20220527065730-0e2a1370bccb // aperture
)

require (
	filippo.io/edwards25519 v1.1.1-0.20231210192602-a7dfd8e4e6b4
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/djherbis/buffer v1.2.0
	github.com/golang/snappy v0.0.4
	github.com/klauspost/compress v1.17.7
	github.com/libp2p/go-libp2p v0.33.0
	github.com/libp2p/go-yamux/v4 v4.0.2-0.20240206065824-7222fbc3459d
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.12.2
	github.com/nats-io/nats-server/v2 v2.10.11
	github.com/nats-io/nats.go v1.33.1
	github.com/nats-io/nkeys v0.4.7
	github.com/paralin/kcp-go-lite v1.0.2-0.20240110212540-c015327ec623
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pierrec/lz4/v4 v4.1.21
	github.com/pion/datachannel v1.5.5
	github.com/pion/sdp/v3 v3.0.6
	github.com/pion/webrtc/v4 v4.0.0-beta.11
	github.com/pkg/errors v0.9.1
	github.com/planetscale/vtprotobuf v0.6.0
	github.com/quic-go/quic-go v0.41.0
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/templexxx/xor v0.0.0-20191217153810-f85b25db303b
	github.com/urfave/cli/v2 v2.27.1
	github.com/valyala/fastjson v1.6.4
	github.com/xtaci/smux v1.5.24
	github.com/zeebo/blake3 v0.2.3
	go.bug.st/serial v1.6.2
	golang.org/x/crypto v0.20.0
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	gonum.org/v1/gonum v0.14.0
	google.golang.org/protobuf v1.32.0
	nhooyr.io/websocket v1.8.10
	storj.io/drpc v0.0.33
)

require (
	github.com/Jeffail/gabs/v2 v2.7.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/pprof v0.0.0-20240207164012-fb44976bdcd5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/klauspost/reedsolomon v1.12.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/nats-io/jwt/v2 v2.4.1 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/ginkgo/v2 v2.15.0 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pion/dtls/v2 v2.2.10 // indirect
	github.com/pion/ice/v3 v3.0.3 // indirect
	github.com/pion/interceptor v0.1.25 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.12 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.13 // indirect
	github.com/pion/rtp v1.8.3 // indirect
	github.com/pion/sctp v1.8.12 // indirect
	github.com/pion/srtp/v3 v3.0.1 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/pion/transport/v3 v3.0.1 // indirect
	github.com/pion/turn/v3 v3.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/templexxx/cpu v0.1.0 // indirect
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xorsimd v0.4.2 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/zeebo/errs v1.2.2 // indirect
	go.uber.org/mock v0.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.15.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.18.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
)
