module github.com/aperturerobotics/bifrost

require (
	github.com/aperturerobotics/controllerbus v0.0.0-20190121075103-ad53165fa70b
	github.com/aperturerobotics/entitygraph v0.0.0-20181226225716-1e77d0ca8bd7
	github.com/aperturerobotics/timestamp v0.2.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d // indirect
	github.com/btcsuite/btcutil v0.0.0-20190112041146-bf1e1be93589 // indirect
	github.com/btcsuite/goleveldb v1.0.0 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/djherbis/buffer v1.0.0
	github.com/gogo/protobuf v1.2.0 // indirect
	github.com/golang/mock v1.2.0 // indirect
	github.com/golang/protobuf v1.2.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db
	github.com/gopherjs/gopherjs v0.0.0-20181103185306-d547d1d9531e
	github.com/gopherjs/websocket v0.0.0-20170522004412-87ee47603f13
	github.com/gorilla/websocket v1.4.0
	github.com/gortc/stun v1.18.2
	github.com/gxed/hashland v0.0.0-20180221191214-d9f6b97f8db2 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d
	github.com/ipfs/go-log v1.5.7 // indirect
	github.com/jbenet/goprocess v0.0.0-20160826012719-b497e2f366b8 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/kkdai/bstream v0.0.0-20181106074824-b3251f7901ec // indirect
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.8.0 // indirect
	github.com/libp2p/go-libp2p-crypto v2.0.1+incompatible
	github.com/libp2p/go-libp2p-net v3.0.15+incompatible // indirect
	github.com/libp2p/go-libp2p-peer v2.4.0+incompatible
	github.com/libp2p/go-libp2p-peerstore v2.0.6+incompatible // indirect
	github.com/libp2p/go-libp2p-protocol v1.0.0 // indirect
	github.com/libp2p/go-libp2p-transport v3.0.15+incompatible
	github.com/libp2p/go-stream-muxer v3.0.1+incompatible // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mr-tron/base58 v1.1.0
	github.com/multiformats/go-multiaddr v1.4.0
	github.com/multiformats/go-multiaddr-dns v0.2.5 // indirect
	github.com/multiformats/go-multiaddr-net v1.7.1
	github.com/multiformats/go-multihash v1.0.8
	github.com/opentracing/opentracing-go v1.0.2 // indirect
	github.com/paralin/kcp-go-lite v4.3.2-0.20181125120641-d71c59f1ca69+incompatible
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pauleyj/gobee v0.0.0-20170221144200-48ad5f04527c
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spaolacci/murmur3 v0.0.0-20180118202830-f09979ecbc72 // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20181023030647-4e92f724b73b // indirect
	github.com/tjfoc/gmsm v1.0.1 // indirect
	github.com/urfave/cli v1.20.0
	github.com/whyrusleeping/go-logging v0.0.0-20170515211332-0457bb6b88fc // indirect
	github.com/xtaci/smux v1.1.0
	golang.org/x/crypto v0.0.0-20190122013713-64072686203f
	golang.org/x/lint v0.0.0-20181217174547-8f45f776aaf1 // indirect
	golang.org/x/net v0.0.0-20190119204137-ed066c81e75e
	golang.org/x/oauth2 v0.0.0-20190115181402-5dab4167f31c // indirect
	golang.org/x/sys v0.0.0-20190122071731-054c452bb702 // indirect
	golang.org/x/tools v0.0.0-20190121143147-24cd39ecf745 // indirect
	google.golang.org/genproto v0.0.0-20190111180523-db91494dd46c // indirect
	google.golang.org/grpc v1.18.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	honnef.co/go/tools v0.0.0-20190109154334-5bcec433c8ea // indirect
)

replace github.com/multiformats/go-multihash => github.com/paralin/go-multihash v0.0.0-20190110102829-0484db56787c

replace github.com/libp2p/go-libp2p-crypto => github.com/paralin/go-libp2p-crypto v0.0.0-20190110112134-4f99fef99f04
