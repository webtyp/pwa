---
PLAN: "feat: New + ServiceWorker — manifest.webmanifest, sw.js, head tags and update signal for webtyp apps"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 11659821305359293204
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `webtyp/pwa` v0.1.0

Master plan: [PWA_ARTIFACTS_MASTER_PLAN.md](https://github.com/webtyp/app/blob/main/docs/PWA_ARTIFACTS_MASTER_PLAN.md),
decisions D-PWA-4, D-PWA-5, D-PWA-8, D-PWA-11. **Read [AGENTS.md](../AGENTS.md) first**: this repo
has two build regimes and the root package legitimately uses the standard library.

## Why

A webtyp application is HTML + CSS + JS + `client.wasm`. To install it and start it without network
it needs a `manifest.webmanifest` and a service worker that keeps one coherent version of the shell
in Cache Storage. The asset compiler (`webtyp/sitec`) will give this library the list of shell
assets with a content hash; this library returns bytes. It never touches disk. The repo holds a
skeleton (`go.mod`, `doc.go`, `README.md`); keep `doc.go`'s package comment.

## Design gate

1. **Prior art.**
   - **Workbox** (`workbox-build`, used by Vite PWA, `next-pwa`): the bundler gives it the emitted
     files; it writes a precache list (URL + revision) into the service worker; the new worker waits
     and `workbox-window` tells the page (`waiting`) so it can offer a reload. We take: precache list
     with revisions, wait-then-ask update, a separate library the bundler calls.
   - **Angular `@angular/service-worker`**: the build writes `ngsw.json` (every file + hash); the hash
     of that table is the app version; `SwUpdate` tells the app a version is ready and activates it
     on reload. We take: **version = hash of the list**, one coherent set per version.
   - **SvelteKit `$service-worker`**: exposes `build` (hashed files) and `version`, used as the cache
     name. We take: cache name = prefix + version, old caches deleted on activate.
   - What differs: no runtime caching strategies, no plugin system, no configuration of routes —
     one fixed strategy (precache + cache-first for exactly the listed URLs), because webtyp emits a
     known, small shell. Large artifacts are another library (OPFS, not Cache Storage).
2. **Novice-name test.** `app, err := pwa.New(cfg, icons)` → "a new PWA app from this config and
   these icons"; it carries `app.Manifest`, `app.HeadTags`, `app.RegisterScript`.
   `worker, err := app.ServiceWorker(shell)` → "the service worker for this shell"; it carries
   `worker.Script` and `worker.Version`. Two steps on purpose: the caller must insert the head tags
   and the register script into the HTML and JS **before** hashing them, and only the hashed shell
   gives the service worker. The types make that order the only one that compiles.
   `update.OnReady(fn)` → "on update ready, call fn"; `update.Apply()` → "apply the update".
3. **Complexity ledger.** New library: +6 types/functions at build time, +2 at run time. A project
   adds one method, `PWA() pwa.Config` (wired in sitec by a later plan). Ways to do the same thing:
   **−1** once `js.ServiceWorker` is deleted (another plan of the same wave, in `webtyp/js`).
4. **Where it belongs.** Its own repo (D-PWA-11): `sitec` compiles assets and must not learn what a
   service worker is; `artifacts` runs in the browser with another responsibility.
5. **What it deletes.** Nothing here (new capability); its existence is what lets `js.ServiceWorker`
   be deleted.

## Stage 1 — `protocol.go` (no build tag, no imports)

```go
package pwa

// Names shared by the files this package writes and the page code in pwa/update.
const (
	ManifestPath       = "/manifest.webmanifest" // where the manifest must be served
	ServiceWorkerPath  = "/sw.js"                // where the service worker must be served (scope "/")
	CachePrefix        = "webtyp-shell-"         // Cache Storage name = CachePrefix + Worker.Version
	EventUpdateReady   = "webtyp-update-ready"   // dispatched on window when a new version waits
	MessageSkipWaiting = "webtyp-skip-waiting"   // posted to the waiting worker to take over
)
```

## Stage 2 — build side (`//go:build !wasm`)

File `pwa.go`:

```go
// Config is what a project declares about itself to be installable.
type Config struct {
	Name            string // full name, shown on install ("Clínica Monjitas")
	ShortName       string // under the home-screen icon; empty = Name
	Description     string // optional
	ThemeColor      string // "#rgb" or "#rrggbb": browser toolbar
	BackgroundColor string // "#rgb" or "#rrggbb": splash screen
}

// Asset is one file of the application shell.
type Asset struct {
	URL      string // absolute path on the same origin: "/", "/style.3f9a1c2b.css"
	Revision string // hash of the content; changes when the content changes
}

// Icon is one install icon (from the project's favicon set).
type Icon struct {
	URL   string // "/icon-512.png"
	Sizes string // "512x512"
	Type  string // "image/png"
}

// App is what makes an application installable. Build it with New.
type App struct {
	Manifest       []byte // serve at ManifestPath
	HeadTags       string // insert in <head> of every HTML page
	RegisterScript string // append to the page's main script
}

// New validates the config and icons and returns the files that do not depend on the shell.
func New(c Config, icons []Icon) (App, error)

// Worker is the service worker of one exact set of shell assets.
type Worker struct {
	Script  []byte // serve at ServiceWorkerPath
	Version string // 16 hex chars; identifies this exact set of shell assets
}

// ServiceWorker returns the service worker that precaches shell. The shell must list every file
// the page needs to start, including "/" (the HTML page, after HeadTags were inserted); it must not
// list ServiceWorkerPath.
func (a App) ServiceWorker(shell []Asset) (Worker, error)
```

Validation, each error a named constant with this exact text (use `errors.New` / `fmt.Errorf` —
standard library is allowed here). `New` checks the first three rows in order; `ServiceWorker` the
rest in order. `ServiceWorker` is a method only so that it cannot be called without a validated
`App`; it does not need anything else from it:

| Condition | Error text |
|---|---|
| `c.Name == ""` | `pwa: Config.Name is empty` |
| a color not matching `#rgb` / `#rrggbb` (hex digits, either case) | `pwa: %s %q is not #rgb or #rrggbb` (field name `ThemeColor` / `BackgroundColor`) |
| no icon with `Sizes == "192x192"` or none with `"512x512"` | `pwa: an installable app needs icons of 192x192 and 512x512; missing %s` |
| `len(shell) == 0` | `pwa: the shell is empty` |
| an asset URL not starting with `/` | `pwa: asset URL %q is not an absolute path` |
| an asset with empty `Revision` | `pwa: asset %q has no revision` |
| an asset URL equal to `ServiceWorkerPath` | `pwa: the service worker cannot precache itself` |
| two assets with the same URL | `pwa: asset %q is listed twice` |
| no asset with URL `/` | `pwa: the shell does not include "/"` |

**Version** (`version.go`): sort a copy of `shell` by URL; SHA-256 over the concatenation of
`URL + "\t" + Revision + "\n"` for each; `Version` = first 16 lowercase hex characters. The same set
in any order gives the same version.

**Manifest** (`manifest.go`): JSON via `encoding/json` from an unexported struct, fields in this
order and with these values:
```json
{"name":"…","short_name":"…","description":"…","start_url":"/","scope":"/","display":"standalone",
 "background_color":"…","theme_color":"…","icons":[{"src":"…","sizes":"…","type":"…"}]}
```
`description` omitted when empty (`omitempty`); `short_name` = `Name` when `ShortName` is empty;
icons in the order given.

**HeadTags**: exactly
`<link rel="manifest" href="/manifest.webmanifest"><meta name="theme-color" content="` + ThemeColor + `">`
(built from the constants; the color is already validated, so no escaping is needed).

**Worker.Script** (`sw.go`): a `text/template` (or string building) that produces exactly this
program, where `{{VERSION}}` is the version and `{{SHELL}}` is the JSON array of the shell URLs in
sorted order (marshal a `[]string` with `encoding/json`). The constants come from `protocol.go`,
never retyped:

```js
const CACHE = "webtyp-shell-{{VERSION}}";
const SHELL = {{SHELL}};
self.addEventListener("install", (e) => {
  e.waitUntil(caches.open(CACHE).then((c) => c.addAll(SHELL.map((u) => new Request(u, { cache: "reload" })))));
});
self.addEventListener("activate", (e) => {
  e.waitUntil(caches.keys().then((keys) => Promise.all(
    keys.filter((k) => k.startsWith("webtyp-shell-") && k !== CACHE).map((k) => caches.delete(k)))));
});
self.addEventListener("message", (e) => {
  if (e.data && e.data.type === "webtyp-skip-waiting") self.skipWaiting();
});
self.addEventListener("fetch", (e) => {
  const req = e.request;
  if (req.method !== "GET") return;
  const url = new URL(req.url);
  if (url.origin !== self.location.origin || !SHELL.includes(url.pathname)) return;
  e.respondWith(caches.open(CACHE).then((c) => c.match(url.pathname)).then((r) => r || fetch(req)));
});
```

Deliberately absent, and the reason goes in a comment above the template: no `skipWaiting()` on
install and no `clients.claim()` — a new version waits until the page applies it (D-PWA-8), so a
page never runs HTML of one version with wasm of another. Requests not in `SHELL` (API calls,
`/__webtyp/ca`, large artifacts) are not touched.

**RegisterScript**: exactly this program (constants interpolated from `protocol.go`):

```js
if ("serviceWorker" in navigator) {
  const sw = navigator.serviceWorker;
  const hadController = !!sw.controller;
  let reloading = false;
  sw.addEventListener("controllerchange", () => {
    if (!hadController || reloading) return;
    reloading = true;
    location.reload();
  });
  sw.register("/sw.js").then((reg) => {
    const notify = () => window.dispatchEvent(new Event("webtyp-update-ready"));
    const watch = (w) => w && w.addEventListener("statechange", () => {
      if (w.state === "installed" && sw.controller) notify();
    });
    if (reg.waiting && sw.controller) notify();
    watch(reg.installing);
    reg.addEventListener("updatefound", () => watch(reg.installing));
  });
}
```

## Stage 3 — run side: `update/update.go` (`//go:build wasm`, package `update`)

```go
// OnReady calls fn once, when a new version of the application has been downloaded and waits to
// take over: immediately if one already waits, otherwise when the service worker reports it. It
// never calls fn when the browser has no service worker. Call it once, from the page.
func OnReady(fn func())

// Apply tells the waiting version to take over; the page then reloads by itself (the registration
// script does it). It returns ErrNoUpdate when no version is waiting.
func Apply() error

// ErrNoUpdate is returned by Apply when no new version is waiting.
var ErrNoUpdate error = noUpdate{}

type noUpdate struct{}

func (noUpdate) Error() string { return "pwa: no update is waiting" }
```

(Same pattern as `files.ErrNotExist` in `webtyp.com/files`: comparable with `==`, no `errors`
import.) `ErrNoUpdate` and `noUpdate` live in `update/errors.go` **without** a build tag, so the
build-side tests can also compare against it.

Implementation through `syscall/js` and `webtyp.com/await`, importing the constants from
`webtyp.com/pwa` (`protocol.go` has no build tag, so it compiles under wasm):
- `sw := js.Global().Get("navigator").Get("serviceWorker")`; undefined → `OnReady` returns, `Apply`
  returns `ErrNoUpdate`.
- `OnReady`: `fired` guard so `fn` runs at most once. In a goroutine: `reg, err :=
  await.Promise(sw.Call("getRegistration"))`; if `reg` is truthy, `reg.waiting` truthy and
  `sw.controller` truthy → call `fn` and return. Otherwise `js.Global().Call("addEventListener",
  pwa.EventUpdateReady, cb)` where `cb` calls `fn` (through the guard) and releases itself.
- `Apply`: `getRegistration` → no registration or no `waiting` → `ErrNoUpdate`; else
  `waiting.Call("postMessage", obj)` with `obj` a JS object `{type: pwa.MessageSkipWaiting}`.

## Stage 4 — tests (`tests/`)

Build side, `tests/generate_test.go` (`//go:build !wasm`):

| Test | Proves |
|---|---|
| `TestGolden` | a fixed `Config`, icons 192/512 → `New`; shell `/`, `/style.3f9a1c2b.css`, `/script.0d4e5f60.js`, `/client.9a8b7c6d.wasm` → `ServiceWorker`; `app.Manifest`, `worker.Script`, `app.HeadTags`, `app.RegisterScript` equal `tests/testdata/{manifest.webmanifest,sw.js,head.html,register.js}` byte for byte (write the golden files from the templates above, by hand, and review them) |
| `TestVersionIgnoresOrder` | same shell shuffled → same `Version`; one `Revision` changed → different `Version`; `len(Version) == 16` |
| `TestErrors` | one row per validation in the table, asserting the exact text, each from the function that owns it |
| `TestShortNameDefaultsToName` | manifest JSON `short_name` equals `Name` |
| `TestManifestIsValidJSON` | `json.Unmarshal` into a map works; `display == "standalone"`, `start_url == "/"` |
| `TestConsumerShaped` | the order a compiler follows: `New` → append `HeadTags` to an HTML string and `RegisterScript` to a JS string → SHA-256 of each as `Asset.Revision` → `ServiceWorker` → the script lists both URLs; changing the HTML changes `Version` |
| `TestServiceWorker_UsesProtocolNames` | the worker script contains `pwa.CachePrefix + worker.Version` and `pwa.MessageSkipWaiting`; the register script contains `pwa.ServiceWorkerPath` and `pwa.EventUpdateReady` |
| `TestServiceWorker_NoSkipWaitingOnInstall` | the text between `"install"` and `"activate"` does not contain `skipWaiting`; the file does not contain `clients.claim` |

Run side, `tests/update_test.go` (`//go:build wasm`, headless browser, the test page has no service
worker registered):

| Test | Proves |
|---|---|
| `TestApply_NoWaitingVersion` | `update.Apply()` returns `ErrNoUpdate` |
| `TestOnReady_NotCalledWithoutUpdate` | `OnReady(fn)`; dispatch nothing; after 100 ms `fn` was not called |
| `TestOnReady_CalledOnceOnEvent` | `OnReady(fn)`; dispatch `new Event(pwa.EventUpdateReady)` on `window` twice → `fn` called exactly once |

## Stage 5 — docs

- `README.md`: what it is (two halves), an "I want X → use Y" table (manifest, head tags and register
  script → `New`; service worker of a hashed shell → `App.ServiceWorker`;
  know a new version is ready → `update.OnReady`; switch to it → `update.Apply`), and one example of
  each. State that `sitec` calls `New` and `ServiceWorker` when a project declares
  `PWA() pwa.Config` — a project never calls them by hand.
- `docs/ARCHITECTURE.md`: the prior-art comparison of the design gate; the update sequence as a
  mermaid sequence diagram (deploy → browser fetches `sw.js` → new worker installs and precaches →
  waits → `EventUpdateReady` → user accepts → `MessageSkipWaiting` → `controllerchange` → reload);
  why no `skipWaiting`/`clients.claim`; what the worker does **not** cache (everything outside the
  shell); the HTTP headers each file needs (hashed names `public, max-age=31536000, immutable`;
  `/`, `sw.js`, `manifest.webmanifest` `no-cache`) — the server applies them, not this library.

## Acceptance

- `gotest` and `gotest -tinygo` green.
- `grep -rn "skipWaiting" sw.go` → only inside the `message` handler.
- `grep -rln "\"encoding/json\"\|\"strings\"\|\"errors\"\|\"fmt\"" update/` → empty.

| Stage | Files | Done when |
|---|---|---|
| 1 | `protocol.go` | constants |
| 2 | `pwa.go`, `version.go`, `manifest.go`, `sw.go` | `New`, `App.ServiceWorker`, validation |
| 3 | `update/update.go` | `OnReady`, `Apply` |
| 4 | `tests/*`, `tests/testdata/*` | tables green, golden files reviewed |
| 5 | `README.md`, `docs/ARCHITECTURE.md` | written |
