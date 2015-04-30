package main

import (
	"errors"
	"./dlist"
	"./mymath"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	glfw "github.com/go-gl/glfw/v3.0/glfw"
)

const (
	width  = 1024
	height = 768
)

type shape struct {
	offset   *mymath.Point
	red      float32
	blue     float32
	green    float32
	alpha    float32
	strip_id int
	path_id  int
	radius   float32
	gap      float32
}

//load shader progs
func make_program(vert_file_name, frag_file_name string) uint32 {
	vert_source, err := ioutil.ReadFile(vert_file_name)
	if err != nil {
		panic(err)
	}
	frag_source, err := ioutil.ReadFile(frag_file_name)
	if err != nil {
		panic(err)
	}
	program, err := newProgram(string(vert_source) + "\x00", string(frag_source) + "\x00")
	return program
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, errors.New(fmt.Sprintf("failed to link program: %v", log))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csource := gl.Str(source)
	gl.ShaderSource(shader, 1, &csource, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

//draw a line strip polygon
func draw_polygon(offsetp *mymath.Point, datap *mymath.Points) {
	data, offset := *datap, *offsetp
	vertex_buffer_data := make([]float32, len(data)*2, len(data)*2)
	for i := 0; i < len(data); i++ {
		pp := data[i]
		p := *pp
		vertex_buffer_data[i*2] = p[0] + offset[0]
		vertex_buffer_data[i*2+1] = p[1] + offset[1]
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertex_buffer_data)*4, gl.Ptr(vertex_buffer_data), gl.STATIC_DRAW)
	gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(vertex_buffer_data)/2))
}

//draw a triangle strip polygon
func draw_filled_polygon(offsetp *mymath.Point, datap *mymath.Points) {
	data, offset := *datap, *offsetp
	vertex_buffer_data := make([]float32, len(data)*2, len(data)*2)
	for i := 0; i < len(data); i++ {
		pp := data[i]
		p := *pp
		vertex_buffer_data[i*2] = p[0] + offset[0]
		vertex_buffer_data[i*2+1] = p[1] + offset[1]
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertex_buffer_data)*4, gl.Ptr(vertex_buffer_data), gl.STATIC_DRAW)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(vertex_buffer_data)/2))
}

func main() {
	runtime.LockOSThread()

	//create window
	if !glfw.Init() {
		fmt.Fprintf(os.Stderr, "Can't open GLFW")
		return
	}
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True) // needed for macs
	window, err := glfw.CreateWindow(width, height, "GL Paths", nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	window.MakeContextCurrent()
	window.SetInputMode(glfw.StickyKeys, 1)
	window.SetInputMode(glfw.StickyMouseButtons, 1)

	//set gl settings
	gl.Init()
	gl.GetError()
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.LineWidth(1.0)

	//create vertex array
	var vertex_array uint32
	gl.GenVertexArrays(1, &vertex_array)
	gl.BindVertexArray(vertex_array)

	//load shaders and get address of shader variables
	prog := make_program("shaders/VertexShader.vert", "shaders/FragmentShader.frag")
	vert_color_id := gl.GetUniformLocation(prog, gl.Str("vert_color\x00"))
	vert_scale_id := gl.GetUniformLocation(prog, gl.Str("vert_scale\x00"))
	vert_offset_id := gl.GetUniformLocation(prog, gl.Str("vert_offset\x00"))

	//use the loaded shader program
	gl.UseProgram(prog)

	//set aspect and offset for 2D drawing
	gl.Uniform2f(vert_scale_id, 2.0/float32(width), -2.0/float32(height))
	gl.Uniform2f(vert_offset_id, -1.0, 1.0)

	//setup vertex buffer ready for use
	var vertex_buffer uint32
	gl.GenBuffers(1, &vertex_buffer)
	vertex_attrib := uint32(gl.GetAttribLocation(prog, gl.Str("vert_vertex\x00")))
	gl.BindBuffer(gl.ARRAY_BUFFER, vertex_buffer)
	gl.EnableVertexAttribArray(vertex_attrib)
	gl.VertexAttribPointer(vertex_attrib, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	//create display list
	dlist := dlist.Newdlist(width, height, 10)

	//create stroke path and strip
	stroke_path_id := dlist.Create_path()
	dlist.Add_rel_path(stroke_path_id, &mymath.Points{
		&mymath.Point{0.0, 100.0},
		&mymath.Point{50.0, 0.0},
		&mymath.Point{0.0, -50.0},
		&mymath.Point{-25.0, 0.0}})
	dlist.Add_bezier(stroke_path_id,
		&mymath.Point{0.0, -100.0},
		&mymath.Point{100.0, 0.0},
		&mymath.Point{100.0, -100.0},
		1.0)
	dlist.Add_rel_path(stroke_path_id, &mymath.Points{
		&mymath.Point{50.0, 0.0},
		&mymath.Point{0.0, 100.0},
		&mymath.Point{50.0, 0.0},
		&mymath.Point{0.0, -50.0},
		&mymath.Point{-25.0, 0.0}})
	dlist.Add_bezier(stroke_path_id,
		&mymath.Point{0.0, -100.0},
		&mymath.Point{100.0, 0.0},
		&mymath.Point{100.0, -100.0},
		1.0)
	stroke_strip_id := dlist.Create_path_strip(stroke_path_id, 10, 3, 2, 16)

	//create bezier path and strip
	bez_path_id := dlist.Create_path()
	dlist.Add_bezier(bez_path_id,
		&mymath.Point{500.0, 0.0},
		&mymath.Point{0.0, 500.0},
		&mymath.Point{500.0, 500.0},
		1.0)
	bez_strip_id := dlist.Create_path_strip(bez_path_id, 15, 3, 1, 16)

	//create circle path and strip
	circle_path_id := dlist.Create_path()
	dlist.Add_abs_path(circle_path_id, mymath.Circle_as_lines(&mymath.Point{0.0, 0.0}, 75.0, 64))
	circle_strip_id := dlist.Create_circle_strip(&mymath.Point{0.0, 0.0}, 100.0, 50.0, 64)

	//add instances to shape map
	shape_map := map[int]*shape{}
	shape_map[0] = &shape{&mymath.Point{25.0, 25.0}, 1.0, 1.0, 1.0, 1.0, bez_strip_id, bez_path_id, 15, 0}
	shape_map[1] = &shape{&mymath.Point{200.0, 300.0}, 1.0, 0.0, 0.0, 1.0, circle_strip_id, circle_path_id, 25, 0}
	shape_map[2] = &shape{&mymath.Point{250.0, 550.0}, 0.0, 1.0, 0.0, 1.0, circle_strip_id, circle_path_id, 25, 0}
	shape_map[3] = &shape{&mymath.Point{600.0, 300.0}, 0.0, 0.0, 1.0, 1.0, stroke_strip_id, stroke_path_id, 10, 0}
	shape_map[4] = &shape{&mymath.Point{600.0, 500.0}, 1.0, 1.0, 0.0, 1.0, stroke_strip_id, stroke_path_id, 10, 0}
	shape_map[5] = &shape{&mymath.Point{800.0, 100.0}, 0.0, 1.0, 1.0, 0.75, circle_strip_id, circle_path_id, 25, 0}
	shape_map[6] = &shape{&mymath.Point{350.0, 250.0}, 1.0, 0.0, 1.0, 0.75, bez_strip_id, bez_path_id, 15, 0}

	//add shape paths to spacial cache
	for shape_id, shape := range shape_map {
		dlist.Add_collision_path(shape.offset, shape.path_id, shape.radius, shape.gap, shape_id)
	}

	mouse_shape_id := 0
	drag_offset_x := float32(0.0)
	drag_offset_y := float32(0.0)

	for {
		//exit of ESC key or close button pressed
		glfw.PollEvents()
		if (window.GetKey(glfw.KeyEscape) == glfw.Press) || window.ShouldClose() {
			break
		}

		//check for mouse down and collide with object
		xpos, ypos := window.GetCursorPosition()
		if window.GetMouseButton(glfw.MouseButton1) == glfw.Press {
			if mouse_shape_id == -1 {
				mouse_shape_id = dlist.Hit_collision_path(&mymath.Point{float32(xpos), float32(ypos)})
				if mouse_shape_id != -1 {
					shape := shape_map[mouse_shape_id]
					offset := *shape.offset
					drag_offset_x = float32(xpos) - offset[0]
					drag_offset_y = float32(ypos) - offset[1]
				}
			}
			if mouse_shape_id != -1 {
				shape := shape_map[mouse_shape_id]
				dlist.Sub_collision_path(shape.offset, shape.path_id, shape.radius, shape.gap, mouse_shape_id)
				shape.offset = &mymath.Point{float32(xpos) - drag_offset_x, float32(ypos) - drag_offset_y}
				dlist.Add_collision_path(shape.offset, shape.path_id, shape.radius, shape.gap, mouse_shape_id)
			}
		} else {
			mouse_shape_id = -1
		}

		//clear background
		gl.Clear(gl.COLOR_BUFFER_BIT)

		//draw shapes in id order !
		keys := make([]int, 0, len(shape_map))
		for id := range shape_map {
			keys = append(keys, id)
		}
		sort.Ints(keys)
		for _, id := range keys {
			shape := shape_map[id]
			gl.Uniform4f(vert_color_id, shape.red, shape.green, shape.blue, shape.alpha)
			draw_filled_polygon(shape.offset, dlist.Get_strip(shape.strip_id))
			gl.Uniform4f(vert_color_id, 0.0, 0.0, 0.0, 1.0)
			draw_polygon(shape.offset, dlist.Get_path(shape.path_id))
		}

		//show window just drawn
		window.SwapBuffers()
	}

	//clean up
	gl.DeleteBuffers(1, &vertex_buffer)
	gl.DeleteVertexArrays(1, &vertex_array)
	gl.DeleteProgram(prog)
	window.Destroy()
	glfw.Terminate()
}
