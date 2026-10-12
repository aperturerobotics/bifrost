package plugin_space

import (
	"context"

	space_exec "github.com/s4wave/spacewave/core/forge/exec"
)

// execRecorder records the Executions a plugin runs and gates their completion.
type execRecorder struct {
	// requests delivers each request the plugin receives.
	requests chan *space_exec.PluginExecRequest
	// finish releases the plugin execution after the test checks its running state.
	finish chan struct{}
}

// Execute records the request and waits for the test to release its completion.
func (r *execRecorder) Execute(ctx context.Context, req *space_exec.PluginExecRequest) (*space_exec.PluginExecResponse, error) {
	// Report the Execution that reached the plugin service.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r.requests <- req.CloneVT():
	}

	// Keep the Execution running until its cleared wait has been checked.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.finish:
		return &space_exec.PluginExecResponse{}, nil
	}
}

// ExecuteStream sends the unary result over the streaming interface.
func (r *execRecorder) ExecuteStream(req *space_exec.PluginExecRequest, stream space_exec.SRPCPluginExecService_ExecuteStreamStream) error {
	// Wait for the test to release the Execution's response.
	resp, err := r.Execute(stream.Context(), req)
	if err != nil {
		return err
	}

	// Send the plugin's terminal response.
	return stream.Send(resp)
}

// _ verifies the plugin service contract.
var _ space_exec.SRPCPluginExecServiceServer = (*execRecorder)(nil)
