module github.com/aperturerobotics/bifrost

go 1.22

require (
	github.com/aperturerobotics/common v0.15.2 // latest
	github.com/aperturerobotics/controllerbus v0.45.0 // latest
	github.com/aperturerobotics/entitygraph v0.9.1 // latest
	github.com/aperturerobotics/protobuf-go-lite v0.6.1 // latest
	github.com/aperturerobotics/starpc v0.32.3 // latest
	github.com/aperturerobotics/util v1.23.0 // master
)

// aperture: use compatibility forks
replace (
	github.com/ipfs/go-log/v2 => github.com/paralin/ipfs-go-logrus v0.0.0-20240410105224-e24cb05f9e98 // master
	github.com/libp2p/go-libp2p => github.com/aperturerobotics/go-libp2p v0.33.1-0.20240504075939-591fc65373be // aperture
	github.com/multiformats/go-multiaddr => github.com/paralin/go-multiaddr v0.12.4-0.20240407071906-6f0354cc6755 // aperture-4
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.0.0-20200820224411-1e751ff168ab // indirect: used by bifrost-nats-server
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20221228081037-b7c2df0c151f // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	github.com/nats-io/nkeys => github.com/nats-io/nkeys v0.3.0 // indirect: used by bifrost-nats-server
	github.com/quic-go/quic-go => github.com/aperturerobotics/quic-go v0.43.1-0.20240504081906-25e38f065e10 // aperture
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.4-0.20240119050608-13332fb58195 // aperture
)

require (
	filippo.io/edwards25519 v1.1.1-0.20231210192602-a7dfd8e4e6b4
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/klauspost/compress v1.17.8
	github.com/libp2p/go-libp2p v0.33.2
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.12.3
	github.com/nats-io/nats-server/v2 v2.10.14
	github.com/nats-io/nats.go v1.34.1
	github.com/nats-io/nkeys v0.4.7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pion/datachannel v1.5.6
	github.com/pion/sdp/v3 v3.0.9
	github.com/pion/webrtc/v4 v4.0.0-beta.19
	github.com/pkg/errors v0.9.1
	github.com/quic-go/quic-go v0.43.0
	github.com/sasha-s/go-deadlock v0.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.27.2
	github.com/zeebo/blake3 v0.2.3
	golang.org/x/crypto v0.22.0
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f
	gonum.org/v1/gonum v0.15.0
	nhooyr.io/websocket v1.8.11 // master
)

require (
	github.com/aperturerobotics/json-iterator-lite v1.0.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20240207164012-fb44976bdcd5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2-0.20240322071716-53ef5820bd48 // indirect
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
	github.com/pion/ice/v3 v3.0.7 // indirect
	github.com/pion/interceptor v0.1.29 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.14 // indirect
	github.com/pion/rtp v1.8.6 // indirect
	github.com/pion/sctp v1.8.16 // indirect
	github.com/pion/srtp/v3 v3.0.1 // indirect
	github.com/pion/stun/v2 v2.0.0 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/pion/transport/v3 v3.0.2 // indirect
	github.com/pion/turn/v3 v3.0.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.20.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
)
