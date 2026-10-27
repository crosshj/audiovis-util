// Struct-by-value callback adapter for the pinned Ultralight C API.
// Built on the development machine; no renderer/window implementation here.
#include <Ultralight/CAPI.h>
#include <stddef.h>

static void (*render_buffer_cb)(unsigned, const ULRenderBuffer*);
static void (*create_geometry_cb)(unsigned, const ULVertexBuffer*, const ULIndexBuffer*);
static void (*update_geometry_cb)(unsigned, const ULVertexBuffer*, const ULIndexBuffer*);
static void (*commands_cb)(const ULCommandList*);
static void render_buffer(unsigned id, ULRenderBuffer b) { render_buffer_cb(id, &b); }
static void create_geometry(unsigned id, ULVertexBuffer v, ULIndexBuffer i) { create_geometry_cb(id, &v, &i); }
static void update_geometry(unsigned id, ULVertexBuffer v, ULIndexBuffer i) { update_geometry_cb(id, &v, &i); }
static void commands(ULCommandList list) { commands_cb(&list); }

void bridge_install(void (*setter)(ULGPUDriver), const ULGPUDriver* callbacks,
    void* rb, void* cg, void* ug, void* cl) {
  render_buffer_cb = rb; create_geometry_cb = cg; update_geometry_cb = ug; commands_cb = cl;
  ULGPUDriver driver = *callbacks;
  driver.create_render_buffer = render_buffer;
  driver.create_geometry = create_geometry;
  driver.update_geometry = update_geometry;
  driver.update_command_list = commands;
  setter(driver);
}
void bridge_projection(ULMatrix4x4 (*apply)(ULMatrix4x4,float,float,bool),
    const ULMatrix4x4* input, float w, float h, bool flip, ULMatrix4x4* out) {
  *out = apply(*input,w,h,flip);
}
unsigned bridge_layout(unsigned i) {
  const unsigned values[] = {
    sizeof(ULGPUState), sizeof(ULCommand), sizeof(ULRenderTarget),
    offsetof(ULGPUState,render_buffer_id), offsetof(ULGPUState,uniform_scalar),
    offsetof(ULGPUState,clip), offsetof(ULGPUState,scissor_rect),
    offsetof(ULCommand,gpu_state), offsetof(ULCommand,geometry_id),
    sizeof(ULRenderBuffer), sizeof(ULVertexBuffer), sizeof(ULIndexBuffer)
  };
  return i < sizeof(values)/sizeof(values[0]) ? values[i] : 0;
}
