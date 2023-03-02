package main

import "time"

const (
	its    = 8192
	escape = 4.0
	width  = 160.0
	height = 50.0
)

func inMandelbrot(cx, cy float64) bool {
	var zx, zy float64
	for i := 0; i < its; i++ {
		zx, zy = zx*zx-zy*zy+cx, 2.0*zx*zy+cy
		if zx*zx+zy*zy > escape {
			return false
		}
	}
	return true
}

func main() {
	ts := time.Now().UnixMilli()
	for y := 0; y < height; y++ {
		i := float64(y*2-height) / float64(height)
		for x := 0; x < width; x++ {
			r := float64(x*2-width*3/2) / float64(width)
			if inMandelbrot(r, i) {
				print("*")
			} else {
				print(" ")
			}
		}
		println()
	}
	println(time.Now().UnixMilli() - ts)
}
