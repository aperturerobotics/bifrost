package objecttype

import (
	"context"

	"github.com/s4wave/spacewave/core/sobject"
)

// callerKey holds the authenticated identity carried by typed dispatch.
type callerKey struct{}

// Caller identifies the connection that requested a typed Resource. An empty
// PeerID denotes a trusted local connection, not an anonymous remote caller.
type Caller struct {
	// PeerID is the authenticated remote Session peer.
	PeerID string
	// Participant is the accepted participant at dispatch, when known.
	// Consumers resolve current permissions from the addressed SharedObject.
	Participant *sobject.SOParticipantConfig
}

// WithCaller binds the authenticated identity selected by a transport or mount.
// Callers of this function must already have authenticated the peer.
func WithCaller(ctx context.Context, caller Caller) context.Context {
	caller.Participant = caller.Participant.CloneVT()
	return context.WithValue(ctx, callerKey{}, caller)
}

// CallerFromContext returns the authenticated caller of typed dispatch.
// It also recognizes the Session peer of a native factory context.
func CallerFromContext(ctx context.Context) Caller {
	if caller, ok := ctx.Value(callerKey{}).(Caller); ok {
		caller.Participant = caller.Participant.CloneVT()
		return caller
	}
	return Caller{PeerID: SessionPeerIDFromContext(ctx).String()}
}
