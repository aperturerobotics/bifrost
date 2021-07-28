package s2s

import (
	"bytes"
	"context"
	"crypto/rand"
	"sync/atomic"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
	// "golang.org/x/crypto/nacl/secretbox"
	// "golang.org/x/crypto/ripemd160"
)

// Handshaker implements the Station to Station handshake protocol.
type Handshaker struct {
	// nInPkt is the number of handled packets
	nInPkt int32
	// writePacketFn writes a packet
	writePacketFn func(data []byte) error
	// lookupPubKey looks up a cached public key for the sender.
	// Returns nil if not found.
	lookupPubKey func(pid peer.ID) crypto.PubKey
	// privKey is the local private key.
	privKey crypto.PrivKey
	// hPub is the handshake ephemeral public key
	hPub *[32]byte
	// hPriv is the handshake ephemeral private key
	hPriv *[32]byte
	// packetCh handles an incoming packet
	packetCh chan []byte
	// localPeerID is the local peer id
	localPeerID peer.ID
	// initiator indicates we initiated the request
	initiator bool
	// extraData is the extra data field
	extraData []byte

	// remotePubKey is the remote public key if known
	remotePubKey crypto.PubKey
	// remotePeerID is the remote peer ID if known
	remotePeerID peer.ID
}

// NewHandshaker builds a new handshaker.
// ExpectedRemotePub speeds up the handshake by indicating an expected remote public key.
// It can be nil.
func NewHandshaker(
	privKey crypto.PrivKey,
	expectedRemotePub crypto.PubKey,
	writePacket func(data []byte) error,
	lookupPubKey func(pid peer.ID) crypto.PubKey,
	initiator bool,
	extraData []byte,
) (*Handshaker, error) {
	// hPub is the handshake public key.
	hPub, hPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	localPeerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	var remotePeerID peer.ID
	if expectedRemotePub != nil {
		remotePeerID, err = peer.IDFromPublicKey(expectedRemotePub)
		if err != nil {
			return nil, err
		}
	}

	return &Handshaker{
		privKey: privKey,

		hPriv:         hPriv,
		hPub:          hPub,
		writePacketFn: writePacket,
		localPeerID:   localPeerID,
		initiator:     initiator,
		remotePeerID:  remotePeerID,
		remotePubKey:  expectedRemotePub,
		packetCh:      make(chan []byte, 3),
		lookupPubKey:  lookupPubKey,
		extraData:     extraData,
	}, nil
}

func checkPacketType(actual, expected PacketType) error {
	if actual != expected {
		return errors.Errorf(
			"unexpected packet type: %s (expected %s)",
			actual.String(),
			expected.String(),
		)
	}

	return nil
}

// Execute executes the handshake with a context.
// Initiator indicates the handshaker is the initiator of the handshake.
// Returning an error cancels the attempt.
func (h *Handshaker) Execute(ctx context.Context) (*identity.Result, error) {
	initiator := h.initiator
	inPkt := &Packet{}
	var remoteEphPub [32]byte
	var sharedKey [32]byte

	// To save network overhead, the nonce simply increments 1, 2, 3 for each packet.
	// This is secure since a new keypair is used each time.
	var nonce [24]byte
	copy(nonce[:], []byte("bifrost"))
	nonce[10] = 0x5
	nextNonce := func() {
		nonce[len(nonce)-1]++
	}

	var rxExtraData []byte
	if initiator {
		if err := h.writeInit(); err != nil {
			return nil, err
		}

		// Read the init ack message.
		if err := h.readPacket(ctx, inPkt, PacketType_PacketType_INIT_ACK); err != nil {
			return nil, err
		}

		initAckMsg := inPkt.GetInitAckPkt()
		if err := initAckMsg.Validate(); err != nil {
			return nil, err
		}

		// Compute the shared secret
		copy(remoteEphPub[:], initAckMsg.GetSenderEphPub())
		box.Precompute(&sharedKey, &remoteEphPub, h.hPriv)

		// Decrypt the ciphertext
		cipherData, valid := box.OpenAfterPrecomputation(
			nil,
			initAckMsg.GetCiphertext(),
			&nonce,
			&sharedKey,
		)
		if !valid {
			return nil, errors.New("ciphertext failed to decrypt")
		}
		nextNonce()

		// Decode the data
		cipher := &Packet_Ciphertext{}
		if err := proto.Unmarshal(cipherData, cipher); err != nil {
			return nil, errors.Wrap(err, "unmarshal ciphertext")
		}

		sendPubKey := !cipher.GetReceiverKeyKnown()

		// Decode the public key from the cipher
		if err := h.decodePubKey(cipher); err != nil {
			return nil, err
		}
		rxExtraData = cipher.GetExtraInfo()

		// Compute the proof
		concatKeys := make([]byte, len(remoteEphPub)+len(*h.hPub))
		copy(concatKeys, (*h.hPub)[:])
		copy(concatKeys[len(remoteEphPub):], remoteEphPub[:])

		// Validate remote signature
		ok, err := h.remotePubKey.Verify(concatKeys, cipher.GetTupleSignature())
		if err != nil {
			return nil, errors.Wrap(err, "validate remote tuple signature")
		}
		if !ok {
			return nil, errors.New("remote tuple signature invalid")
		}

		// Reuse cipher
		cipher.Reset()
		cipher.ExtraInfo = h.extraData
		cipher.TupleSignature, err = h.privKey.Sign(concatKeys)
		if err != nil {
			return nil, errors.Wrap(err, "sign ephemeral keypair")
		}

		// Check sending the public key
		cipher.ReceiverKeyKnown = true
		if sendPubKey {
			b, err := crypto.MarshalPublicKey(h.privKey.GetPublic())
			if err != nil {
				return nil, errors.Wrap(err, "marshal local public key")
			}

			cipher.SenderPubKey = b
		}

		// Encrypt the ciphertext.
		cipherData, err = proto.Marshal(cipher)
		if err != nil {
			return nil, err
		}

		// Send the packet completing the handshake.
		ackPkt := &Packet_Complete{}
		ackPkt.Ciphertext = box.SealAfterPrecomputation(nil, cipherData, &nonce, &sharedKey)
		if err := h.writeComplete(ackPkt); err != nil {
			return nil, err
		}
	} else {
		// Read the init message.
		if err := h.readPacket(ctx, inPkt, PacketType_PacketType_INIT); err != nil {
			return nil, err
		}

		initMsg := inPkt.GetInitPkt()
		if err := initMsg.Validate(); err != nil {
			return nil, errors.Wrap(err, "init packet")
		}

		// Check if the sender key is known
		{
			remotePeerID := peer.ID(initMsg.GetSenderPeerId())
			if remotePeerID != h.remotePeerID {
				h.remotePeerID = remotePeerID
				h.remotePubKey = nil
			}

			remotePub, err := remotePeerID.ExtractPublicKey()
			if err != nil {
				return nil, errors.Wrap(err, "invalid remote peer id")
			}
			if remotePub == nil && h.lookupPubKey != nil {
				remotePub = h.lookupPubKey(remotePeerID)
			}
			if remotePub != nil {
				h.remotePubKey = remotePub
			}
		}

		ackCipher := &Packet_Ciphertext{}
		ackCipher.ReceiverKeyKnown = h.remotePubKey != nil
		ackCipher.ExtraInfo = h.extraData

		// Check the peer id
		sendPubKey := !bytes.Equal(initMsg.GetReceiverPeerId(), []byte(h.localPeerID))
		if sendPubKey {
			b, err := crypto.MarshalPublicKey(h.privKey.GetPublic())
			if err != nil {
				return nil, errors.Wrap(err, "marshal local public key")
			}

			ackCipher.SenderPubKey = b
		}

		// Compute the shared secret key
		copy(remoteEphPub[:], initMsg.GetSenderEphPub())
		box.Precompute(&sharedKey, &remoteEphPub, h.hPriv)

		// Compute the proof
		concatKeys := make([]byte, len(remoteEphPub)+len(*h.hPub))
		copy(concatKeys, remoteEphPub[:])
		copy(concatKeys[len(remoteEphPub):], (*h.hPub)[:])

		var err error
		ackCipher.TupleSignature, err = h.privKey.Sign(concatKeys)
		if err != nil {
			return nil, errors.Wrap(err, "sign ephemeral keypair")
		}

		// Encrypt the ciphertext.
		cipherData, err := proto.Marshal(ackCipher)
		if err != nil {
			return nil, err
		}

		ackPkt := &Packet_InitAck{}
		ackPkt.Ciphertext = box.SealAfterPrecomputation(nil, cipherData, &nonce, &sharedKey)
		ackPkt.SenderEphPub = (*h.hPub)[:]
		if err := h.writeInitAck(ackPkt); err != nil {
			return nil, err
		}

		// Await the incoming complete
		inPkt.Reset()
		if err := h.readPacket(ctx, inPkt, PacketType_PacketType_COMPLETE); err != nil {
			return nil, err
		}
		nextNonce()

		// Decrypt it
		completeCipherData, ok := box.OpenAfterPrecomputation(
			nil,
			inPkt.GetCompletePkt().GetCiphertext(),
			&nonce,
			&sharedKey,
		)
		if !ok {
			return nil, errors.New("handshake complete packet failed to decrypt")
		}

		ackCipher.Reset()
		if err := proto.Unmarshal(completeCipherData, ackCipher); err != nil {
			return nil, errors.Wrap(err, "handshake complete cipher failed to unmarshal")
		}

		// Decode the public key from the cipher
		if err := h.decodePubKey(ackCipher); err != nil {
			return nil, err
		}
		rxExtraData = ackCipher.GetExtraInfo()

		ok, err = h.remotePubKey.Verify(concatKeys, ackCipher.GetTupleSignature())
		if err != nil {
			return nil, errors.Wrap(err, "validate remote tuple signature")
		}
		if !ok {
			return nil, errors.New("remote tuple signature invalid")
		}
	}

	res := &identity.Result{Peer: h.remotePubKey, ExtraData: rxExtraData}
	res.Secret = sharedKey
	return res, nil
}

// writeInit writes the initial packet.
func (h *Handshaker) writeInit() error {
	return h.writePacket(&Packet{
		PacketType: PacketType_PacketType_INIT,
		InitPkt: &Packet_Init{
			SenderPeerId:   []byte(h.localPeerID),
			ReceiverPeerId: []byte(h.remotePeerID),
			SenderEphPub:   (*h.hPub)[:],
		},
	})
}

// writeInitAck writes the init packet acknowledgment packet.
func (h *Handshaker) writeInitAck(initAck *Packet_InitAck) error {
	return h.writePacket(&Packet{
		PacketType: PacketType_PacketType_INIT_ACK,
		InitAckPkt: initAck,
	})
}

// writeComplete writes the complete packet.
func (h *Handshaker) writeComplete(initAck *Packet_Complete) error {
	return h.writePacket(&Packet{
		PacketType:  PacketType_PacketType_COMPLETE,
		CompletePkt: initAck,
	})
}

// writePacket tries to write a packet.
func (h *Handshaker) writePacket(packet *Packet) error {
	dat, err := proto.Marshal(packet)
	if err != nil {
		return err
	}

	return h.writePacketFn(dat)
}

// readPacket attempts to read a packet.
func (h *Handshaker) readPacket(ctx context.Context, packet *Packet, expectedType PacketType) error {
	var pktDat []byte
	select {
	case <-ctx.Done():
		return ctx.Err()
	case pktDat = <-h.packetCh:
	}

	if err := proto.Unmarshal(pktDat, packet); err != nil {
		return errors.Errorf("unmarshal packet %v %v: %v", pktDat, expectedType.String(), err.Error())
	}

	if expectedType != PacketType(0) {
		if err := checkPacketType(packet.GetPacketType(), expectedType); err != nil {
			return err
		}
	}

	return nil
}

// decodePubKey decodes the remote public key from the ciphertext.
func (h *Handshaker) decodePubKey(cipher *Packet_Ciphertext) error {
	if pubKeyDat := cipher.GetSenderPubKey(); len(pubKeyDat) != 0 {
		var err error
		h.remotePubKey, err = crypto.UnmarshalPublicKey(pubKeyDat)
		if err != nil {
			return errors.Wrap(err, "unmarshal public key")
		}

		h.remotePeerID, err = peer.IDFromPublicKey(h.remotePubKey)
		if err != nil {
			return errors.Wrap(err, "generate remote peer id")
		}
	}

	if h.remotePubKey == nil {
		return errors.New("empty remote public key")
	}

	return nil
}

// Handle handles an incoming packet.
// The buffer is re-used upon return.
// Returns if another packet is expected.
func (h *Handshaker) Handle(data []byte) bool {
	// copy the buf
	b2 := make([]byte, len(data))
	copy(b2, data)

	select {
	case h.packetCh <- b2:
	default:
		return false
	}

	nPkt := atomic.AddInt32(&h.nInPkt, 1)

	if h.initiator {
		return nPkt < 1 // 1 packet expected
	}

	return nPkt < 2
}

// Close cleans up any resources allocated by the handshake.
func (h *Handshaker) Close() {
	h.hPriv = nil
	h.hPub = nil
	h.privKey = nil
	h.writePacketFn = nil
}
