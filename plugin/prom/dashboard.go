package prom_plugin

import (
	"context"

	"github.com/s4wave/spacewave/db/world"
	s4wave_prom "github.com/s4wave/spacewave/sdk/prom"
)

// dashboardService serves one Dashboard's edits and changes.
type dashboardService struct {
	// eng is the Space's World engine.
	eng world.Engine
	// objectKey is the Dashboard's object key.
	objectKey string
}

// WatchDashboard streams the Dashboard now and after each commit that changes
// it. A Dashboard is small, so each response carries all of it.
func (s *dashboardService) WatchDashboard(_ *s4wave_prom.WatchDashboardRequest, strm s4wave_prom.SRPCPromDashboardResourceService_WatchDashboardStream) error {
	var sent *s4wave_prom.Dashboard
	return watchWorld(strm.Context(), s.eng, 0, func(ctx context.Context, ws world.WorldState) error {
		dash, err := s4wave_prom.ReadDashboard(ctx, ws, s.objectKey)
		if err != nil || (sent != nil && sent.EqualVT(dash)) {
			return err
		}
		sent = dash
		return strm.Send(dash)
	})
}

// PutPanel adds or replaces a panel.
func (s *dashboardService) PutPanel(ctx context.Context, req *s4wave_prom.PutPanelRequest) (*s4wave_prom.PutPanelResponse, error) {
	var id string
	err := s.update(ctx, func(dash *s4wave_prom.Dashboard) error {
		var err error
		id, err = dash.PutPanel(req.GetPanel())
		return err
	})
	if err != nil {
		return nil, err
	}
	return &s4wave_prom.PutPanelResponse{Id: id}, nil
}

// RemovePanel removes a panel.
func (s *dashboardService) RemovePanel(ctx context.Context, req *s4wave_prom.RemovePanelRequest) (*s4wave_prom.RemovePanelResponse, error) {
	err := s.update(ctx, func(dash *s4wave_prom.Dashboard) error {
		dash.RemovePanel(req.GetId())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &s4wave_prom.RemovePanelResponse{}, nil
}

// MovePanel moves a panel to a new position.
func (s *dashboardService) MovePanel(ctx context.Context, req *s4wave_prom.MovePanelRequest) (*s4wave_prom.MovePanelResponse, error) {
	err := s.update(ctx, func(dash *s4wave_prom.Dashboard) error {
		return dash.MovePanel(req.GetId(), int(req.GetIndex()))
	})
	if err != nil {
		return nil, err
	}
	return &s4wave_prom.MovePanelResponse{}, nil
}

// CreateTarget creates a Scrape target in the Dashboard's Space.
func (s *dashboardService) CreateTarget(ctx context.Context, req *s4wave_prom.CreateTargetRequest) (*s4wave_prom.CreateTargetResponse, error) {
	err := world.ExecTransaction(ctx, s.eng, true, func(ctx context.Context, wtx world.WorldState) error {
		return s4wave_prom.CreateTarget(ctx, wtx, req.GetObjectKey(), req.GetMeta())
	})
	if err != nil {
		return nil, err
	}
	return &s4wave_prom.CreateTargetResponse{}, nil
}

// update applies edit to the Dashboard in one write transaction.
func (s *dashboardService) update(ctx context.Context, edit func(*s4wave_prom.Dashboard) error) error {
	return world.ExecTransaction(ctx, s.eng, true, func(ctx context.Context, wtx world.WorldState) error {
		return s4wave_prom.UpdateDashboard(ctx, wtx, s.objectKey, edit)
	})
}

// _ is a type assertion
var _ s4wave_prom.SRPCPromDashboardResourceServiceServer = (*dashboardService)(nil)
