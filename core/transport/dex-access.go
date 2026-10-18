package transport

import dex_solicit "github.com/s4wave/spacewave/db/dex/solicit"

// WithDEXRequestAuthorizer checks each incoming block request against the
// creator's current authority. It leaves outbound reads and public transports
// without this option unchanged.
func WithDEXRequestAuthorizer(authorize dex_solicit.RequestAuthorizer) SessionTransportOption {
	return func(t *SessionTransport) {
		t.authorizeDEXRequest = authorize
	}
}
