// Package dashboard embeds the live-demo web UI served by the proxy.
package dashboard

import "embed"

//go:embed index.html
var FS embed.FS

// IndexHTML is the entry file of the embedded dashboard.
const IndexHTML = "index.html"
