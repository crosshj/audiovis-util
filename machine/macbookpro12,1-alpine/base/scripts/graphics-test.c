/*
 * Small direct-display smoke test for the MacBookPro12,1 base.
 * Build on Alpine with: cc -O2 graphics-test.c -o sdl3-kmsdrm-smoke \
 *   $(pkg-config --cflags --libs sdl3) -lGL
 */
#define GL_GLEXT_PROTOTYPES 1
#include <SDL3/SDL.h>
#include <SDL3/SDL_opengl.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static GLuint compile_shader(GLenum type, const char *source) {
	GLuint shader = glCreateShader(type);
	glShaderSource(shader, 1, &source, NULL);
	glCompileShader(shader);
	GLint ok = 0;
	glGetShaderiv(shader, GL_COMPILE_STATUS, &ok);
	if (!ok) {
		char log[2048] = {0};
		glGetShaderInfoLog(shader, sizeof log, NULL, log);
		fprintf(stderr, "Shader compile failed: %s\n", log);
		exit(2);
	}
	return shader;
}

int main(void) {
	SDL_SetHint(SDL_HINT_VIDEO_DRIVER, "kmsdrm");
	SDL_SetHint(SDL_HINT_KMSDRM_REQUIRE_DRM_MASTER, "1");
	/* Intel's legacy pageflip path works more reliably on this 2015 iGPU. */
	SDL_SetHint(SDL_HINT_KMSDRM_ATOMIC, "0");
	if (!SDL_Init(SDL_INIT_VIDEO)) {
		fprintf(stderr, "SDL_Init failed: %s\n", SDL_GetError());
		return 1;
	}
	SDL_GL_SetAttribute(SDL_GL_CONTEXT_MAJOR_VERSION, 3);
	SDL_GL_SetAttribute(SDL_GL_CONTEXT_MINOR_VERSION, 2);
	SDL_GL_SetAttribute(SDL_GL_CONTEXT_PROFILE_MASK, SDL_GL_CONTEXT_PROFILE_CORE);
	SDL_GL_SetAttribute(SDL_GL_DOUBLEBUFFER, 1);
	SDL_Window *window = SDL_CreateWindow("Audio Utility graphics test", 0, 0,
		SDL_WINDOW_OPENGL | SDL_WINDOW_FULLSCREEN);
	if (!window) {
		fprintf(stderr, "SDL_CreateWindow failed: %s\n", SDL_GetError());
		return 1;
	}
	if (!SDL_ShowWindow(window) || !SDL_SetWindowFullscreen(window, true)) {
		fprintf(stderr, "Could not show fullscreen KMSDRM window: %s\n", SDL_GetError());
		return 1;
	}
	SDL_GLContext context = SDL_GL_CreateContext(window);
	if (!context) {
		fprintf(stderr, "SDL_GL_CreateContext failed: %s\n", SDL_GetError());
		return 1;
	}

	printf("SDL video driver: %s\n", SDL_GetCurrentVideoDriver());
	printf("GL_VENDOR: %s\n", glGetString(GL_VENDOR));
	printf("GL_RENDERER: %s\n", glGetString(GL_RENDERER));
	printf("GL_VERSION: %s\n", glGetString(GL_VERSION));
	printf("GLSL_VERSION: %s\n", glGetString(GL_SHADING_LANGUAGE_VERSION));
	fflush(stdout);

	const char *vs = "#version 150 core\n"
		"out vec3 color;\n"
		"void main(){ vec2 p[3]=vec2[3](vec2(0.0,0.72),vec2(-0.72,-0.55),vec2(0.72,-0.55));"
		" gl_Position=vec4(p[gl_VertexID],0.0,1.0); color=vec3(0.95,0.80,0.20); }\n";
	const char *fs = "#version 150 core\n"
		"in vec3 color; out vec4 outColor; void main(){ outColor=vec4(color,1.0); }\n";
	GLuint program = glCreateProgram();
	glAttachShader(program, compile_shader(GL_VERTEX_SHADER, vs));
	glAttachShader(program, compile_shader(GL_FRAGMENT_SHADER, fs));
	glLinkProgram(program);
	GLint linked = 0;
	glGetProgramiv(program, GL_LINK_STATUS, &linked);
	if (!linked) {
		char log[2048] = {0};
		glGetProgramInfoLog(program, sizeof log, NULL, log);
		fprintf(stderr, "Program link failed: %s\n", log);
		return 2;
	}

	GLuint vao = 0;
	glGenVertexArrays(1, &vao);
	glBindVertexArray(vao);
	int width = 0, height = 0;
	SDL_GetWindowSizeInPixels(window, &width, &height);
	glViewport(0, 0, width, height);
	glClearColor(0.02f, 0.18f, 0.72f, 1.0f);
	glUseProgram(program);
	if (!SDL_GL_SetSwapInterval(1)) {
		fprintf(stderr, "Could not enable vsync: %s\n", SDL_GetError());
	}
	glClear(GL_COLOR_BUFFER_BIT);
	glDrawArrays(GL_TRIANGLES, 0, 3);
	glFinish();
	GLubyte center_pixel[4] = {0};
	glReadBuffer(GL_BACK);
	glReadPixels(width / 2, height / 2, 1, 1, GL_RGBA, GL_UNSIGNED_BYTE, center_pixel);
	GLenum error = glGetError();
	if (error == GL_NO_ERROR && !(center_pixel[0] > 180 && center_pixel[1] > 140 && center_pixel[2] < 100)) {
		fprintf(stderr, "Shader triangle pixel is missing at screen center (RGBA %u,%u,%u,%u).\n",
			center_pixel[0], center_pixel[1], center_pixel[2], center_pixel[3]);
		return 2;
	}
	printf("GLSL shader output checked: center RGBA %u,%u,%u,%u.\n",
		center_pixel[0], center_pixel[1], center_pixel[2], center_pixel[3]);
	printf("KMSDRM fullscreen window: %dx%d.\n", width, height);
	printf("The blue background and yellow triangle stay up for 30 seconds, then return to the original console.\n");
	fflush(stdout);

	time_t deadline = time(NULL) + 30;
	bool first_frame = true;
	while (time(NULL) < deadline) {
		glViewport(0, 0, width, height);
		glClear(GL_COLOR_BUFFER_BIT);
		glDrawArrays(GL_TRIANGLES, 0, 3);
		glFinish();
		if (glGetError() != GL_NO_ERROR) {
			fprintf(stderr, "OpenGL frame rendering failed.\n");
			return 2;
		}
		if (!SDL_GL_SwapWindow(window)) {
			fprintf(stderr, "SDL_GL_SwapWindow failed: %s\n", SDL_GetError());
			return 2;
		}
		if (first_frame) {
			printf("PASS: SDL KMSDRM pageflip is continuously presenting the shader output.\n");
			fflush(stdout);
			first_frame = false;
		}
		SDL_PumpEvents();
		SDL_Delay(16);
	}
	SDL_GL_DestroyContext(context);
	SDL_DestroyWindow(window);
	SDL_Quit();
	return 0;
}
