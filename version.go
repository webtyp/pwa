//go:build !wasm

package pwa

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func computeVersion(shell []Asset) string {
	sorted := make([]Asset, len(shell))
	copy(sorted, shell)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].URL < sorted[j].URL
	})

	h := sha256.New()
	for _, a := range sorted {
		h.Write([]byte(a.URL + "\t" + a.Revision + "\n"))
	}
	digest := h.Sum(nil)
	return hex.EncodeToString(digest[:8])
}
