package flowgraph_run

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
)

// Validate requires the node and a well-formed ordered list of events.
func (r *DeliverEventsRequest) Validate() error {
	if r.GetNodeId() == "" {
		return errors.New("deliver events: node ID is empty")
	}
	for i, event := range r.GetEvents() {
		if event.GetObjectKey() == "" {
			return errors.Errorf("deliver events: event %d has no object key", i)
		}
		if event.GetEventId() == "" {
			return errors.Errorf("deliver events: event %d has no event ID", i)
		}
	}
	return nil
}

// DeliverEvents queues one arrival per event along each connection from an
// event-source node into a Step input, and stores the cursor after the last
// event. It applies only when the run's stored cursor for the node equals the
// cursor the request read; otherwise it changes nothing and returns
// DELIVER_EVENTS_RESULT_STALE_CURSOR, and the caller resumes from the stored
// cursor. The arrivals and the cursor are written together, so an event is
// never queued without its cursor or the cursor moved without its event.
// Call it in a write transaction; the run controller for runKey routes the
// arrivals.
func DeliverEvents(
	ctx context.Context,
	ws world.WorldState,
	runKey string,
	req *DeliverEventsRequest,
) (DeliverEventsResult, error) {
	// Read the run and its pinned graph.
	if err := req.Validate(); err != nil {
		return DeliverEventsResult_DELIVER_EVENTS_RESULT_UNKNOWN, err
	}
	obj, run, graph, err := readRun(ctx, ws, runKey)
	defer world.ReleaseObjectState(obj)
	if err != nil {
		return DeliverEventsResult_DELIVER_EVENTS_RESULT_UNKNOWN, err
	}

	// Require a live run and a node of its pinned graph.
	if run.GetState() == s4wave_flowgraph.FlowgraphRunState_FLOWGRAPH_RUN_STATE_COMPLETE {
		return DeliverEventsResult_DELIVER_EVENTS_RESULT_UNKNOWN, errors.New("cannot deliver events to a completed run")
	}
	if _, ok := graph.GetNodes()[req.GetNodeId()]; !ok {
		return DeliverEventsResult_DELIVER_EVENTS_RESULT_UNKNOWN, errors.Errorf("deliver events: node %q is not in the run's graph", req.GetNodeId())
	}

	// Reject a delivery that read a cursor another reader has moved.
	if !bytes.Equal(run.GetCursor()[req.GetNodeId()], req.GetCursor()) {
		return DeliverEventsResult_DELIVER_EVENTS_RESULT_STALE_CURSOR, nil
	}

	// Queue each event along the node's connections and move the cursor.
	next := run.CloneVT()
	connections := outgoing(graph, req.GetNodeId())
	for _, event := range req.GetEvents() {
		for _, connection := range connections {
			next.Arrivals = append(next.Arrivals, &s4wave_flowgraph.FlowgraphArrival{
				NodeId: connection.GetInputNode(),
				Port:   connection.GetInputPort(),
				Key:    event.GetKey(),
				Source: &s4wave_flowgraph.FlowgraphArrival_Event{
					Event: &s4wave_flowgraph.FlowgraphEvent{
						ObjectKey: event.GetObjectKey(),
						EventId:   event.GetEventId(),
					},
				},
			})
		}
	}
	if len(req.GetNextCursor()) == 0 {
		delete(next.Cursor, req.GetNodeId())
	} else {
		if next.Cursor == nil {
			next.Cursor = make(map[string][]byte)
		}
		next.Cursor[req.GetNodeId()] = bytes.Clone(req.GetNextCursor())
	}

	// Write the arrivals and the cursor in this one transaction.
	_, _, err = world.AccessObjectState(ctx, obj, true, func(cursor *block.Cursor) error {
		cursor.SetBlock(next, true)
		return nil
	})
	if err != nil {
		return DeliverEventsResult_DELIVER_EVENTS_RESULT_UNKNOWN, err
	}
	return DeliverEventsResult_DELIVER_EVENTS_RESULT_DELIVERED, nil
}
