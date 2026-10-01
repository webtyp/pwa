package update

// ErrNoUpdate is returned by Apply when no new version is waiting.
var ErrNoUpdate error = noUpdate{}

type noUpdate struct{}

func (noUpdate) Error() string { return "pwa: no update is waiting" }
