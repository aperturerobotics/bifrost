package floodsub

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// subscription implements the pubsub subscription handle.
type subscription struct {
	// ctx is the context for publishing
	ctx context.Context
	// relOnce guards release
	relOnce sync.Once
	// m is the floodsub instance
	m *FloodSub
	// channelID is the channel identifier
	channelID string

	// mtx guards handlers
	mtx sync.Mutex
	// handlers are the active handlers
	handlers map[*subscriptionHandler]struct{}
}

// GetChannelId returns the channel id.
func (s *subscription) GetChannelId() string {
	return s.channelID
}

// Publish writes to the channel with a private key.
func (s *subscription) Publish(privKey crypto.PrivKey, data []byte) error {
	return s.m.Publish(s.ctx, s.channelID, privKey, data)
}

// AddHandler adds a callback that is called with each received message.
// The callback should not block.
// Returns a remove function.
// The handler(s) are also removed when the subscription is released.
func (b *subscription) AddHandler(cb func(m pubsub.Message)) func() {
	sh := &subscriptionHandler{cb: cb}
	b.mtx.Lock()
	b.handlers[sh] = struct{}{}
	b.mtx.Unlock()
	relOnce := sync.Once{}
	return func() {
		relOnce.Do(func() {
			b.mtx.Lock()
			delete(b.handlers, sh)
			b.mtx.Unlock()
		})
	}
}

// Release releases the handle.
func (s *subscription) Release() {
	s.mtx.Lock()
	for h := range s.handlers {
		delete(s.handlers, h)
	}
	s.mtx.Unlock()
	s.relOnce.Do(func() {
		chid := s.channelID
		s.m.mtx.Lock()
		subs := s.m.channels[chid]
		if subs != nil {
			delete(subs, s)
		}
		if len(subs) == 0 {
			defer s.m.wake()
		}
		s.m.mtx.Unlock()
	})
}

// _ is a type assertion
var _ pubsub.Subscription = ((*subscription)(nil))
