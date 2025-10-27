package goasteroids

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Star struct {
	x          float32
	y          float32
	r          float32
	brightness float32
	color      color.RGBA // Cache color to avoid allocation each frame
}

func NewStar() *Star {
	brightness := rand.Float32() * 0xff
	return &Star{
		x:          rand.Float32() * ScreenWidth,
		y:          rand.Float32() * ScreenHeight,
		r:          rand.Float32() * (3 - 1),
		brightness: brightness,
		color: color.RGBA{
			R: uint8(0xbb * brightness / 0xff),
			G: uint8(0xdd * brightness / 0xff),
			B: uint8(0xff * brightness / 0xff),
			A: 0xff,
		},
	}
}

func (s *Star) Draw(screen *ebiten.Image) {
	vector.FillCircle(screen, s.x, s.y, s.r, s.color, true)
}

func (s *Star) Update() {}

func GenerateStars(n int) []*Star {
	var stars []*Star
	for i := 0; i < n; i++ {
		stars = append(stars, NewStar())
	}
	return stars
}
