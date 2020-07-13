package nats

import (
	nats_server "github.com/nats-io/nats-server/v2/server"
)

// Validate validates the configuration.
func (c *Config) Validate() error { return nil }

// ApplyOptions applies the nats server options.
func (c *Config) ApplyOptions(opts *nats_server.Options) error {
	if c.GetTrace() {
		opts.Trace = true
		opts.TraceVerbose = true
	}

	// TODO implement remaining options
	/*
		NoSigs                bool          `json:"-"`
		NoSublistCache        bool          `json:"-"`
		NoHeaderSupport       bool          `json:"-"`
		DisableShortFirstPing bool          `json:"-"`
		MaxConn               int           `json:"max_connections"`
		MaxSubs               int           `json:"max_subscriptions,omitempty"`
		Nkeys                 []*NkeyUser   `json:"-"`
		Users                 []*User       `json:"-"`
		Accounts              []*Account    `json:"-"`
		NoAuthUser            string        `json:"-"`
		SystemAccount         string        `json:"-"`
		NoSystemAccount       bool          `json:"-"`
		AllowNewAccounts      bool          `json:"-"`
		Username              string        `json:"-"`
		Password              string        `json:"-"`
		Authorization         string        `json:"-"`
		PingInterval          time.Duration `json:"ping_interval"`
		MaxPingsOut           int           `json:"ping_max"`
		AuthTimeout           float64       `json:"auth_timeout"`
		MaxControlLine        int32         `json:"max_control_line"`
		MaxPayload            int32         `json:"max_payload"`
		MaxPending            int64         `json:"max_pending"`
		Cluster               ClusterOpts   `json:"cluster,omitempty"`
		Gateway               GatewayOpts   `json:"gateway,omitempty"`
		LeafNode              LeafNodeOpts  `json:"leaf,omitempty"`
		JetStream             bool          `json:"jetstream"`
		JetStreamMaxMemory    int64         `json:"-"`
		JetStreamMaxStore     int64         `json:"-"`
		StoreDir              string        `json:"-"`
		Websocket             WebsocketOpts `json:"-"`
		// RoutePeers are route peer IDs to connect to
		RoutePeers          []string      `json:"-"`
		WriteDeadline       time.Duration `json:"-"`
		MaxClosedClients    int           `json:"-"`
		LameDuckDuration    time.Duration `json:"-"`
		LameDuckGracePeriod time.Duration `json:"-"`

		// MaxTracedMsgLen is the maximum printable length for traced messages.
		MaxTracedMsgLen int `json:"-"`

		// Operating a trusted NATS server
		TrustedKeys      []string              `json:"-"`
		TrustedOperators []*jwt.OperatorClaims `json:"-"`
		AccountResolver  AccountResolver       `json:"-"`
		resolverPreloads map[string]string

		CustomClientAuthentication Authentication `json:"-"`
		CustomRouterAuthentication Authentication `json:"-"`

		// CheckConfig configuration file syntax test was successful and exit.
		CheckConfig bool `json:"-"`

		// ConnectErrorReports specifies the number of failed attempts
		// at which point server should report the failure of an initial
		// connection to a route, gateway or leaf node.
		// See DEFAULT_CONNECT_ERROR_REPORTS for default value.
		ConnectErrorReports int

		// ReconnectErrorReports is similar to ConnectErrorReports except
		// that this applies to reconnect events.
		ReconnectErrorReports int
	*/
	return nil
}
