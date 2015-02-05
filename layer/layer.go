//package name
package layer

//package imports
import (
	"../mymath"
	"math"
)

////////////////////////
//public structure/types
////////////////////////

type Point struct {
	X float32
	Y float32
}

type Line struct {
	P1     *Point
	P2     *Point
	Radius float32
	Gap    float32
}

/////////////////////////
//private structure/types
/////////////////////////

type record struct {
	count int
	id    int
	line  *Line
}

type aabb struct {
	minx int
	miny int
	maxx int
	maxy int
}

type bucket []*record
type buckets []bucket

//////////////
//Layer object
//////////////

type Layer struct {
	width   int
	height  int
	scalex  float32
	scaley  float32
	buckets buckets
	count   int
}

////////////////
//public methods
////////////////

func Newlayer(width, height int, sx, sy float32) *Layer {
	l := Layer{}
	l.init(width, height, sx, sy)
	return &l
}

func (self *Layer) Add_Line(l *Line, id int) {
	new_record := record{0, id, l}
	bb := self.aabb(l)
	for y := bb.miny; y < bb.maxy; y++ {
		for x := bb.minx; x < bb.maxx; x++ {
			b := y*self.width + x
			self.buckets[b] = append(self.buckets[b], &new_record)
		}
	}
}

func (self *Layer) Sub_Line(l *Line, id int) {
	bb := self.aabb(l)
	for y := bb.miny; y < bb.maxy; y++ {
		for x := bb.minx; x < bb.maxx; x++ {
			b := y*self.width + x
			for i := len(self.buckets[b]) - 1; i >= 0; i-- {
				record := self.buckets[b][i]
				if record.id == id {
					if lines_equal(record.line, l) {
						self.buckets[b] = append(self.buckets[b][:i], self.buckets[b][i+1:]...)
						break
					}
				}
			}
		}
	}
}

func (self *Layer) Hit_Line(l *Line) int {
	self.count += 1
	bb := self.aabb(l)
	l1_p1 := mymath.Point{l.P1.X, l.P1.Y}
	l1_p2 := mymath.Point{l.P2.X, l.P2.Y}
	l2_p1 := mymath.Point{0.0, 0.0}
	l2_p2 := mymath.Point{0.0, 0.0}
	for y := bb.miny; y < bb.maxy; y++ {
		for x := bb.minx; x < bb.maxx; x++ {
			for _, record := range self.buckets[y*self.width+x] {
				if record.count != self.count {
					record.count = self.count
					r := l.Radius + record.line.Radius
					if l.Gap >= record.line.Gap {
						r += l.Gap
					} else {
						r += record.line.Gap
					}
					l2_p1[0], l2_p1[1] = record.line.P1.X, record.line.P1.Y
					l2_p2[0], l2_p2[1] = record.line.P2.X, record.line.P2.Y
					if mymath.Collide_thick_lines_2d(&l1_p1, &l1_p2, &l2_p1, &l2_p2, r) {
						return record.id
					}
				}
			}
		}
	}
	return -1
}

func (self *Layer) Add_path(offsetp *mymath.Point, pathp *mymath.Points, radius, gap float32, id int) {
	path, offset := *pathp, *offsetp
	pp1 := path[0]
	p1 := *pp1
	lp1 := &Point{p1[0] + offset[0], p1[1] + offset[1]}
	for i := 1; i < len(path); i++ {
		lp0 := lp1
		pp1 = path[i]
		p1 = *pp1
		lp1 = &Point{p1[0] + offset[0], p1[1] + offset[1]}
		self.Add_Line(&Line{lp0, lp1, radius, gap}, id)
	}
}

func (self *Layer) Sub_path(offsetp *mymath.Point, pathp *mymath.Points, radius, gap float32, id int) {
	path, offset := *pathp, *offsetp
	pp1 := path[0]
	p1 := *pp1
	lp1 := &Point{p1[0] + offset[0], p1[1] + offset[1]}
	for i := 1; i < len(path); i++ {
		lp0 := lp1
		pp1 = path[i]
		p1 = *pp1
		lp1 = &Point{p1[0] + offset[0], p1[1] + offset[1]}
		self.Sub_Line(&Line{lp0, lp1, radius, gap}, id)
	}
}

/////////////////
//private methods
/////////////////

func (self *Layer) init(width, height int, sx, sy float32) {
	self.width = width
	self.height = height
	self.scalex = sx
	self.scaley = sy
	self.buckets = make(buckets, (width * height), (width * height))
	for i := 0; i < (width * height); i++ {
		self.buckets[i] = bucket{}
	}
	self.count = 0
	return
}

func (self *Layer) aabb(l *Line) *aabb {
	x1, y1, x2, y2 := l.P1.X, l.P1.Y, l.P2.X, l.P2.Y
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	r := l.Radius + l.Gap
	minx := int(math.Floor(float64((x1 - r) * self.scalex)))
	miny := int(math.Floor(float64((y1 - r) * self.scaley)))
	maxx := int(math.Ceil(float64((x2 + r) * self.scalex)))
	maxy := int(math.Ceil(float64((y2 + r) * self.scaley)))
	if minx < 0 {
		minx = 0
	}
	if miny < 0 {
		miny = 0
	}
	if maxx > self.width {
		maxx = self.width
	}
	if maxy > self.height {
		maxy = self.height
	}
	return &aabb{minx, miny, maxx, maxy}
}

func lines_equal(l1, l2 *Line) bool {
	if l1 == l2 {
		return true
	}
	if l1.P1.X != l2.P1.X {
		return false
	}
	if l1.P1.Y != l2.P1.Y {
		return false
	}
	if l1.P2.X != l2.P2.X {
		return false
	}
	if l1.P2.Y != l2.P2.Y {
		return false
	}
	if l1.Radius != l2.Radius {
		return false
	}
	if l1.Gap != l2.Gap {
		return false
	}
	return true
}
