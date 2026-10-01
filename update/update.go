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
	nav := js.Global().Get("navigator")
	if nav.IsUndefined() || nav.IsNull() {
		return
	}
	sw := nav.Get("serviceWorker")
	if sw.IsUndefined() || sw.IsNull() {
		return
	}

	fired := false
	callOnce := func() {
		if !fired {
			fired = true
			fn()
		}
	}

	go func() {
		regVal, err := await.Promise(sw.Call("getRegistration"))
		if err == nil && !regVal.IsUndefined() && !regVal.IsNull() {
			waiting := regVal.Get("waiting")
			controller := sw.Get("controller")
			if !waiting.IsUndefined() && !waiting.IsNull() && !controller.IsUndefined() && !controller.IsNull() {
				callOnce()
				return
			}
		}

		var cb js.Func
		cb = js.FuncOf(func(this js.Value, args []js.Value) any {
			callOnce()
			cb.Release()
			return nil
		})
		js.Global().Call("addEventListener", pwa.EventUpdateReady, cb)
	}()
}

// Apply tells the waiting version to take over; the page then reloads by itself (the registration
// script does it). It returns ErrNoUpdate when no version is waiting.
func Apply() error {
	nav := js.Global().Get("navigator")
	if nav.IsUndefined() || nav.IsNull() {
		return ErrNoUpdate
	}
	sw := nav.Get("serviceWorker")
	if sw.IsUndefined() || sw.IsNull() {
		return ErrNoUpdate
	}

	regVal, err := await.Promise(sw.Call("getRegistration"))
	if err != nil || regVal.IsUndefined() || regVal.IsNull() {
		return ErrNoUpdate
	}

	waiting := regVal.Get("waiting")
	if waiting.IsUndefined() || waiting.IsNull() {
		return ErrNoUpdate
	}

	msgObj := js.Global().Get("Object").New()
	msgObj.Set("type", pwa.MessageSkipWaiting)
	waiting.Call("postMessage", msgObj)
	return nil
}
