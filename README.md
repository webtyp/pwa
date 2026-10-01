# pwa

Build-time generator of `manifest.webmanifest` and `sw.js` for webtyp applications, plus the page-side update signal (`pwa/update`).

## Overview

`pwa` makes a webtyp application installable and able to start without network. It has two halves that share one set of protocol constants (`protocol.go`):

1. **Build time** (root package `webtyp.com/pwa`, `//go:build !wasm`): Converts a list of shell assets with content hashes into `manifest.webmanifest`, `sw.js`, `<head>` tags, and the registration script. `webtyp/sitec` calls `pwa.Generate` when a project declares `PWA() pwa.Config`.
2. **Run time** (`webtyp.com/pwa/update`, `//go:build wasm`): Runs in the browser page to detect when a new application version is ready (`update.OnReady`) and trigger update activation when agreed (`update.Apply`).

## API Summary

| I want to… | Use |
|---|---|
| Generate PWA assets at build time | `pwa.Generate(cfg, icons, shell)` |
| Know when a new version is waiting | `update.OnReady(fn)` |
| Switch to the waiting version | `update.Apply()` |

> **Note:** `webtyp/sitec` calls `pwa.Generate` when a project declares `PWA() pwa.Config`. A project never calls `pwa.Generate` by hand.

## Examples

### 1. Build Time: `pwa.Generate` (called by `sitec`)

```go
package main

import (
	"fmt"
	"log"

	"webtyp.com/pwa"
)

func main() {
	cfg := pwa.Config{
		Name:            "Clínica Monjitas",
		ShortName:       "Monjitas",
		ThemeColor:      "#0055ff",
		BackgroundColor: "#ffffff",
	}

	icons := []pwa.Icon{
		{URL: "/icon-192.png", Sizes: "192x192", Type: "image/png"},
		{URL: "/icon-512.png", Sizes: "512x512", Type: "image/png"},
	}

	shell := []pwa.Asset{
		{URL: "/", Revision: "a1b2c3d4"},
		{URL: "/style.css", Revision: "e5f6g7h8"},
		{URL: "/client.wasm", Revision: "i9j0k1l2"},
	}

	build, err := pwa.Generate(cfg, icons, shell)
	if err != nil {
		log.Fatalf("pwa.Generate failed: %v", err)
	}

	fmt.Printf("Generated version: %s\n", build.Version)
	// Output build.Manifest, build.ServiceWorker, build.HeadTags, build.RegisterScript...
}
```

### 2. Run Time: `update.OnReady` and `update.Apply` (in Wasm page code)

```go
//go:build wasm

package main

import (
	"log"

	"webtyp.com/pwa/update"
)

func main() {
	update.OnReady(func() {
		log.Println("A new version is ready!")
		// Notify user / show update prompt UI
		if err := update.Apply(); err != nil {
			log.Printf("Failed to apply update: %v", err)
		}
	})
}
```
