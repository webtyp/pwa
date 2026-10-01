//go:build wasm

package update

import (
	"syscall/js"

	"webtyp.com/await"
	"webtyp.com/pwa"
)

// OnReady calls fn once, when a new version of the application has been downloaded and waits to
// take over: immediately if one already waits, otherwise when the service worker reports it. It
// never calls fn when the browser has no service worker. Call it once, from the page.
func OnReady(fn func()) {
	sw := serviceWorker()
	if !present(sw) {
		return
	}
	win := js.Global()
	fired := false
	var cb js.Func
	// fire runs fn at most once and removes the listener before releasing it, so a later
	// event never reaches a released function.
	fire := func() {
		if fired {
			return
		}
		fired = true
		win.Call("removeEventListener", pwa.EventUpdateReady, cb)
		cb.Release()
		fn()
	}
	cb = js.FuncOf(func(js.Value, []js.Value) any {
		fire()
		return nil
	})
	// The listener is registered before OnReady returns: an event dispatched right after the
	// call must not be lost while the registration lookup below is still pending.
	win.Call("addEventListener", pwa.EventUpdateReady, cb)

	go func() {
		reg, err := await.Promise(sw.Call("getRegistration"))
		if err != nil || !present(reg) {
			return
		}
		if present(reg.Get("waiting")) && present(sw.Get("controller")) {
			fire()
		}
	}()
}

// Apply tells the waiting version to take over; the page then reloads by itself (the registration
// script does it). It returns ErrNoUpdate when no version is waiting.
func Apply() error {
	sw := serviceWorker()
	if !present(sw) {
		return ErrNoUpdate
	}
	reg, err := await.Promise(sw.Call("getRegistration"))
	if err != nil || !present(reg) {
		return ErrNoUpdate
	}
	waiting := reg.Get("waiting")
	if !present(waiting) {
		return ErrNoUpdate
	}
	msg := js.Global().Get("Object").New()
	msg.Set("type", pwa.MessageSkipWaiting)
	waiting.Call("postMessage", msg)
	return nil
}

func serviceWorker() js.Value {
	nav := js.Global().Get("navigator")
	if !present(nav) {
		return js.Undefined()
	}
	return nav.Get("serviceWorker")
}

func present(v js.Value) bool { return !v.IsUndefined() && !v.IsNull() }
