//package name
package dlist

//package imports
import (
	"../layer"
	"../mymath"
)

////////////////////////
//public structure/types
////////////////////////

/////////////////////////
//private structure/types
/////////////////////////

//////////////
//dlist object
//////////////

type Dlist struct {
	layer  *layer.Layer
	width  int
	height int
	scale  int
	paths  []*mymath.Points
	strips []*mymath.Points
}

////////////////
//public methods
////////////////

func Newdlist(width, height, scale int) *Dlist {
	d := Dlist{}
	d.init(width, height, scale)
	return &d
}

func (self *Dlist) Get_path(id int) *mymath.Points {
	return self.paths[id]
}

func (self *Dlist) Get_strip(id int) *mymath.Points {
	return self.strips[id]
}

func (self *Dlist) Create_path() int {
	self.paths = append(self.paths, &mymath.Points{})
	return len(self.paths) - 1
}

func (self *Dlist) Add_rel_path(id int, points *mymath.Points) {
	rx := float32(0.0)
	ry := float32(0.0)
	ex := float32(0.0)
	ey := float32(0.0)
	path := *self.paths[id]
	if len(path) != 0 {
		epp := path[len(path)-1]
		ep := *epp
		ex = ep[0]
		ey = ep[1]
	} else {
		path = append(path, &mymath.Point{0.0, 0.0})
	}
	for _, pp := range *points {
		p := *pp
		rx += p[0]
		ry += p[1]
		p[0] = rx + ex
		p[1] = ry + ey
		path = append(path, pp)
	}
	self.paths[id] = &path
}

func (self *Dlist) Add_abs_path(id int, pointsp *mymath.Points) {
	ex := float32(0.0)
	ey := float32(0.0)
	start := 0
	path := *self.paths[id]
	if len(path) != 0 {
		epp := path[len(path)-1]
		ep := *epp
		ex = ep[0]
		ey = ep[1]
		start = 1
	}
	points := *pointsp
	for i := start; i < len(points); i++ {
		pp := points[i]
		p := *pp
		p[0] += ex
		p[1] += ey
		path = append(path, pp)
	}
	self.paths[id] = &path
}

func (self *Dlist) Add_bezier(id int, p2, p3, p4 *mymath.Point, dist float32) {
	ex := float32(0.0)
	ey := float32(0.0)
	start := 0
	path := *self.paths[id]
	if len(path) != 0 {
		epp := path[len(path)-1]
		ep := *epp
		ex = ep[0]
		ey = ep[1]
		start = 1
	}
	points := *mymath.Bezier_path_as_lines(&mymath.Point{0.0, 0.0}, p2, p3, p4, dist)
	for i := start; i < len(points); i++ {
		pp := points[i]
		p := *pp
		p[0] += ex
		p[1] += ey
		path = append(path, pp)
	}
	self.paths[id] = &path
}

func (self *Dlist) Create_path_strip(id int, radius float32, capstyle, joinstyle, resolution int) int {
	points := mymath.Thicken_path_as_tristrip(self.paths[id], radius, capstyle, joinstyle, resolution)
	self.strips = append(self.strips, points)
	return len(self.strips) - 1
}

func (self *Dlist) Create_circle_strip(center *mymath.Point, radius1, radius2 float32, resolution int) int {
	points := mymath.Circle_as_tristrip(center, radius1, radius2, resolution)
	self.strips = append(self.strips, points)
	return len(self.strips) - 1
}

func (self *Dlist) Add_collision_path(offset *mymath.Point, path_id int, radius, gap float32, id int) {
	self.layer.Add_path(offset, self.paths[path_id], radius, gap, id)
}

func (self *Dlist) Sub_collision_path(offset *mymath.Point, path_id int, radius, gap float32, id int) {
	self.layer.Sub_path(offset, self.paths[path_id], radius, gap, id)
}

func (self *Dlist) Hit_collision_path(offsetp *mymath.Point) int {
	offset := *offsetp
	x := offset[0]
	y := offset[1]
	l := layer.Point{x, y}
	line := layer.Line{&l, &l, 0.01, 0.0}
	return self.layer.Hit_Line(&line)
}

/////////////////
//private methods
/////////////////

func (self *Dlist) init(width, height, scale int) {
	self.paths = nil
	self.width = width
	self.height = height
	self.scale = scale
	cols := width/scale
	rows := height/scale
	self.layer = layer.Newlayer(cols + 1, rows + 1, 1.0/(float32(width)/float32(cols)), 1.0/(float32(height)/float32(rows)))
	return
}
