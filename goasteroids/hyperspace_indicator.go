package goasteroids

import (
	"go-asteroids/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type HyperSpaceIndicator struct {
	position Vector
	rotation float64
	sprite   *ebiten.Image
}

func NewHyperSpaceIndicator(pos Vector) *HyperSpaceIndicator {
	return &HyperSpaceIndicator{
		position: pos,
		sprite:   assets.HyperSpaceIndicator,
	}
}

func (hsi *HyperSpaceIndicator) Update() {}

func (hsi *HyperSpaceIndicator) Draw(screen *ebiten.Image) {
	bounds := hsi.sprite.Bounds()
	halfW := float64(bounds.Dx() / 2)
	halfH := float64(bounds.Dy() / 2)

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(halfW, halfH)

	cm := colorm.ColorM{}
	cm.Scale(1.0, 1.0, 1.0, 0.2)

	op.GeoM.Translate(hsi.position.X, hsi.position.Y)
	colorm.DrawImage(screen, hsi.sprite, cm, op)
}
