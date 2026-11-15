package plugin

import (
	"log"
	"runtime/debug"
)

// The C ABI bridge (main.go) recovers panics of synchronous host calls only. A
// panic in a background goroutine would take the whole CLIProxyAPI process down,
// so every goroutine this plugin starts goes through safeGo or runRecovered.

// logPanic is replaceable in tests.
var logPanic = func(name string, value any, stack []byte) {
	log.Printf("plugin: %s panic: %v\n%s", name, value, stack)
}

// runRecovered runs fn and reports whether it returned normally. A panic is
// logged with its stack and swallowed; it is never re-raised. Deferred calls
// inside fn (unlocks, WaitGroup.Done) still run while the panic unwinds.
func runRecovered(name string, fn func()) (ok bool) {
	defer func() {
		if v := recover(); v != nil {
			logPanic(name, v, debug.Stack())
			ok = false
		}
	}()
	fn()
	return true
}

// safeGo starts fn in a new goroutine guarded by runRecovered.
func safeGo(name string, fn func()) {
	go runRecovered(name, fn)
}
