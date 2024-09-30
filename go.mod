module github.com/aperturerobotics/bifrost

go 1.22.0

toolchain go1.23.1

require (
	github.com/aperturerobotics/common v0.18.6 // latest
	github.com/aperturerobotics/controllerbus v0.47.7 // latest
	github.com/aperturerobotics/entitygraph v0.10.0 // latest
	github.com/aperturerobotics/protobuf-go-lite v0.7.0 // latest
	github.com/aperturerobotics/starpc v0.33.11 // latest
	github.com/aperturerobotics/util v1.25.9 // latest
)

// aperture: use compatibility forks
replace (
	github.com/ipfs/go-log/v2 => github.com/paralin/ipfs-go-logrus v0.0.0-20240410105224-e24cb05f9e98 // master
	github.com/libp2p/go-libp2p => github.com/aperturerobotics/go-libp2p v0.33.1-0.20240511223728-e0b67c111765 // aperture
	github.com/libp2p/go-msgio => github.com/aperturerobotics/go-libp2p-msgio v0.0.0-20240511033615-1b69178aa5c8 // aperture
	github.com/multiformats/go-multiaddr => github.com/aperturerobotics/go-multiaddr v0.12.4-0.20240407071906-6f0354cc6755 // aperture
	github.com/multiformats/go-multihash => github.com/aperturerobotics/go-multihash v0.2.3 // aperture
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.0.0-20200820224411-1e751ff168ab // indirect: used by bifrost-nats-server
	github.com/nats-io/nats-server/v2 => github.com/aperturerobotics/bifrost-nats-server/v2 v2.1.8-0.20221228081037-b7c2df0c151f // aperture-2.0
	github.com/nats-io/nats.go => github.com/aperturerobotics/bifrost-nats-client v1.10.1-0.20200831103200-24c3d0464e58 // aperture-2.0
	github.com/nats-io/nkeys => github.com/nats-io/nkeys v0.3.0 // indirect: used by bifrost-nats-server
	github.com/quic-go/quic-go => github.com/aperturerobotics/quic-go v0.47.1-0.20240930074049-d9f13f608464 // aperture
	github.com/sirupsen/logrus => github.com/aperturerobotics/logrus v1.9.4-0.20240119050608-13332fb58195 // aperture
)

require (
	filippo.io/edwards25519 v1.1.1-0.20231210192602-a7dfd8e4e6b4
	github.com/blang/semver/v4 v4.0.0 // latest
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/coder/websocket v1.8.13-0.20240919094401-cef8e11d00b0 // master
	github.com/klauspost/compress v1.17.10
	github.com/libp2p/go-libp2p v0.36.4
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.13.0
	github.com/nats-io/nats-server/v2 v2.10.21
	github.com/nats-io/nats.go v1.37.0
	github.com/nats-io/nkeys v0.4.7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pion/datachannel v1.5.9
	github.com/pion/sdp/v3 v3.0.9
	github.com/pion/webrtc/v4 v4.0.0-beta.30
	github.com/pkg/errors v0.9.1
	github.com/quic-go/quic-go v0.47.0 // latest
	github.com/sasha-s/go-deadlock v0.3.5
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.27.4
	github.com/zeebo/blake3 v0.2.4
	golang.org/x/crypto v0.27.0
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0
	gonum.org/v1/gonum v0.15.1
)

require (
	github.com/aperturerobotics/json-iterator-lite v1.0.0 // indirect
	github.com/cloudflare/circl v1.3.8 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20240207164012-fb44976bdcd5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-yamux/v4 v4.0.2-0.20240826150533-e92055b23e0e // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.5.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/nats-io/jwt/v2 v2.4.1 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/ginkgo/v2 v2.15.0 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pion/dtls/v3 v3.0.2 // indirect
	github.com/pion/ice/v4 v4.0.1 // indirect
	github.com/pion/interceptor v0.1.30 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.14 // indirect
	github.com/pion/rtp v1.8.9 // indirect
	github.com/pion/sctp v1.8.33 // indirect
	github.com/pion/srtp/v3 v3.0.3 // indirect
	github.com/pion/stun/v3 v3.0.0 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v4 v4.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.1-0.20190317074736-539464a789e9 // indirect
	github.com/wlynxg/anet v0.0.3 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	lukechampine.com/blake3 v1.2.1 // indirect
)
