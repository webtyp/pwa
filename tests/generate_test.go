//go:build !wasm

package tests

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"webtyp.com/pwa"
)

func validConfig() pwa.Config {
	return pwa.Config{
		Name:            "Clínica Monjitas",
		ThemeColor:      "#0055ff",
		BackgroundColor: "#ffffff",
	}
}

func validIcons() []pwa.Icon {
	return []pwa.Icon{
		{URL: "/icon-192.png", Sizes: "192x192", Type: "image/png"},
		{URL: "/icon-512.png", Sizes: "512x512", Type: "image/png"},
	}
}

func validShell() []pwa.Asset {
	return []pwa.Asset{
		{URL: "/", Revision: "rev1"},
		{URL: "/style.3f9a1c2b.css", Revision: "rev2"},
		{URL: "/script.0d4e5f60.js", Revision: "rev3"},
		{URL: "/client.9a8b7c6d.wasm", Revision: "rev4"},
	}
}

func TestGenerate_Golden(t *testing.T) {
	cfg := validConfig()
	icons := validIcons()
	shell := validShell()

	b, err := pwa.Generate(cfg, icons, shell)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manifestGolden, err := os.ReadFile("testdata/manifest.webmanifest")
	if err != nil {
		t.Fatalf("failed to read golden manifest: %v", err)
	}
	if string(b.Manifest) != string(manifestGolden) {
		t.Errorf("Manifest mismatch.\nGot:  %s\nWant: %s", string(b.Manifest), string(manifestGolden))
	}

	headGolden, err := os.ReadFile("testdata/head.html")
	if err != nil {
		t.Fatalf("failed to read golden head.html: %v", err)
	}
	if b.HeadTags != string(headGolden) {
		t.Errorf("HeadTags mismatch.\nGot:  %s\nWant: %s", b.HeadTags, string(headGolden))
	}

	registerGolden, err := os.ReadFile("testdata/register.js")
	if err != nil {
		t.Fatalf("failed to read golden register.js: %v", err)
	}
	if b.RegisterScript != string(registerGolden) {
		t.Errorf("RegisterScript mismatch.\nGot:  %s\nWant: %s", b.RegisterScript, string(registerGolden))
	}

	swGolden, err := os.ReadFile("testdata/sw.js")
	if err != nil {
		t.Fatalf("failed to read golden sw.js: %v", err)
	}
	if string(b.ServiceWorker) != string(swGolden) {
		t.Errorf("ServiceWorker mismatch.\nGot:  %s\nWant: %s", string(b.ServiceWorker), string(swGolden))
	}
}

func TestGenerate_VersionIgnoresOrder(t *testing.T) {
	cfg := validConfig()
	icons := validIcons()

	shell1 := []pwa.Asset{
		{URL: "/", Revision: "r1"},
		{URL: "/a.js", Revision: "r2"},
		{URL: "/b.css", Revision: "r3"},
	}
	shell2 := []pwa.Asset{
		{URL: "/b.css", Revision: "r3"},
		{URL: "/", Revision: "r1"},
		{URL: "/a.js", Revision: "r2"},
	}

	b1, err := pwa.Generate(cfg, icons, shell1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b2, err := pwa.Generate(cfg, icons, shell2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b1.Version != b2.Version {
		t.Errorf("Versions should match for reordered shell assets: %q vs %q", b1.Version, b2.Version)
	}
	if len(b1.Version) != 16 {
		t.Errorf("Expected version length 16, got %d (%q)", len(b1.Version), b1.Version)
	}

	shell3 := []pwa.Asset{
		{URL: "/", Revision: "r1"},
		{URL: "/a.js", Revision: "r2_changed"},
		{URL: "/b.css", Revision: "r3"},
	}
	b3, err := pwa.Generate(cfg, icons, shell3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b1.Version == b3.Version {
		t.Errorf("Version should change when revision changes")
	}
}

func TestGenerate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     pwa.Config
		icons   []pwa.Icon
		shell   []pwa.Asset
		wantErr string
	}{
		{
			name:    "empty config name",
			cfg:     pwa.Config{Name: ""},
			icons:   validIcons(),
			shell:   validShell(),
			wantErr: "pwa: Config.Name is empty",
		},
		{
			name:    "invalid theme color",
			cfg:     pwa.Config{Name: "App", ThemeColor: "invalid"},
			icons:   validIcons(),
			shell:   validShell(),
			wantErr: `pwa: ThemeColor "invalid" is not #rgb or #rrggbb`,
		},
		{
			name:    "invalid background color",
			cfg:     pwa.Config{Name: "App", ThemeColor: "#fff", BackgroundColor: "blue"},
			icons:   validIcons(),
			shell:   validShell(),
			wantErr: `pwa: BackgroundColor "blue" is not #rgb or #rrggbb`,
		},
		{
			name: "missing 192x192 icon",
			cfg:  validConfig(),
			icons: []pwa.Icon{
				{URL: "/icon-512.png", Sizes: "512x512", Type: "image/png"},
			},
			shell:   validShell(),
			wantErr: "pwa: an installable app needs icons of 192x192 and 512x512; missing 192x192",
		},
		{
			name: "missing 512x512 icon",
			cfg:  validConfig(),
			icons: []pwa.Icon{
				{URL: "/icon-192.png", Sizes: "192x192", Type: "image/png"},
			},
			shell:   validShell(),
			wantErr: "pwa: an installable app needs icons of 192x192 and 512x512; missing 512x512",
		},
		{
			name:    "empty shell",
			cfg:     validConfig(),
			icons:   validIcons(),
			shell:   []pwa.Asset{},
			wantErr: "pwa: the shell is empty",
		},
		{
			name:  "asset URL not absolute",
			cfg:   validConfig(),
			icons: validIcons(),
			shell: []pwa.Asset{
				{URL: "style.css", Revision: "r1"},
			},
			wantErr: `pwa: asset URL "style.css" is not an absolute path`,
		},
		{
			name:  "asset missing revision",
			cfg:   validConfig(),
			icons: validIcons(),
			shell: []pwa.Asset{
				{URL: "/style.css", Revision: ""},
			},
			wantErr: `pwa: asset "/style.css" has no revision`,
		},
		{
			name:  "asset equals service worker path",
			cfg:   validConfig(),
			icons: validIcons(),
			shell: []pwa.Asset{
				{URL: pwa.ServiceWorkerPath, Revision: "r1"},
			},
			wantErr: "pwa: the service worker cannot precache itself",
		},
		{
			name:  "duplicate asset URL",
			cfg:   validConfig(),
			icons: validIcons(),
			shell: []pwa.Asset{
				{URL: "/", Revision: "r1"},
				{URL: "/style.css", Revision: "r2"},
				{URL: "/style.css", Revision: "r3"},
			},
			wantErr: `pwa: asset "/style.css" is listed twice`,
		},
		{
			name:  "missing root asset",
			cfg:   validConfig(),
			icons: validIcons(),
			shell: []pwa.Asset{
				{URL: "/index.html", Revision: "r1"},
			},
			wantErr: `pwa: the shell does not include "/"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pwa.Generate(tc.cfg, tc.icons, tc.shell)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("error mismatch.\nGot:  %q\nWant: %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestGenerate_ShortNameDefaultsToName(t *testing.T) {
	cfg := validConfig()
	cfg.ShortName = ""
	b, err := pwa.Generate(cfg, validIcons(), validShell())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b.Manifest, &m); err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
	}
	if m["short_name"] != cfg.Name {
		t.Errorf("short_name should default to Name %q, got %q", cfg.Name, m["short_name"])
	}
}

func TestGenerate_ManifestIsValidJSON(t *testing.T) {
	b, err := pwa.Generate(validConfig(), validIcons(), validShell())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b.Manifest, &m); err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
	}

	if m["display"] != "standalone" {
		t.Errorf("expected display == standalone, got %v", m["display"])
	}
	if m["start_url"] != "/" {
		t.Errorf("expected start_url == /, got %v", m["start_url"])
	}
}

func TestServiceWorker_UsesProtocolNames(t *testing.T) {
	b, err := pwa.Generate(validConfig(), validIcons(), validShell())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	swStr := string(b.ServiceWorker)
	if !strings.Contains(swStr, pwa.CachePrefix+b.Version) {
		t.Errorf("SW expected to contain cache name with prefix and version")
	}
	if !strings.Contains(swStr, pwa.MessageSkipWaiting) {
		t.Errorf("SW expected to contain MessageSkipWaiting constant")
	}

	regStr := b.RegisterScript
	if !strings.Contains(regStr, pwa.ServiceWorkerPath) {
		t.Errorf("RegisterScript expected to contain ServiceWorkerPath")
	}
	if !strings.Contains(regStr, pwa.EventUpdateReady) {
		t.Errorf("RegisterScript expected to contain EventUpdateReady")
	}
}

func TestServiceWorker_NoSkipWaitingOnInstall(t *testing.T) {
	b, err := pwa.Generate(validConfig(), validIcons(), validShell())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	swStr := string(b.ServiceWorker)

	installIdx := strings.Index(swStr, `"install"`)
	activateIdx := strings.Index(swStr, `"activate"`)
	if installIdx == -1 || activateIdx == -1 || installIdx >= activateIdx {
		t.Fatalf("could not find install and activate event blocks in expected order")
	}

	installBlock := swStr[installIdx:activateIdx]
	if strings.Contains(installBlock, "skipWaiting") {
		t.Errorf("skipWaiting must not appear inside install handler")
	}

	if strings.Contains(swStr, "clients.claim") {
		t.Errorf("SW must not contain clients.claim")
	}
}
