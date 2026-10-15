package plugin_host_wazero_quickjs

import _ "embed"

// pluginQuickjsBoot is the boot harness that starts a quickjs plugin.
//
//go:generate go run -v ./gen/main.go
//go:embed plugin-quickjs.esm.js
var pluginQuickjsBoot string
