package resource_session

import (
	"context"

	s4wave_session "github.com/s4wave/spacewave/sdk/session"
)

// ConfirmSASMatch sends the user's SAS emoji verification decision to the
// bilateral confirmation exchange running over the bifrost link.
func (r *SessionResource) ConfirmSASMatch(ctx context.Context, req *s4wave_session.ConfirmSASMatchRequest) (*s4wave_session.ConfirmSASMatchResponse, error) {
	engine, err := r.getPairingEngine()
	if err != nil {
		return nil, err
	}

	engine.ConfirmSAS(req.GetConfirmed())
	return &s4wave_session.ConfirmSASMatchResponse{}, nil
}
