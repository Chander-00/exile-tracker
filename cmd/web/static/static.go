package static

import "embed"

//go:embed *.js *.ico *.json
var Files embed.FS
