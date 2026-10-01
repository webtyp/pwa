//go:build !wasm

package tests

import (
	"testing"

	"webtyp.com/pwa"
)

func TestHashedNameIsRecognised(t *testing.T) {
	for _, name := range []string{"style.css", "client.wasm", "LICENSE", "a.b.js"} {
		h := pwa.HashedName(name, []byte("content of "+name))
		if !pwa.IsHashedName("/" + h) {
			t.Errorf("IsHashedName(%q) = false for a name HashedName wrote", h)
		}
		if pwa.IsHashedName("/" + name) {
			t.Errorf("IsHashedName(%q) = true for an unhashed name", name)
		}
	}
	if pwa.HashedName("style.css", []byte("x")) != pwa.HashedName("style.css", []byte("x")) {
		t.Error("HashedName is not deterministic")
	}
}

func TestIsHashedName_RejectsLookalikes(t *testing.T) {
	for _, p := range []string{"/", "/sw.js", "/manifest.webmanifest", "/style.3F9A1C2B.css", "/style.3f9a1c2.css", "/.3f9a1c2b.css", "/icons/style.css"} {
		if pwa.IsHashedName(p) {
			t.Errorf("IsHashedName(%q) = true", p)
		}
	}
	if !pwa.IsHashedName("/assets/style.3f9a1c2b.css") || !pwa.IsHashedName("style.3f9a1c2b.css") {
		t.Error("hashed names under a prefix or relative must be recognised")
	}
}

func TestCacheControl(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/", pwa.CacheRevalidate},
		{pwa.ServiceWorkerPath, pwa.CacheRevalidate},
		{pwa.ManifestPath, pwa.CacheRevalidate},
		{"/icon-192.png", pwa.CacheRevalidate},
		{"/style.3f9a1c2b.css", pwa.CacheImmutable},
		{"/client.9a8b7c6d.wasm", pwa.CacheImmutable},
		{pwa.ArtifactsDir + "decider-0.8b.q4.wtypw", pwa.CacheNoStore},
	}
	for _, c := range cases {
		if got := pwa.CacheControl(c.path); got != c.want {
			t.Errorf("CacheControl(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}
