//go:build !wasm

package pwa

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

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

// Build is everything an application needs to be installable and start offline.
type Build struct {
	Manifest       []byte // serve at ManifestPath
	ServiceWorker  []byte // serve at ServiceWorkerPath
	HeadTags       string // insert in <head> of every HTML page
	RegisterScript string // append to the page's main script
	Version        string // 16 hex chars; identifies this exact set of shell assets
}

var hexColorRegexp = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func validateColor(field, value string) error {
	if !hexColorRegexp.MatchString(value) {
		return fmt.Errorf("pwa: %s %q is not #rgb or #rrggbb", field, value)
	}
	return nil
}

// Generate validates the inputs and returns the files. The shell must list every file the page
// needs to start, including "/" (the HTML page); it must not list ServiceWorkerPath.
func Generate(c Config, icons []Icon, shell []Asset) (Build, error) {
	if c.Name == "" {
		return Build{}, errors.New("pwa: Config.Name is empty")
	}

	if err := validateColor("ThemeColor", c.ThemeColor); err != nil {
		return Build{}, err
	}
	if err := validateColor("BackgroundColor", c.BackgroundColor); err != nil {
		return Build{}, err
	}

	has192 := false
	has512 := false
	for _, ic := range icons {
		if ic.Sizes == "192x192" {
			has192 = true
		}
		if ic.Sizes == "512x512" {
			has512 = true
		}
	}
	if !has192 {
		return Build{}, errors.New("pwa: an installable app needs icons of 192x192 and 512x512; missing 192x192")
	}
	if !has512 {
		return Build{}, errors.New("pwa: an installable app needs icons of 192x192 and 512x512; missing 512x512")
	}

	if len(shell) == 0 {
		return Build{}, errors.New("pwa: the shell is empty")
	}

	seenURLs := make(map[string]bool)
	hasRoot := false

	for _, a := range shell {
		if !strings.HasPrefix(a.URL, "/") {
			return Build{}, fmt.Errorf("pwa: asset URL %q is not an absolute path", a.URL)
		}
		if a.Revision == "" {
			return Build{}, fmt.Errorf("pwa: asset %q has no revision", a.URL)
		}
		if a.URL == ServiceWorkerPath {
			return Build{}, errors.New("pwa: the service worker cannot precache itself")
		}
		if seenURLs[a.URL] {
			return Build{}, fmt.Errorf("pwa: asset %q is listed twice", a.URL)
		}
		seenURLs[a.URL] = true
		if a.URL == "/" {
			hasRoot = true
		}
	}

	if !hasRoot {
		return Build{}, errors.New("pwa: the shell does not include \"/\"")
	}

	ver := computeVersion(shell)
	manifestBytes, err := generateManifest(c, icons)
	if err != nil {
		return Build{}, err
	}

	swBytes, err := generateServiceWorker(ver, shell)
	if err != nil {
		return Build{}, err
	}

	headTags := generateHeadTags(c.ThemeColor)
	regScript := generateRegisterScript()

	return Build{
		Manifest:       manifestBytes,
		ServiceWorker:  swBytes,
		HeadTags:       headTags,
		RegisterScript: regScript,
		Version:        ver,
	}, nil
}
