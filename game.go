package main

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"math/rand"
	"os"
	"time"
)

var (
	ColorBackGround   = color.White
	ColorGridNegative = color.RGBA{R: 238, G: 238, B: 238, A: 255}
	ColorGridPositive = color.RGBA{R: 207, G: 111, B: 193, A: 255}
)

func NewCB(w, h int, ratio float32) *CB {
	cb := &CB{
		Rect:      image.Rect(0, 0, w, h),
		Sequences: make([]int, w*h),
		C:         make(chan *image.Paletted, 4),
	}
	for x := 0; x < cb.Rect.Dx(); x++ {
		for y := 0; y < cb.Rect.Dy(); y++ {
			if rand.Float32() < ratio {
				cb.SetPositive(x, y)
			}
		}
	}
	return cb
}

type CB struct {
	Rect      image.Rectangle
	Sequences []int
	C         chan *image.Paletted
}

func (cb *CB) Get(x, y int) int     { return cb.Sequences[x+y*cb.Rect.Dx()] }
func (cb *CB) Set(x, y int, v int)  { cb.Sequences[x+y*cb.Rect.Dx()] = v }
func (cb *CB) SetPositive(x, y int) { cb.Set(x, y, 1) }
func (cb *CB) SetNegative(x, y int) { cb.Set(x, y, 0) }

func (cb *CB) GetLib(x, y int) int {
	sum := 0
	for px := x - 1; px <= x+1; px++ {
		for py := y - 1; py <= y+1; py++ {
			if px == x && py == y {
				continue
			}
			if px < 0 || px >= cb.Rect.Dx() || py < 0 || py >= cb.Rect.Dy() {
				continue
			}
			sum += cb.Get(px, py)
		}
	}
	return sum
}
func (cb *CB) Draw() *image.Paletted {
	w := cb.Rect.Size().X*12 + 2
	h := cb.Rect.Size().Y*12 + 2
	m := image.NewPaletted(image.Rect(0, 0, w, h), palette.Plan9)
	for px := 0; px < w; px++ {
		for py := 0; py < h; py++ {
			m.Set(px, py, ColorBackGround)
		}
	}
	for x := 0; x < cb.Rect.Dx(); x++ {
		for y := 0; y < cb.Rect.Dy(); y++ {
			var color color.Color
			if cb.Get(x, y) == 1 {
				color = ColorGridPositive
			} else {
				color = ColorGridNegative
			}
			for px := 12*x + 2; px < 12*(x+1); px++ {
				for py := 12*y + 2; py < 12*(y+1); py++ {
					m.Set(px, py, color)
				}
			}
		}
	}
	return m
}

func (cb *CB) Gen() {
	for x := 0; x < cb.Rect.Dx(); x++ {
		for y := 0; y < cb.Rect.Dy(); y++ {
			switch cb.GetLib(x, y) {
			case 0, 1, 4, 5, 6, 7, 8:
				cb.SetNegative(x, y)
			case 3:
				cb.SetPositive(x, y)
			case 2:
			}
		}
	}
	cb.C <- cb.Draw()
}

func (cb *CB) GenN(n int) {
	for i := 0; i < n; i++ {
		cb.Gen()
	}
	close(cb.C)
}

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	cb := NewCB(60, 10, 0.6)
	go cb.GenN(64)
	anim := gif.GIF{LoopCount: 64}
	for m := range cb.C {
		anim.Delay = append(anim.Delay, 8)
		anim.Image = append(anim.Image, m)
	}
	gif.EncodeAll(os.Stdout, &anim)
}
