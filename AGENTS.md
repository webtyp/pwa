# AGENTS.md — webtyp/pwa

Working notes for AI agents operating in this library. End-user docs: [README.md](README.md).

## Mission

`pwa` makes a webtyp application installable and able to start without network. It has two
halves that share one set of names (`protocol.go`):

- **Build time** (root package, `//go:build !wasm`): from the list of shell assets with their
  content hash, it produces the bytes of `manifest.webmanifest` and `sw.js`, the `<head>` tags and
  the registration script. `webtyp/sitec` calls it when a project declares `PWA() pwa.Config`.
  It never reads or writes files and knows nothing about how the assets were built.
- **Run time** (`update/`, `//go:build wasm`): the page learns that a new version is ready
  (`update.OnReady`) and lets it take over when the user agrees (`update.Apply`).

It does not handle large artifacts (model weights): that is `webtyp/artifacts`, OPFS, at run time.

## Two build regimes in one repo — do not mix them

| Files | Build tag | Imports allowed |
|---|---|---|
| `pwa.go`, `manifest.go`, `sw.go` (root) | `//go:build !wasm` | the Go standard library (`encoding/json`, `crypto/sha256`, `strings`, `sort`, `errors`, `fmt`): this code runs inside the build tool, never in a browser. **Do not "fix" these imports.** |
| `protocol.go` (root) | none | constants only — no imports |
| `update/*.go` | `//go:build wasm` | browser rules below |

Browser rules for `update/` (compiled by TinyGo; the build that decides is `gotest -tinygo`):

| Do not import | Use instead |
|---|---|
| `fmt`, `errors`, `strconv`, `strings` | `webtyp.com/fmt` |
| `encoding/json` (~1 MB of wasm under TinyGo) | `webtyp.com/json` |
| `net/http` | `webtyp.com/fetch` |
| `context`, `time` | `webtyp.com/context`, `webtyp.com/time` |
| `map[K]V` | a slice of structs, or `fmt.KeyValue` |

`GOOS=js GOARCH=wasm go build` succeeding proves nothing: it uses the full standard library and does
not imply TinyGo. Waiting on a JavaScript promise: `webtyp.com/await`.

## Tests

All tests in `tests/` (`package tests`, public API only). Build-side tests carry `//go:build !wasm`,
browser tests `//go:build wasm`.

```bash
go install webtyp.com/devflow/cmd/gotest@latest   # once; needs gotest >= v0.4.108
gotest
gotest -tinygo
```

Never run `gopush` or `codejob`.
