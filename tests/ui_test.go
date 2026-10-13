//go:build !wasm

package tests

import (
	"strings"
	"testing"

	"webtyp.com/pwa/ui"
)

func TestUpdatePrompt(t *testing.T) {
	prompt := ui.UpdatePrompt()
	html := prompt.String()

	if !strings.Contains(html, "updateprompt") {
		t.Errorf("expected String() to contain root class 'updateprompt', got %q", html)
	}

	if strings.Contains(html, "New version") {
		t.Errorf("expected button to be hidden, got %q", html)
	}
}
