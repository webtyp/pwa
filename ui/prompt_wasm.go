//go:build wasm

package ui

import "webtyp.com/pwa/update"

func arm(p *prompt) {
	update.OnReady(func() {
		p.visible.Set(true)
	})
}

func (p *prompt) apply() {
	if err := update.Apply(); err == update.ErrNoUpdate {
		p.visible.Set(false)
	}
}
