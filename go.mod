module github.com/aperturerobotics/bifrost

replace github.com/multiformats/go-multihash => github.com/paralin/go-multihash v0.0.0-20190328060944-4974b52a2e95

replace github.com/libp2p/go-libp2p-crypto => github.com/paralin/go-libp2p-crypto v0.0.0-20181130162722-b150863d61f7

require (
	github.com/aperturerobotics/controllerbus v0.0.0-20190412141224-a86f75f58cec
	github.com/aperturerobotics/entitygraph v0.0.0-20190314052401-c4dff866fe8f
	github.com/aperturerobotics/timestamp v0.2.2-0.20190226083629-0175fc7d961e
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.0.0-20190629003639-c26ffa870fd8 // indirect
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/djherbis/buffer v1.0.1-0.20181027144806-3c732ee9b562
	github.com/gogo/protobuf v1.2.2-0.20190611061853-dadb62585089 // indirect
	github.com/golang/protobuf v1.3.2-0.20190701182201-6c65a5562fc0
	github.com/golang/snappy v0.0.1
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c
	github.com/gopherjs/websocket v0.0.0-20170522004412-87ee47603f13
	github.com/gorilla/websocket v1.4.1-0.20190629185528-ae1634f6a989
	github.com/gortc/stun v1.19.1-0.20190509220556-d73420a61edc
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/klauspost/reedsolomon v1.9.3-0.20190625143044-789068412913 // indirect
	github.com/libp2p/go-libp2p-crypto v0.0.1
	github.com/mr-tron/base58 v1.1.1
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-dns v0.0.2-0.20190321165136-7d0de25ce05c // indirect
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/multiformats/go-multihash v0.0.6
	github.com/paralin/kcp-go-lite v4.3.2-0.20190202132049-1e12d0a0fd45+incompatible
	github.com/patrickmn/go-cache v2.1.1-0.20180815053127-5633e0862627+incompatible
	github.com/pauleyj/gobee v0.0.0-20190212035730-6270c53072a4
	github.com/pkg/errors v0.8.2-0.20190227000051-27936f6d90f9
	github.com/sirupsen/logrus v1.4.0
	github.com/spaolacci/murmur3 v0.0.0-20180118202830-f09979ecbc72 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20181023030647-4e92f724b73b // indirect
	github.com/tjfoc/gmsm v1.0.2-0.20190220015903-f915c2cebf58 // indirect
	github.com/urfave/cli v1.20.1-0.20190203184040-693af58b4d51
	github.com/xtaci/smux v1.1.1
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	google.golang.org/grpc v1.19.0
)
