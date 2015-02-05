package main

import (
	"./dlist"
	"./mymath"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sort"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
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
func make_program(vert_file_name, frag_file_name string) gl.Program {
	vert_source, err := ioutil.ReadFile(vert_file_name)
	if err != nil {
		panic(err)
	}
	frag_source, err := ioutil.ReadFile(frag_file_name)
	if err != nil {
		panic(err)
	}
	vs := glh.Shader{gl.VERTEX_SHADER, string(vert_source)}
	fs := glh.Shader{gl.FRAGMENT_SHADER, string(frag_source)}
	return glh.NewProgram(vs, fs)
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
	gl.BufferData(gl.ARRAY_BUFFER, len(vertex_buffer_data)*4, vertex_buffer_data, gl.STATIC_DRAW)
	gl.DrawArrays(gl.LINE_STRIP, 0, len(vertex_buffer_data)/2)
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
	gl.BufferData(gl.ARRAY_BUFFER, len(vertex_buffer_data)*4, vertex_buffer_data, gl.STATIC_DRAW)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, len(vertex_buffer_data)/2)
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
	window, err := glfw.CreateWindow(width, height, "PCB Viewer", nil, nil)
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
	vertex_array := gl.GenVertexArray()
	vertex_array.Bind()

	//load shaders and get address of shader variables
	prog := make_program("shaders/VertexShader.vert", "shaders/FragmentShader.frag")
	vert_color_id := prog.GetUniformLocation("vert_color")
	vert_scale_id := prog.GetUniformLocation("vert_scale")
	vert_offset_id := prog.GetUniformLocation("vert_offset")

	//use the loaded shader program
	prog.Use()

	//set aspect and offset for 2D drawing
	vert_scale_id.Uniform2f(2.0/float32(width), -2.0/float32(height))
	vert_offset_id.Uniform2f(-1.0, 1.0)

	//setup vertex buffer ready for use
	vertex_buffer := gl.GenBuffer()
	vertex_attrib := gl.AttribLocation(0)
	vertex_buffer.Bind(gl.ARRAY_BUFFER)
	vertex_attrib.EnableArray()
	vertex_attrib.AttribPointer(2, gl.FLOAT, false, 0, nil)

	//create display list
	dlist := dlist.Newdlist(width, height, 10)

	//add stroke path and strip
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

	//add bezier path and strip
	bez_path_id := dlist.Create_path()
	dlist.Add_bezier(bez_path_id,
		&mymath.Point{900.0, 0.0},
		&mymath.Point{0.0, 700.0},
		&mymath.Point{900.0, 700.0},
		1.0)
	bez_strip_id := dlist.Create_path_strip(bez_path_id, 15, 3, 1, 16)

	//add circle strip
	circle_path_id := dlist.Create_path()
	dlist.Add_abs_path(circle_path_id, mymath.Circle_as_lines(&mymath.Point{0.0, 0.0}, 75.0, 64))
	circle_strip_id := dlist.Create_circle_strip(&mymath.Point{0.0, 0.0}, 100.0, 50.0, 64)

	//add instances to shape map
	shape_map := map[int]*shape{}
	shape_map[0] = &shape{&mymath.Point{25.0, 25.0}, 1.0, 1.0, 1.0, 1.0, bez_strip_id, bez_path_id, 15, 0}
	shape_map[1] = &shape{&mymath.Point{200.0, 300.0}, 1.0, 0.0, 0.0, 1.0, circle_strip_id, circle_path_id, 25, 0}
	shape_map[2] = &shape{&mymath.Point{250.0, 550.0}, 1.0, 1.0, 0.0, 1.0, circle_strip_id, circle_path_id, 25, 0}
	shape_map[3] = &shape{&mymath.Point{600.0, 300.0}, 1.0, 0.0, 1.0, 1.0, stroke_strip_id, stroke_path_id, 10, 0}
	shape_map[4] = &shape{&mymath.Point{600.0, 500.0}, 0.0, 1.0, 1.0, 1.0, stroke_strip_id, stroke_path_id, 10, 0}
	shape_map[5] = &shape{&mymath.Point{800.0, 100.0}, 0.0, 0.0, 1.0, 1.0, circle_strip_id, circle_path_id, 25, 0}

	//add shapes to spacial cache
	for id, shape := range shape_map {
		dlist.Add_collision_path(shape.offset, shape.path_id, shape.radius, shape.gap, id)
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
			vert_color_id.Uniform4f(shape.red, shape.green, shape.blue, shape.alpha)
			draw_filled_polygon(shape.offset, dlist.Get_strip(shape.strip_id))
			vert_color_id.Uniform4f(0.0, 0.0, 0.0, 1.0)
			draw_polygon(shape.offset, dlist.Get_path(shape.path_id))
		}

		//show window just drawn
		window.SwapBuffers()
	}

	//clean up
	vertex_buffer.Delete()
	vertex_array.Delete()
	prog.Delete()
	window.Destroy()
	glfw.Terminate()
}
