// Package web embeds the Next.js static export so the Go binary can serve
// the dashboard without a separate runtime.
package web

import "embed"

//go:embed all:out
var Assets embed.FS
