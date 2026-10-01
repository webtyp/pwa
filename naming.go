package pwa

// The naming and caching contract between the compiler that writes the files (sitec), the
// servers that send them (server/httpd, goflare) and the service worker. No build tag and no
// imports: it compiles everywhere, and both sides read the same rule instead of re-deriving it.

// HashLen is how many lowercase hex characters of the content's SHA-256 a hashed name carries:
// style.3f9a1c2b.css.
const HashLen = 8

// ArtifactsDir is where large artifacts (model weights, caches) are served from. They are never
// part of the precached shell and never HTTP-cached: webtyp.com/artifacts keeps them in OPFS, and
// an HTTP cache would hold a second copy of hundreds of MB.
const ArtifactsDir = "/artifacts/"

// Cache-Control values, one per kind of file.
const (
	CacheImmutable  = "public, max-age=31536000, immutable" // a content-hashed name never changes
	CacheRevalidate = "no-cache"                            // a fixed name (/, sw.js, the manifest): ask every time
	CacheNoStore    = "no-store"                            // large artifacts: OPFS keeps them, not the HTTP cache
)

// IsHashedName reports whether the last path segment has the form <name>.<HashLen hex>.<ext>
// or <name>.<HashLen hex>, as written by HashedName.
func IsHashedName(urlPath string) bool {
	base := urlPath
	for i := len(urlPath) - 1; i >= 0; i-- {
		if urlPath[i] == '/' {
			base = urlPath[i+1:]
			break
		}
	}
	// Try "<name>.<hash>.<ext>" first, then "<name>.<hash>".
	end := len(base)
	if dot := lastDot(base, end); dot > 0 && isHashAt(base, dot) {
		return true
	}
	return isHashAt(base, end)
}

// isHashAt reports whether base[end-HashLen:end] is lowercase hex preceded by a dot that is not
// the first character.
func isHashAt(base string, end int) bool {
	start := end - HashLen
	if start < 2 || base[start-1] != '.' {
		return false
	}
	for i := start; i < end; i++ {
		c := base[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func lastDot(s string, end int) int {
	for i := end - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

// CacheControl returns the Cache-Control header a server sends for urlPath: CacheNoStore under
// ArtifactsDir, CacheImmutable for a hashed name, CacheRevalidate for everything else.
func CacheControl(urlPath string) string {
	if len(urlPath) >= len(ArtifactsDir) && urlPath[:len(ArtifactsDir)] == ArtifactsDir {
		return CacheNoStore
	}
	if IsHashedName(urlPath) {
		return CacheImmutable
	}
	return CacheRevalidate
}
