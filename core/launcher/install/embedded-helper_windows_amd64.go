package launcher_install

import "embed"

// helperFS holds the helper the build placed for this architecture.
//
//go:embed all:helper/windows-amd64
var helperFS embed.FS

// helperDir is the helperFS directory holding this architecture's helper.
const helperDir = "helper/windows-amd64"
