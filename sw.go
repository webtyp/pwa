//go:build !wasm

package pwa

import (
	"encoding/json"
	"sort"
	"strings"
)

// No skipWaiting() on install and no clients.claim() — a new version waits until the page applies
// it (D-PWA-8), so a page never runs HTML of one version with wasm of another.
const swTemplate = `const CACHE = "{{CACHE_PREFIX}}{{VERSION}}";
const SHELL = {{SHELL}};
self.addEventListener("install", (e) => {
  e.waitUntil(caches.open(CACHE).then((c) => c.addAll(SHELL.map((u) => new Request(u, { cache: "reload" })))));
});
self.addEventListener("activate", (e) => {
  e.waitUntil(caches.keys().then((keys) => Promise.all(
    keys.filter((k) => k.startsWith("{{CACHE_PREFIX}}") && k !== CACHE).map((k) => caches.delete(k)))));
});
self.addEventListener("message", (e) => {
  if (e.data && e.data.type === "{{MESSAGE_SKIP_WAITING}}") self.skipWaiting();
});
self.addEventListener("fetch", (e) => {
  const req = e.request;
  if (req.method !== "GET") return;
  const url = new URL(req.url);
  if (url.origin !== self.location.origin || !SHELL.includes(url.pathname)) return;
  e.respondWith(caches.open(CACHE).then((c) => c.match(url.pathname)).then((r) => r || fetch(req)));
});`

const registerScriptTemplate = `if ("serviceWorker" in navigator) {
  const sw = navigator.serviceWorker;
  const hadController = !!sw.controller;
  let reloading = false;
  sw.addEventListener("controllerchange", () => {
    if (!hadController || reloading) return;
    reloading = true;
    location.reload();
  });
  sw.register("{{SERVICE_WORKER_PATH}}").then((reg) => {
    const notify = () => window.dispatchEvent(new Event("{{EVENT_UPDATE_READY}}"));
    const watch = (w) => w && w.addEventListener("statechange", () => {
      if (w.state === "installed" && sw.controller) notify();
    });
    if (reg.waiting && sw.controller) notify();
    watch(reg.installing);
    reg.addEventListener("updatefound", () => watch(reg.installing));
  });
}`

func generateServiceWorker(version string, shell []Asset) ([]byte, error) {
	urls := make([]string, len(shell))
	for i, a := range shell {
		urls[i] = a.URL
	}
	sort.Strings(urls)

	shellJSON, err := json.Marshal(urls)
	if err != nil {
		return nil, err
	}

	res := swTemplate
	res = strings.ReplaceAll(res, "{{CACHE_PREFIX}}", CachePrefix)
	res = strings.ReplaceAll(res, "{{VERSION}}", version)
	res = strings.ReplaceAll(res, "{{SHELL}}", string(shellJSON))
	res = strings.ReplaceAll(res, "{{MESSAGE_SKIP_WAITING}}", MessageSkipWaiting)

	return []byte(res), nil
}

func generateRegisterScript() string {
	res := registerScriptTemplate
	res = strings.ReplaceAll(res, "{{SERVICE_WORKER_PATH}}", ServiceWorkerPath)
	res = strings.ReplaceAll(res, "{{EVENT_UPDATE_READY}}", EventUpdateReady)
	return res
}
