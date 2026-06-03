package web

import "embed"

// FS holds embedded web assets.
//
//go:embed signin_fallback.html
var FS embed.FS
