//go:build !wasm

package pwa

import (
	"encoding/json"
)

type manifestIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

type manifestJSON struct {
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	Description     string         `json:"description,omitempty"`
	StartURL        string         `json:"start_url"`
	Scope           string         `json:"scope"`
	Display         string         `json:"display"`
	BackgroundColor string         `json:"background_color"`
	ThemeColor      string         `json:"theme_color"`
	Icons           []manifestIcon `json:"icons"`
}

func generateManifest(c Config, icons []Icon) ([]byte, error) {
	shortName := c.ShortName
	if shortName == "" {
		shortName = c.Name
	}

	mIcons := make([]manifestIcon, len(icons))
	for i, ic := range icons {
		mIcons[i] = manifestIcon{
			Src:   ic.URL,
			Sizes: ic.Sizes,
			Type:  ic.Type,
		}
	}

	m := manifestJSON{
		Name:            c.Name,
		ShortName:       shortName,
		Description:     c.Description,
		StartURL:        "/",
		Scope:           "/",
		Display:         "standalone",
		BackgroundColor: c.BackgroundColor,
		ThemeColor:      c.ThemeColor,
		Icons:           mIcons,
	}

	return json.Marshal(m)
}

func generateHeadTags(themeColor string) string {
	return `<link rel="manifest" href="` + ManifestPath + `"><meta name="theme-color" content="` + themeColor + `">`
}
