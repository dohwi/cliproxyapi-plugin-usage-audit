#define _GNU_SOURCE
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "loader.h"

#ifdef _WIN32
#define strdup _strdup
#endif

typedef struct { void *ptr; size_t len; } buffer;
typedef struct { uint32_t abi_version; void *host_ctx; void *call; void *free_buffer; } host_api;
typedef int (*plugin_call_fn)(char *, uint8_t *, size_t, buffer *);
typedef void (*plugin_free_fn)(void *, size_t);
typedef void (*plugin_shutdown_fn)(void);
typedef struct { uint32_t abi_version; plugin_call_fn call; plugin_free_fn free_buffer; plugin_shutdown_fn shutdown; } plugin_api;
typedef int (*plugin_init_fn)(host_api *, plugin_api *);

/* ABI response buffers are length-delimited, never assumed NUL-terminated. */
static int contains(buffer out, const char *needle) {
    size_t n = strlen(needle);
    if (n == 0) return 1;
    if (!out.ptr || out.len < n) return 0;
    for (size_t i = 0; i <= out.len - n; i++) {
        if (memcmp((uint8_t *)out.ptr + i, needle, n) == 0) return 1;
    }
    return 0;
}

static int reject_invalid(plugin_api *p, char *method, uint8_t *request, size_t len, const char *needle) {
    buffer out = {0};
    int rc = p->call(method, request, len, &out);
    int bad = rc != 1 || !contains(out, needle);
    if (out.ptr) p->free_buffer(out.ptr, out.len);
    return bad;
}

static int call_and_check(plugin_api *p, const char *method, const char *json, const char *needle) {
    buffer out = {0};
    char *m = strdup(method);
    if (!m) return 1;
    uint8_t *req = (uint8_t *)json;
    int rc = p->call(m, req, strlen(json), &out);
    free(m);
    if (!out.ptr) { fprintf(stderr, "CALL method=%s rc=%d response=<null>\n", method, rc); return 1; }
    printf("CALL method=%s rc=%d response=%.*s\n", method, rc, (int)out.len, (char *)out.ptr);
    int bad = rc != 0 || (needle && !contains(out, needle));
    p->free_buffer(out.ptr, out.len);
    return bad;
}

int main(int argc, char **argv) {
    if (argc != 2) { fprintf(stderr, "usage: %s plugin-library\n", argv[0]); return 2; }
    lib_handle h = lib_open(argv[1]);
    if (!h) { fprintf(stderr, "load library: %s\n", lib_error()); return 3; }
    plugin_init_fn init = (plugin_init_fn)lib_sym(h, "cliproxy_plugin_init");
    if (!init) { fprintf(stderr, "missing init export: %s\n", lib_error()); lib_close(h); return 4; }
    plugin_api p = {0};
    host_api bad_host = {.abi_version = 99};
    if (init(NULL, NULL) == 0 || init(&bad_host, &p) == 0) { lib_close(h); return 5; }
    int rc = init(NULL, &p);
    printf("INIT rc=%d abi_version=%u call=%p free=%p shutdown=%p\n", rc, p.abi_version, (void*)p.call, (void*)p.free_buffer, (void*)p.shutdown);
    if (rc != 0 || p.abi_version != 1 || !p.call || !p.free_buffer || !p.shutdown) { lib_close(h); return 5; }
    if (p.call(NULL, NULL, 0, NULL) != 1 ||
        reject_invalid(&p, NULL, NULL, 0, "invalid_method") ||
        reject_invalid(&p, "plugin.register", NULL, (size_t)-1, "request_too_large") ||
        reject_invalid(&p, "plugin.register", (uint8_t *)"{}", (size_t)(8 << 20) + 1, "request_too_large") ||
        reject_invalid(&p, "plugin.register", NULL, 1, "invalid_request")) {
        p.shutdown(); lib_close(h); return 5;
    }
    const char *register_json = "{\"schema_version\":6,\"config_yaml\":\"\"}";
    if (call_and_check(&p, "plugin.register", register_json, "\"ok\":true")) { p.shutdown(); lib_close(h); return 6; }
    if (call_and_check(&p, "management.register", "{}", "usage-audit/summary")) { p.shutdown(); lib_close(h); return 7; }
    /* Dashboard data is served from authenticated management routes only. */
    if (call_and_check(&p, "management.register", "{}", "/v0/management/plugins/usage-audit/data/summary") ||
        call_and_check(&p, "management.register", "{}", "/v0/management/plugins/usage-audit/data/records")) { p.shutdown(); lib_close(h); return 12; }
    const char *usage = "{\"Provider\":\"dummy\",\"Model\":\"smoke-model\",\"Alias\":\"dummy/smoke-model\",\"APIKey\":\"identifier-only\",\"SessionID\":\"smoke-session\",\"AuthIndex\":\"smoke-auth\",\"AuthType\":\"dummy\",\"Generate\":true,\"RequestedAt\":\"2025-01-01T00:00:00Z\",\"Latency\":1000,\"Detail\":{\"InputTokens\":3,\"OutputTokens\":2,\"TotalTokens\":5}}";
    if (call_and_check(&p, "usage.handle", usage, "\"ok\":true")) { p.shutdown(); lib_close(h); return 8; }
    sleep_ms(200);
    const char *mgmt = "{\"Method\":\"GET\",\"Path\":\"/v0/management/plugins/usage-audit/summary\",\"Query\":{}}";
    if (call_and_check(&p, "management.handle", mgmt, NULL)) { p.shutdown(); lib_close(h); return 9; }
    if (call_and_check(&p, "management.handle", "{\"Method\":\"GET\",\"Path\":\"/v0/management/plugins/usage-audit/records\",\"Query\":{}}", NULL)) { p.shutdown(); lib_close(h); return 10; }
    if (call_and_check(&p, "management.handle", "{\"Method\":\"GET\",\"Path\":\"/v0/management/plugins/usage-audit/data/summary\",\"Query\":{}}", "\"StatusCode\":200") ||
        call_and_check(&p, "management.handle", "{\"Method\":\"GET\",\"Path\":\"/v0/management/plugins/usage-audit/data/records\",\"Query\":{}}", "\"StatusCode\":200")) { p.shutdown(); lib_close(h); return 13; }
    /* The unauthenticated resource prefix must not serve any usage data. */
    if (call_and_check(&p, "management.handle", "{\"Method\":\"GET\",\"Path\":\"/v0/resource/plugins/usage-audit/data/summary\",\"Query\":{}}", "\"StatusCode\":404") ||
        call_and_check(&p, "management.handle", "{\"Method\":\"GET\",\"Path\":\"/v0/resource/plugins/usage-audit/data/records\",\"Query\":{}}", "\"StatusCode\":404")) { p.shutdown(); lib_close(h); return 14; }
    p.shutdown();
    printf("SHUTDOWN ok\n");
    if (lib_close(h) != 0) { fprintf(stderr, "unload library: %s\n", lib_error()); return 11; }
    printf("NATIVE_SMOKE_PASS\n");
    return 0;
}
