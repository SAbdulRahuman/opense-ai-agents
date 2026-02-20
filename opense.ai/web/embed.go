// Package web embeds the Next.js static export for serving from the Go binary.
//
// The web/out/ directory is produced by "npm run build" (next build with
// output: "export") and is embedded at compile-time using go:embed.
//
// Usage in the API server:
//
//	import "github.com/seenimoa/openseai/web"
//	fs := web.DistFS()  // returns io/fs.FS rooted at out/
package web

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed all:out
var dist embed.FS

// DistFS returns a filesystem rooted at the embedded out/ directory.
// This is ready to use with http.FileServerFS or http.FS.
func DistFS() fs.FS {
	sub, err := fs.Sub(dist, "out")
	if err != nil {
		log.Fatalf("web.DistFS: %v", err)
	}
	return sub
}
