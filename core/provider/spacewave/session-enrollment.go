package provider_spacewave

import (
	"context"

	"github.com/pkg/errors"
	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
)

// LinkSession registers the receiving client's independent key through the
// calling Session's authority. The provider remains the membership authority.
func (a *ProviderAccount) LinkSession(ctx context.Context, sourceKey crypto.PrivKey, receivingPeer peer.ID, label string) (*api.RegisterSessionResponse, error) {
	if sourceKey == nil {
		return nil, errors.New("authorizing Session is locked")
	}
	sourcePeer, err := peer.IDFromPrivateKey(sourceKey)
	if err != nil {
		return nil, err
	}
	if sourcePeer == receivingPeer {
		return nil, errors.New("the receiving client must use its own Session key")
	}
	client := NewSessionClient(a.p.httpCli, a.p.endpoint, a.p.GetSigningEnvPrefix(), sourceKey, sourcePeer.String())
	result, err := client.LinkSession(ctx, receivingPeer, label)
	if err != nil {
		return nil, err
	}
	if result.GetAccountId() != a.accountID {
		return nil, errors.New("Session enrollment returned another account")
	}
	return result, nil
}

// LinkSession authorizes one receiving USER Session without exporting a key.
func (c *SessionClient) LinkSession(ctx context.Context, receivingPeer peer.ID, label string) (*api.RegisterSessionResponse, error) {
	if _, err := receivingPeer.ExtractPublicKey(); err != nil {
		return nil, errors.Wrap(err, "invalid receiving Session key")
	}
	body, err := (&api.RegisterSessionRequest{SessionPeerId: receivingPeer.String(), Type: session.SessionType_SESSION_TYPE_USER, Label: label}).MarshalVT()
	if err != nil {
		return nil, err
	}
	data, err := c.doPost(ctx, "/api/account/session/link", "application/octet-stream", body, nil, SeedReasonMutation)
	if err != nil {
		return nil, err
	}
	result := &api.RegisterSessionResponse{}
	if err := result.UnmarshalVT(data); err != nil {
		return nil, err
	}
	if result.GetPeerId() != receivingPeer.String() || result.GetAccountId() == "" {
		return nil, errors.New("provider did not register the receiving Session")
	}
	return result, nil
}
