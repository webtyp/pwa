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

// App is what makes an application installable. Build it with New.
type App struct {
	Manifest       []byte // serve at ManifestPath
	HeadTags       string // insert in <head> of every HTML page
	RegisterScript string // append to the page's main script
}

// Worker is the service worker of one exact set of shell assets.
type Worker struct {
	Script  []byte // serve at ServiceWorkerPath
	Version string // 16 hex chars; identifies this exact set of shell assets
}

const (
	errEmptyName      = "pwa: Config.Name is empty"
	errBadColor       = "pwa: %s %q is not #rgb or #rrggbb"
	errMissingIcon    = "pwa: an installable app needs icons of 192x192 and 512x512; missing %s"
	errEmptyShell     = "pwa: the shell is empty"
	errNotAbsolute    = "pwa: asset URL %q is not an absolute path"
	errNoRevision     = "pwa: asset %q has no revision"
	errSelfPrecache   = "pwa: the service worker cannot precache itself"
	errListedTwice    = "pwa: asset %q is listed twice"
	errNoRoot         = `pwa: the shell does not include "/"`
	fieldThemeColor   = "ThemeColor"
	fieldBackground   = "BackgroundColor"
	iconSizesSmall    = "192x192"
	iconSizesLarge    = "512x512"
	rootURL           = "/"
	absolutePathStart = "/"
)

var hexColorRegexp = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func validateColor(field, value string) error {
	if !hexColorRegexp.MatchString(value) {
		return fmt.Errorf(errBadColor, field, value)
	}
	return nil
}

// New validates the config and icons and returns the files that do not depend on the shell:
// insert HeadTags and RegisterScript before hashing the shell, then call ServiceWorker.
func New(c Config, icons []Icon) (App, error) {
	if c.Name == "" {
		return App{}, errors.New(errEmptyName)
	}
	if err := validateColor(fieldThemeColor, c.ThemeColor); err != nil {
		return App{}, err
	}
	if err := validateColor(fieldBackground, c.BackgroundColor); err != nil {
		return App{}, err
	}
	for _, size := range []string{iconSizesSmall, iconSizesLarge} {
		if !hasIcon(icons, size) {
			return App{}, fmt.Errorf(errMissingIcon, size)
		}
	}
	manifest, err := generateManifest(c, icons)
	if err != nil {
		return App{}, err
	}
	return App{
		Manifest:       manifest,
		HeadTags:       generateHeadTags(c.ThemeColor),
		RegisterScript: generateRegisterScript(),
	}, nil
}

func hasIcon(icons []Icon, sizes string) bool {
	for _, ic := range icons {
		if ic.Sizes == sizes {
			return true
		}
	}
	return false
}

// ServiceWorker returns the service worker that precaches shell. The shell must list every file
// the page needs to start, including "/" (the HTML page, after HeadTags were inserted); it must
// not list ServiceWorkerPath. It is a method only so that it cannot be called without a
// validated App.
func (a App) ServiceWorker(shell []Asset) (Worker, error) {
	if err := validateShell(shell); err != nil {
		return Worker{}, err
	}
	version := computeVersion(shell)
	script, err := generateServiceWorker(version, shell)
	if err != nil {
		return Worker{}, err
	}
	return Worker{Script: script, Version: version}, nil
}

func validateShell(shell []Asset) error {
	if len(shell) == 0 {
		return errors.New(errEmptyShell)
	}
	seen := make(map[string]bool, len(shell)) // build side: never compiled to wasm
	hasRoot := false
	for _, a := range shell {
		if !strings.HasPrefix(a.URL, absolutePathStart) {
			return fmt.Errorf(errNotAbsolute, a.URL)
		}
		if a.Revision == "" {
			return fmt.Errorf(errNoRevision, a.URL)
		}
		if a.URL == ServiceWorkerPath {
			return errors.New(errSelfPrecache)
		}
		if seen[a.URL] {
			return fmt.Errorf(errListedTwice, a.URL)
		}
		seen[a.URL] = true
		if a.URL == rootURL {
			hasRoot = true
		}
	}
	if !hasRoot {
		return errors.New(errNoRoot)
	}
	return nil
}
