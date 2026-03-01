package static

import "embed"

//go:embed *.js *.ico
var Files embed.FS
