package web

import "embed"

// FS contains the complete offline dashboard.
//
//go:embed static
var FS embed.FS
