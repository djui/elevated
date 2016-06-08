// See http://iquilezles.org/www/material/function2009/function2009.pdf
// See https://github.com/tomdalling/opengl-series/blob/master/source/01_project_skeleton/source/main.cpp
// See http://www.tomdalling.com/blog/modern-opengl/01-getting-started-in-xcode-and-visual-cpp/

package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const windowWidth = 800
const windowHeight = 600

type state struct {
	vao, vbo uint32
}

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("[E] Failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	w, err := glfw.CreateWindow(windowWidth, windowHeight, "Elevated", nil, nil)
	if err != nil {
		log.Fatalln("[E] Failed to create window:", err)
	}

	w.MakeContextCurrent()
	w.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		log.Fatalln("[E] Failed to initialize glow:", err)
	}

	log.Println("[I] OpenGL version:", gl.GoStr(gl.GetString(gl.VERSION)))
	log.Println("[I] GLSL version:", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	log.Println("[I] Vendor:", gl.GoStr(gl.GetString(gl.VENDOR)))
	log.Println("[I] Renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		log.Fatalln("[E] Failed to compile shader:", err)
	}
	defer gl.DeleteShader(vertexShader)

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatalln("[E] Failed to compile shader:", err)
	}
	defer gl.DeleteShader(fragmentShader)

	p, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatalln("[E] Failed to create program:", err)
	}

	s := &state{}

	if err := LoadTriangle(p, s); err != nil {
		log.Fatalln("[E] Failed to load triangle:", err)
	}

	setupScene()

	for !w.ShouldClose() {
		drawScene(w, p, s)

		glfw.PollEvents()
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
	if key == glfw.KeyEscape {
		window.SetShouldClose(true)
	}
}

func setupScene() {
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	//gl.ClearColor(1.0, 1.0, 1.0, 1.0)
	//gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.ClearColor(0.5, 0.5, 0.5, 1.0)
}

func drawScene(window *glfw.Window, p program, s *state) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(uint32(p))

	gl.BindVertexArray(s.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(0)

	gl.UseProgram(0)

	window.SwapBuffers()
}

func LoadTriangle(p program, s *state) error {
	vertexData := []float64{
		//  X     Y     Z
		0.0, 0.8, 0.0,
		-0.8, -0.8, 0.0,
		0.8, -0.8, 0.0,
	}

	gl.GenVertexArrays(1, &s.vao)
	gl.BindVertexArray(s.vao)

	gl.GenBuffers(1, &s.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexData)*4, gl.Ptr(vertexData), gl.STATIC_DRAW)

	vertAttrib, err := p.attr("vert")
	if err != nil {
		return err
	}
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return nil
}

type program uint32

func (p program) attr(s string) (uint32, error) {
	a := gl.GetAttribLocation(uint32(p), glStr(s))
	if a == -1 {
		return 0, fmt.Errorf("program attribute not found: %v", s)
	}

	return uint32(a), nil
}

func newProgram(shaders ...uint32) (program, error) {
	if len(shaders) == 0 {
		return 0, fmt.Errorf("no shaders provided")
	}

	p := gl.CreateProgram()

	for _, shader := range shaders {
		gl.AttachShader(p, shader)
	}

	gl.LinkProgram(p)

	for _, shader := range shaders {
		gl.DetachShader(p, shader)
	}

	var status int32
	gl.GetProgramiv(p, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(p, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(p, logLength, nil, gl.Str(log))

		gl.DeleteProgram(p)

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	return program(p), nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		gl.DeleteShader(shader)

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func glStr(s string) *uint8 {
	return gl.Str(s + "\x00")
}

func shaderResource(s string) string {
	return "#version 330\n\n" + s + "\x00"
}

var vertexShaderSource = shaderResource(`
in vec3 vert;

void main() {
    gl_Position = vec4(vert, 1);
}
`)

var fragmentShaderSource = shaderResource(`
out vec4 outputColor;

void main() {
    outputColor = vec4(1.0, 1.0, 1.0, 1.0);
}
`)
