#define _GNU_SOURCE
#include <dlfcn.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

typedef struct { void *ptr; size_t len; } buffer;
typedef struct { uint32_t abi_version; void *host_ctx; void *call; void *free_buffer; } host_api;
typedef int (*plugin_call_fn)(char *, uint8_t *, size_t, buffer *);
typedef void (*plugin_free_fn)(void *, size_t);
typedef void (*plugin_shutdown_fn)(void);
typedef struct { uint32_t abi_version; plugin_call_fn call; plugin_free_fn free_buffer; plugin_shutdown_fn shutdown; } plugin_api;
typedef int (*plugin_init_fn)(host_api *, plugin_api *);

static int call_and_check(plugin_api *p, const char *method, const char *json, const char *needle) {
    buffer out = {0};
    char *m = strdup(method);
    uint8_t *req = (uint8_t *)json;
    int rc = p->call(m, req, strlen(json), &out);
    free(m);
    if (!out.ptr) { fprintf(stderr, "CALL method=%s rc=%d response=<null>\n", method, rc); return 1; }
    printf("CALL method=%s rc=%d response=%.*s\n", method, rc, (int)out.len, (char *)out.ptr);
    int bad = (needle && !strstr((char *)out.ptr, needle));
    p->free_buffer(out.ptr, out.len);
    return bad;
}

int main(int argc, char **argv) {
    if (argc != 2) { fprintf(stderr, "usage: %s plugin.so\n", argv[0]); return 2; }
    void *h = dlopen(argv[1], RTLD_NOW | RTLD_LOCAL);
    if (!h) { fprintf(stderr, "dlopen: %s\n", dlerror()); return 3; }
    plugin_init_fn init = (plugin_init_fn)dlsym(h, "cliproxy_plugin_init");
    if (!init) { fprintf(stderr, "dlsym init: %s\n", dlerror()); dlclose(h); return 4; }
    plugin_api p = {0};
    int rc = init(NULL, &p);
    printf("INIT rc=%d abi_version=%u call=%p free=%p shutdown=%p\n", rc, p.abi_version, (void*)p.call, (void*)p.free_buffer, (void*)p.shutdown);
    if (rc != 0 || p.abi_version != 1 || !p.call || !p.free_buffer || !p.shutdown) { dlclose(h); return 5; }
    const char *register_json = "{\"schema_version\":6,\"config_yaml\":\"\"}";
    if (call_and_check(&p, "plugin.register", register_json, "\"ok\":true")) { p.shutdown(); dlclose(h); return 6; }
    if (call_and_check(&p, "management.register", "{}", "usage-audit/summary")) { p.shutdown(); dlclose(h); return 7; }
    const char *usage = "{\"Provider\":\"dummy\",\"Model\":\"smoke-model\",\"Alias\":\"dummy/smoke-model\",\"APIKey\":\"identifier-only\",\"SessionID\":\"smoke-session\",\"AuthIndex\":\"smoke-auth\",\"AuthType\":\"dummy\",\"Generate\":true,\"RequestedAt\":\"2025-01-01T00:00:00Z\",\"Latency\":1000,\"Detail\":{\"InputTokens\":3,\"OutputTokens\":2,\"TotalTokens\":5}}";
    if (call_and_check(&p, "usage.handle", usage, "\"ok\":true")) { p.shutdown(); dlclose(h); return 8; }
    usleep(200000);
    const char *mgmt = "{\"Method\":\"GET\",\"Path\":\"/v0/management/plugins/usage-audit/summary\",\"Query\":{}}";
    if (call_and_check(&p, "management.handle", mgmt, NULL)) { p.shutdown(); dlclose(h); return 9; }
    if (call_and_check(&p, "management.handle", "{\"Method\":\"GET\",\"Path\":\"/v0/management/plugins/usage-audit/records\",\"Query\":{}}", NULL)) { p.shutdown(); dlclose(h); return 10; }
    p.shutdown();
    printf("SHUTDOWN ok\n");
    if (dlclose(h) != 0) { fprintf(stderr, "dlclose: %s\n", dlerror()); return 11; }
    printf("NATIVE_SMOKE_PASS\n");
    return 0;
}
