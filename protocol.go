package pwa

// Names shared by the files Generate writes and the page code in pwa/update.
const (
	ManifestPath       = "/manifest.webmanifest" // where the manifest must be served
	ServiceWorkerPath  = "/sw.js"                // where the service worker must be served (scope "/")
	CachePrefix        = "webtyp-shell-"         // Cache Storage name = CachePrefix + Build.Version
	EventUpdateReady   = "webtyp-update-ready"   // dispatched on window when a new version waits
	MessageSkipWaiting = "webtyp-skip-waiting"   // posted to the waiting worker to take over
)
