---
PLAN: "feat(ui): pwa/ui.UpdatePrompt — \"New version · Reload\" when a new version waits (D-PWA-8)"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Prerequisites (published tags):** `webtyp.com/components` **v0.8.26**, `webtyp.com/dom`,
> `webtyp.com/html`, `webtyp.com/lang`. If any is missing, stop and report it.

# Plan — `pwa/ui`: tell the person a new version is ready

## 0. Context (read first)

`webtyp.com/pwa` (this repo) builds the manifest and the service worker of an installable web app.
Decision D-PWA-8 (<https://github.com/webtyp/app/blob/main/docs/PWA_ARTIFACTS_MASTER_PLAN.md>): a
new version downloads in the background and **waits**; the page never reloads by itself in the
middle of someone's work. The page learns that a version waits with `update.OnReady(fn)` and makes
it take over with `update.Apply()` (package `webtyp.com/pwa/update`, `//go:build wasm`; read
`update/update.go`). What is missing is the visible part, and every PWA app needs the same one: a
small "New version · Reload" button that appears only when a version waits.

Apps place it in the header slot `platformd.Platform.Status func() Component`
(`webtyp.com/layout` v0.3.44). This package only builds the component; it never imports `layout`.
House pattern: a library's UI lives in its own `ui/` sub-package, so the core package (`pwa`,
used by the build tool `sitec`) stays free of UI code.

## Design gate

### 1. Prior art
- **Workbox `workbox-window` + "A new version is available — Reload" snackbar** (Google's PWA
  guidance): listen for the waiting worker, ask the person, then `skipWaiting`. Ours: `OnReady` →
  button → `Apply`.
- **Angular `SwUpdate.versionUpdates` + `activateUpdate()`**: same flow, the app owns the prompt.
- **VS Code / Slack desktop "Restart to update" button:** appears only when an update is ready and
  never restarts by itself.

### 2. Novice-name test
`ui.UpdatePrompt()` → "the update prompt".

### 3. Complexity ledger
```
Concepts the developer must learn   +1 (UpdatePrompt)
Files they must touch to do X        +1 line in the app's header slot
Lines at the call site               1
Ways to do the same thing            0 (no app showed it before)
```

### 4. Where it belongs
`webtyp.com/pwa/ui`, next to `pwa/update` which it drives (house pattern for a library's UI). It
composes `components/actionbutton` and never re-creates a button.

### 5. What this deletes
Nothing (new capability).

## 1. API (`ui/`, package `ui`)

```go
// UpdatePrompt is a "New version · Reload" button that stays hidden until a new version of the app
// has been downloaded and waits (update.OnReady); clicking it lets that version take over
// (update.Apply) and the page reloads. Outside the browser it renders hidden and does nothing.
func UpdatePrompt() Component
```

Nothing else is exported.

### 1.1 Files
- `ui/prompt.go` (no build tag): the unexported struct (embeds `dom.Element` by value) with
  `visible *SignalBool` (starts `false`); `Render()` = a `Span` with the root class and
  `dom.Show(visible, …)` of an `actionbutton.ActionButton{Text: "New version · Reload", Variant:
  "primary", OnClick: …}`. The click calls an unexported `apply()`. `UpdatePrompt()` builds the
  struct and calls the unexported `arm(p)`.
- `ui/prompt_wasm.go` (`//go:build wasm`): `arm(p)` calls `update.OnReady(func() { p.visible.Set(true) })`;
  `apply()` calls `update.Apply()` and, on `update.ErrNoUpdate`, hides the button again.
- `ui/prompt_stdlib.go` (`//go:build !wasm`): `arm` and `apply` do nothing.
- `widget.Name("updateprompt")`, `WidgetKind()` = `widget.Region`. Classes only through
  `Name.Root()`; never a class literal.

### 1.2 Text and translation
`"New version · Reload"` is a `lang.Text` (English). Add a `lang.json` at the module root, same
format as <https://github.com/webtyp/components/blob/main/lang.json>:
`{"languages": ["es"], "keys": {"New version · Reload": ["Versión nueva · Recargar"]}}` (skill
**translations**).

## 2. Stages

| Stage | Files | Content |
|---|---|---|
| 1 | `go.mod`, `go.sum` | `go get` the deps of the header |
| 2 | `ui/prompt.go`, `ui/prompt_wasm.go`, `ui/prompt_stdlib.go` | §1.1 |
| 3 | `lang.json` | §1.2 |
| 4 | `tests/ui_test.go` | §3 |
| 5 | `README.md`, `docs/ARCHITECTURE.md` | "Showing the update": the one-line app wiring `Status: func() Component { return ui.UpdatePrompt() }` and that the button appears only when a version waits |

## 3. Tests (`tests/ui_test.go`, `//go:build !wasm`)
1. `ui.UpdatePrompt().Render().String()` contains the root class `updateprompt` and the button is
   hidden (rendered inside the hidden `Show`, as `dom.Show` serialises it for a false signal).
2. Every existing test keeps passing (`update_test.go` and the others).

## 4. Code rules (non-negotiable)
- `ui/` compiles to TinyGo WASM: `webtyp.com/fmt`; no `errors`, `strconv`, `strings`, `map`.
  `syscall/js` is never imported by `ui/` (it uses `update`).
- Dot-import `webtyp.com/dom` and `webtyp.com/html` (house style). Never compose an element id.
- The root package `pwa` must **not** import `ui` or any UI package.
- No exported symbol beyond §1. Tests in `tests/`; never export for a test.

## 5. Acceptance criteria
- `gotest ./...` green (stdlib and wasm).
- `go list -deps . | grep -E "webtyp.com/(dom|html|components)"` (run at the module root) → empty.
- **Before opening the PR:** `git diff --stat main` lists every file of the stages table, and no file
  is deleted that this plan does not order deleted.
