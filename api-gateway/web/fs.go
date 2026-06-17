// Package web expõe o frontend HTML embutido no binário via go:embed.
package web

import "embed"

//go:embed index.html
var FS embed.FS
