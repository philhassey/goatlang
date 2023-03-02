package main

import (
	"math"
	"math/rand"
	"time"
)

var total = 0

type Sphere struct {
	x, y, r, vx, vy float64
}

func (o *Sphere) Loop(t float64) {
	o.x += o.vx * t
	o.y += o.vy * t
	if o.x < -128 || o.x > 128 {
		o.vx = -o.vx
	}
	if o.y < -128 || o.y > 128 {
		o.vy = -o.vy
	}
	for _, b := range SPHERES {
		if b == o {
			continue
		}
		dx, dy := b.x-o.x, b.y-o.y
		dd := math.Sqrt(dx*dx + dy*dy)
		if dd < (o.r+b.r) && dd > 0 {
			dx, dy = dx/dd, dy/dd
			inc := (o.r + b.r) - dd
			o.x, o.y = o.x-inc*dx, o.y-inc*dy
			b.x, b.y = b.x+inc*dx, b.y+inc*dy
			total++
		}
	}
}

var SPHERES []*Sphere

func tick() {
	t := 1 / 60.0
	for _, o := range SPHERES {
		o.Loop(t)
	}
}

func main() {
	for i := 0; i < 100; i++ {
		SPHERES = append(SPHERES, &Sphere{
			x:  rand.Float64()*256 - 128,
			y:  rand.Float64()*256 - 128,
			r:  4 + rand.Float64()*5,
			vx: rand.Float64()*100 - 50,
			vy: rand.Float64()*100 - 50,
		})
	}

	ts := time.Now().UnixMilli()
	for i := 0; i < 1000; i++ {
		tick()
	}
	println(total)
	println(time.Now().UnixMilli() - ts)
}
