const CACHE = "webtyp-shell-0d20add74192f96e";
const SHELL = ["/","/client.9a8b7c6d.wasm","/script.0d4e5f60.js","/style.3f9a1c2b.css"];
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