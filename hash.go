//go:build !wasm

package pwa

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"strings"
)

// HashedName inserts the first HashLen hex characters of the content's SHA-256 before the
// extension: HashedName("style.css", c) == "style.3f9a1c2b.css"; a name without extension gets
// ".<hash>". It is the only function that writes such names; IsHashedName recognises them.
func HashedName(name string, content []byte) string {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])[:HashLen]
	ext := path.Ext(name)
	if ext == "" {
		return name + "." + hash
	}
	return strings.TrimSuffix(name, ext) + "." + hash + ext
}
