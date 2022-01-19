package drpc_e2e

import (
	context "context"

	"github.com/aperturerobotics/bifrost/protocol"
	"storj.io/drpc"
)

// ProtocolID is the protocol ID for the end-to-end test.
const ProtocolID protocol.ID = "bifrost/stream/drpc/e2e"

// Server is the e2e server.
type Server struct {
}

// NewServer constructs the server.
func NewServer() *Server {
	return &Server{}
}

// Register registers the API with a DRPC mux.
func (s *Server) Register(mux drpc.Mux) error {
	return DRPCRegisterEndToEnd(mux, s)
}

// Mock performs the mock request.
func (s *Server) Mock(ctx context.Context, req *MockRequest) (*MockResponse, error) {
	return &MockResponse{
		ReqBody: req.GetBody(),
	}, nil
}

// _ is a type assertion
var _ DRPCEndToEndServer = ((*Server)(nil))
