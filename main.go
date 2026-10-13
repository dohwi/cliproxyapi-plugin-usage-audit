// Package main implements the usage-audit plugin for CLIProxyAPI v7.
// It receives usage records via the native C ABI plugin interface (usage.handle),
// stores them in an append-only JSONL log, and exposes aggregation/query
// routes under /v0/management/plugins/usage-audit/.
//
// Safety contracts:
//   - Storage/queue errors never block or crash the request path.
//   - Only AuthIndex (stable runtime index) is stored; raw tokens/hashes are never persisted.
//   - zero-token Detail fields are stored as-is (zero ≠ missing).
//   - Generate=false (count-tokens-only) is preserved and never conflated with true.
//   - Bounded async queue (capacity 2048); overflow drops the incoming record with a counter.
package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
	void*  ptr;
	size_t len;
} cliproxy_buffer;

typedef int  (*cliproxy_host_call_fn)(void*, const char*, const uint8_t*, size_t, cliproxy_buffer*);
typedef void (*cliproxy_host_free_fn)(void*, size_t);

typedef struct {
	uint32_t               abi_version;
	void*                  host_ctx;
	cliproxy_host_call_fn  call;
	cliproxy_host_free_fn  free_buffer;
} cliproxy_host_api;

typedef int  (*cliproxy_plugin_call_fn)(char*, uint8_t*, size_t, cliproxy_buffer*);
typedef void (*cliproxy_plugin_free_fn)(void*, size_t);
typedef void (*cliproxy_plugin_shutdown_fn)(void);

typedef struct {
	uint32_t                  abi_version;
	cliproxy_plugin_call_fn   call;
	cliproxy_plugin_free_fn   free_buffer;
	cliproxy_plugin_shutdown_fn shutdown;
} cliproxy_plugin_api;

extern int  cliproxyPluginCall(char*, uint8_t*, size_t, cliproxy_buffer*);
extern void cliproxyPluginFree(void*, size_t);
extern void cliproxyPluginShutdown(void);

static const cliproxy_host_api* stored_host;

static void store_host_api(const cliproxy_host_api* host) { stored_host = host; }
*/
import "C"

import (
	"encoding/json"
	"unsafe"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/plugin"
)

// abiVersion must match pluginabi.ABIVersion = 1.
const (
	abiVersion    uint32 = 1
	maxRequestLen        = 8 << 20 // Bound JSON envelopes before copying caller-owned memory.
	maxCIntValue         = uint64(1<<31 - 1)
)

func main() {}

// cliproxy_plugin_init is the C ABI entry point called by the host after dlopen.
//
//export cliproxy_plugin_init
func cliproxy_plugin_init(host *C.cliproxy_host_api, p *C.cliproxy_plugin_api) (rc C.int) {
	defer func() {
		if recover() != nil {
			rc = 1
		}
	}()
	if p == nil || (host != nil && uint32(host.abi_version) != abiVersion) {
		return 1
	}
	C.store_host_api(host)
	p.abi_version = C.uint32_t(abiVersion)
	p.call = C.cliproxy_plugin_call_fn(C.cliproxyPluginCall)
	p.free_buffer = C.cliproxy_plugin_free_fn(C.cliproxyPluginFree)
	p.shutdown = C.cliproxy_plugin_shutdown_fn(C.cliproxyPluginShutdown)
	return 0
}

// cliproxyPluginCall dispatches all host→plugin RPC calls.
//
//export cliproxyPluginCall
func cliproxyPluginCall(method *C.char, request *C.uint8_t, requestLen C.size_t, response *C.cliproxy_buffer) (rc C.int) {
	if response != nil {
		response.ptr = nil
		response.len = 0
	}
	defer func() {
		if recover() != nil {
			// Panic values may contain private data. Never unwind through C.
			_ = writeResponse(response, plugin.ErrorEnvelope("internal_panic", "plugin call failed"))
			rc = 1
		}
	}()
	if response == nil {
		return 1
	}
	if method == nil {
		_ = writeResponse(response, plugin.ErrorEnvelope("invalid_method", "method is required"))
		return 1
	}
	if requestLen > C.size_t(maxRequestLen) {
		_ = writeResponse(response, plugin.ErrorEnvelope("request_too_large", "request exceeds maximum length"))
		return 1
	}
	if (request == nil && requestLen != 0) || uint64(requestLen) > maxCIntValue {
		_ = writeResponse(response, plugin.ErrorEnvelope("invalid_request", "invalid request pointer or length"))
		return 1
	}
	var reqBytes []byte
	if requestLen != 0 {
		reqBytes = C.GoBytes(unsafe.Pointer(request), C.int(requestLen))
	}
	raw, ok := plugin.HandleMethod(C.GoString(method), reqBytes)
	if !writeResponse(response, raw) || !ok {
		return 1
	}
	return 0
}

// cliproxyPluginFree releases memory allocated by cliproxyPluginCall.
//
//export cliproxyPluginFree
func cliproxyPluginFree(ptr unsafe.Pointer, _ C.size_t) {
	defer func() { _ = recover() }()
	if ptr != nil {
		C.free(ptr)
	}
}

// cliproxyPluginShutdown is called by the host before unloading the plugin.
//
//export cliproxyPluginShutdown
func cliproxyPluginShutdown() {
	defer func() { _ = recover() }()
	plugin.Shutdown()
}

func writeResponse(response *C.cliproxy_buffer, raw []byte) bool {
	if response == nil || len(raw) == 0 {
		return false
	}
	ptr := C.CBytes(raw)
	if ptr == nil {
		return false
	}
	response.ptr = ptr
	response.len = C.size_t(len(raw))
	return true
}

// okEnvelope wraps a JSON value in the host envelope. Used by plugin.go tests.
func okEnvelope(result json.RawMessage) []byte {
	type envelope struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result,omitempty"`
	}
	raw, _ := json.Marshal(envelope{OK: true, Result: result})
	return raw
}
