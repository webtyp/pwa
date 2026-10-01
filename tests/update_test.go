//go:build wasm

package tests

import (
	"sync/atomic"
	"syscall/js"
	"testing"
	"time"

	"webtyp.com/pwa"
	"webtyp.com/pwa/update"
)

func TestApply_NoWaitingVersion(t *testing.T) {
	err := update.Apply()
	if err != update.ErrNoUpdate {
		t.Errorf("expected ErrNoUpdate, got %v", err)
	}
}

func TestOnReady_NotCalledWithoutUpdate(t *testing.T) {
	var called int32
	update.OnReady(func() {
		atomic.AddInt32(&called, 1)
	})

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&called) != 0 {
		t.Errorf("OnReady callback was called without update ready signal")
	}
}

func TestOnReady_CalledOnceOnEvent(t *testing.T) {
	var called int32
	update.OnReady(func() {
		atomic.AddInt32(&called, 1)
	})

	// DispatchEvent on window twice
	win := js.Global()
	evtCtor := win.Get("Event")
	if !evtCtor.IsUndefined() && !evtCtor.IsNull() {
		evt1 := evtCtor.New(pwa.EventUpdateReady)
		win.Call("dispatchEvent", evt1)
		evt2 := evtCtor.New(pwa.EventUpdateReady)
		win.Call("dispatchEvent", evt2)
	}

	time.Sleep(50 * time.Millisecond)

	if count := atomic.LoadInt32(&called); count != 1 {
		t.Errorf("expected OnReady callback to be called exactly once, got %d", count)
	}
}
