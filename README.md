# pwa
<img src="docs/img/badges.svg">

Build-time generator of `manifest.webmanifest` and `sw.js` for webtyp applications, plus the page-side update signal (`pwa/update`).

## Overview

`pwa` makes a webtyp application installable and able to start without network. It has two halves that share one set of protocol constants (`protocol.go`):

1. **Build time** (root package `webtyp.com/pwa`, `//go:build !wasm`): Converts a list of shell assets with content hashes into `manifest.webmanifest`, `sw.js`, `<head>` tags, and the registration script. `webtyp/sitec` calls `pwa.New` and `App.ServiceWorker` when a project declares `PWA() pwa.Config`.
2. **Run time** (`webtyp.com/pwa/update`, `//go:build wasm`): Runs in the browser page to detect when a new application version is ready (`update.OnReady`) and trigger update activation when agreed (`update.Apply`).

## API Summary

| I want to… | Use |
|---|---|
| Manifest, `<head>` tags and register script | `app, err := pwa.New(cfg, icons)` |
| Service worker for the hashed shell | `worker, err := app.ServiceWorker(shell)` |
| Know when a new version is waiting | `update.OnReady(fn)` |
| Switch to the waiting version | `update.Apply()` |

> **Note:** `webtyp/sitec` calls both steps when a project declares `PWA() pwa.Config`; a project
> never calls them by hand. The order matters and the types enforce it: insert `app.HeadTags` into
> the HTML and `app.RegisterScript` into the main script **before** hashing them, because the
> service worker precaches the final files.

## Examples

### 1. Build Time: `pwa.New` and `App.ServiceWorker` (called by `sitec`)

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

	app, err := pwa.New(cfg, icons)
	if err != nil {
		log.Fatal(err)
	}
	// Insert app.HeadTags into every HTML page and append app.RegisterScript to the main
	// script, then hash the final files:
	shell := []pwa.Asset{
		{URL: "/", Revision: "a1b2c3d4"},
		{URL: "/style.3f9a1c2b.css", Revision: "3f9a1c2b"},
		{URL: "/client.9a8b7c6d.wasm", Revision: "9a8b7c6d"},
	}
	worker, err := app.ServiceWorker(shell)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("version %s\n", worker.Version)
	// Serve app.Manifest at pwa.ManifestPath and worker.Script at pwa.ServiceWorkerPath.
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
