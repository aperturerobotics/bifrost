package nats

import (
	"context"
	"io"
	"sync"

	hash "github.com/aperturerobotics/bifrost/hash"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/bifrost/pubsub/util/pubmessage"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// subscription implements the pubsub subscription handle.
type subscription struct {
	// ctx is the context for the subscription + publishing
	ctx context.Context
	// ctxCancel is the context cancel handle
	ctxCancel context.CancelFunc
	// ncRel releases the nats client reference
	ncRel func()
	// relOnce guards release
	relOnce sync.Once
	// m is the nats instance
	m *Nats
	// client is the nats client
	client *natsClient
	// channelID is the channel identifier
	channelID string
	// sub is the nats subscription handle
	sub *nats.Subscription
	// privKey is the private key used for the sub
	privKey crypto.PrivKey

	// mtx guards below
	mtx sync.Mutex
	// handlers are the active handlers
	handlers map[*subscriptionHandler]struct{}
}

// newSubscription constructs a new subscription handle.
func newSubscription(
	ctx context.Context,
	m *Nats, c *natsClient, ncRel func(),
	sub *nats.Subscription,
	privKey crypto.PrivKey,
	channelID string,
) *subscription {
	ssub := &subscription{
		client: c,
		m:      m,
		sub:    sub,

		channelID: channelID,
		ncRel:     ncRel,
		privKey:   privKey,
	}
	ssub.ctx, ssub.ctxCancel = context.WithCancel(ctx)
	go ssub.executeSubscription()
	return ssub
}

// GetPeerId returns the peer ID for this subscription derived from private key.
func (s *subscription) GetPeerId() peer.ID {
	return s.client.npeerID
}

// GetChannelId returns the channel id.
func (s *subscription) GetChannelId() string {
	return s.channelID
}

// Publish writes to the channel.
func (s *subscription) Publish(data []byte) error {
	pht := s.m.conf.GetPublishHashType()
	if pht == hash.HashType_HashType_UNKNOWN {
		pht = hash.HashType_HashType_SHA256
	}

	channelID := s.channelID
	privKey := s.privKey
	msg, _, err := pubmessage.NewPubMessage(channelID, privKey, pht, data)
	if err != nil {
		return err
	}

	enc, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return s.PublishRaw(enc)
}

// PublishRaw publishes pre-encoded data (should be a PubMessage).
func (s *subscription) PublishRaw(data []byte) error {
	return s.client.Publish(s.channelID, data)
}

// AddHandler adds a callback that is called with each received message.
// The callback should not block.
// Returns a remove function.
// The handler(s) are also removed when the subscription is released.
func (s *subscription) AddHandler(cb func(m pubsub.Message)) func() {
	sh := &subscriptionHandler{cb: cb}
	s.mtx.Lock()
	if s.handlers == nil {
		s.handlers = make(map[*subscriptionHandler]struct{})
	}
	s.handlers[sh] = struct{}{}
	s.mtx.Unlock()
	relOnce := sync.Once{}
	return func() {
		relOnce.Do(func() {
			s.mtx.Lock()
			if s.handlers != nil {
				delete(s.handlers, sh)
			}
			s.mtx.Unlock()
		})
	}
}

// Release releases the handle.
func (s *subscription) Release() {
	s.relOnce.Do(func() {
		_ = s.sub.Unsubscribe()
		if s.ncRel != nil {
			s.ncRel()
		}
		s.ctxCancel()
	})
	s.mtx.Lock()
	if s.handlers != nil {
		for h := range s.handlers {
			delete(s.handlers, h)
		}
		s.handlers = nil
	}
	s.mtx.Unlock()
}

// logger attaches additional logging fields.
func (s *subscription) logger() *logrus.Entry {
	npp := s.client.npeerIDPretty
	return s.m.le.
		WithField("channel-id", s.channelID).
		WithField("peer-id", npp)
}

// executeSubscription manages the subscription.
func (s *subscription) executeSubscription() {
	ctx := s.ctx
	defer s.Release()

	for {
		msg, err := s.sub.NextMsgWithContext(ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}
			s.logger().WithError(err).Warn("error processing subscription messages: exiting")
			return
		}
		if err := s.handleNatsMessage(msg); err != nil {
			if err != io.EOF && err != context.Canceled {
				s.logger().WithError(err).Warnf("rx invalid nats message: %s", string(msg.Data))
			}
		}
	}
}

// handleNatsMessage handles an incoming nats.Msg
func (s *subscription) handleNatsMessage(msg *nats.Msg) error {
	// decode nats.msg to SignedMsg.
	pm := &peer.SignedMsg{}
	pm, err := peer.UnmarshalSignedMsg(msg.Data)
	if err != nil {
		return errors.Wrap(err, "unmarshal nats message data")
	}
	inner, _, peerID, err := pubmessage.ExtractAndVerify(pm)
	if err != nil {
		return err
	}
	s.handlePubMessages([]*pubmessage.Message{
		pubmessage.NewMessage(peerID, inner),
	})
	return nil
}

// handlePubMessages handles an incoming nats.Msg
func (s *subscription) handlePubMessages(msgs []*pubmessage.Message) {
	s.mtx.Lock()
	if s.handlers != nil {
		for _, msg := range msgs {
			for handler := range s.handlers {
				handler.cb(msg)
			}
		}
	}
	s.mtx.Unlock()
}

// _ is a type assertion
var _ pubsub.Subscription = ((*subscription)(nil))
