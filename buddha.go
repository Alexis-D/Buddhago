package main

import (
	"flag"
	//"fmt"
	"image"
	"image/png"
	"os"
	"rand"
	//"runtime"
)

const (
	RED uint64 = iota
	GREEN
	BLUE
)

// define the "global" vars that may be changed
// by the command line arguments
var (
	width *uint64 = flag.Uint64("w", 512,
		"The width of the output image.")
	height *uint64 = flag.Uint64("h", 512,
		"The height of the output image.")
	xmin *float64 = flag.Float64("xmin", -2.,
		"The leftmost part of the plan.")
	xmax *float64 = flag.Float64("xmax", 1.,
		"The rightmost part of the plan.")
	ymin *float64 = flag.Float64("ymin", -1.5,
		"The bottom part of the plan.")
	ymax *float64 = flag.Float64("ymax", 1.5,
		"The top part of the plan.")
	riter *uint64 = flag.Uint64("r", 250,
		"Max number of iterations for the red channel.")
	giter *uint64 = flag.Uint64("g", 100,
		"Max number of iterations for the green channel.")
	biter *uint64 = flag.Uint64("b", 500,
		"Max number of iterations for the blue channel.")
	goroutines *uint64 = flag.Uint64("goroutines", 8,
		"Number of goroutines to use.")
	points *uint64 = flag.Uint64("points", 2560000,
		"Number of points to generate.")
	out *string = flag.String("o", "out.png",
		"Output file (.png).")
	help *bool = flag.Bool("help", false,
		"Display the help message.")
)

// used to speed up computation
var (
	xratio float64 = 0
	yratio float64 = 0
)

// used in generatePoints to send the point & the color in the same channel
type complexPoint struct {
	z     complex128
	color uint64
}

// convert z to it's coordinate in pixel in
// the output image
func cpx2px(z complex128) (uint64, uint64, os.Error) {
	re, im := real(z), imag(z)
	x := uint64((re - *xmin) * xratio)
	y := uint64((*ymax - im) * yratio)

	if x < 0 || x >= *width || y < 0 || y >= *height {
		return 0, 0, os.NewError("Point outside of the image.")
	}

	return x, y, nil
}

// return true if z is this in the plan
func inPlan(z complex128) bool {
	re, im := real(z), imag(z)
	return re >= *xmin && re <= *xmax && im >= *ymin && im <= *ymax
}

// return the maximum of a, b and c
func max(a, b, c uint64) uint64 {
	if a > b {
		if a > c {
			return a
		}

		return c
	}

	if b > c {
		return b
	}

	return c
}

// try npoints different c points, and if for some c the suite
// z_(n+1) = z_n ** 2 + c escape, then it send the points
// in ch, once it finished it send one value in quit
func generatePoints(npoints uint64, ch chan complexPoint, quit chan bool) {
	defer func() { quit <- true }()

	// store the maximum number of iterations allowed
	m := max(*riter, *giter, *biter)

	for i := uint64(0); i < npoints; i++ {
		// pick a new random c
		re := rand.Float64()*(*xmax-*xmin) + *xmin
		im := rand.Float64()*(*ymax-*ymin) + *ymin
		c := complex(re, im)
		z := complex(0, 0)

		// to store the "path"
		p := make([]complex128, 0)

		// used to know if it escaped, so not in the for
		j := uint64(0)

		for ; j < m; j++ {
			z *= z
			z += c

			p = append(p, z)

			// it escaped
			// use > 4, and not cmath.Abs because it's faster
			re, im = real(z), imag(z)

			if re*re+im*im > 4 {
				// continue until we are outside of the plan
				// because it's ugly otherwise (draw a circle)
				for inPlan(z) {
					z *= z
					z += c
					p = append(p, z)
				}

				break
			}
		}

		// send the points to the channel only
		// if j is below the max number of iteration
		// of the color
		f := func(iter *uint64, color uint64) {
			if j < *iter {
				for _, z := range p {
					ch <- complexPoint{z, color}
				}
			}
		}

		f(riter, RED)
		f(giter, BLUE)
		f(biter, GREEN)
	}
}

// Combine the three color channel and write them
// to the out file
func renderImage(red [][]uint64, green [][]uint64, blue [][]uint64) {
	im := image.NewNRGBA64(int(*width), int(*height))

	var (
		minr             uint64 = 1<<64 - 1
		ming             uint64 = minr
		minb             uint64 = minr
		maxr, maxg, maxb uint64 = 0, 0, 0
	)

	f := func(count uint64, min, max *uint64) {
		if count < *min {
			*min = count
		}

		if count > *max {
			*max = count
		}
	}

	// pick the min/max for each channel
	// to be able to do the colouring
	for y := 0; y < int(*height); y++ {
		for x := 0; x < int(*width); x++ {
			r, g, b := red[y][x], green[y][x], blue[y][x]

			f(r, &minr, &maxr)
			f(g, &ming, &maxg)
			f(b, &minb, &maxb)
		}
	}

	// set each pixel to its color
	for y := 0; y < int(*height); y++ {
		for x := 0; x < int(*width); x++ {
			r, g, b := red[y][x], green[y][x], blue[y][x]
			var m float64 = 1<<16 - 1

			im.SetNRGBA64(x, y, image.NRGBA64Color{
				uint16(m * float64(r) / float64(maxr-minr)),
				uint16(m * float64(g) / float64(maxg-ming)),
				uint16(m * float64(b) / float64(maxb-minb)),
				1<<16 - 1})
		}
	}

	w, _ := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer w.Close()

	// write the png
	png.Encode(w, im)
}

// do some initialization, get all the points
// and then generate the output file
func renderBuddha() {
	// to store the count of each pixel
	red := make([][]uint64, *height)
	green := make([][]uint64, *height)
	blue := make([][]uint64, *height)

	for i := uint64(0); i < *height; i++ {
		red[i] = make([]uint64, *width)
		green[i] = make([]uint64, *width)
		blue[i] = make([]uint64, *width)
	}

	ch := make(chan complexPoint)
	quit := make(chan bool)

	for i := uint64(0); i < *goroutines; i++ {
		// yeah!
		go generatePoints(*points / *goroutines, ch, quit)
	}

	// finished counts the number of finished goroutines
	for finished := uint64(0); finished != *goroutines; {
		select {
		case p := <-ch:
			if x, y, err := cpx2px(p.z); err == nil {
				switch p.color {
				case RED:
					red[y][x]++
				case GREEN:
					green[y][x]++
				case BLUE:
					blue[y][x]++
				}
			}
			// we don't care if there's an error, just
			// discard the point
		case <-quit:
			finished++
		}
	}

	renderImage(red, green, blue)
}

func main() {
	flag.Parse()
	//runtime.GOMAXPROCS(4) // should be set through $GOMAXPROCS ideally

	if *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	xratio = float64(*width) / (*xmax - *xmin)
	yratio = float64(*height) / (*ymax - *ymin)

	renderBuddha()
}
