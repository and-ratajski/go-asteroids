package main

import (
	"go-asteroids/assets"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	rotationPerSecond = math.Pi
	maxAcceleration   = 8.0
	ScreenWidth       = 1280
	ScreenHeight      = 720 // 16:9 aspect ratio
)

var curtAcceleration float64

type Player struct {
	game     *Game
	sprite   *ebiten.Image
	position Vector
	rotation float64
	velocity float64
}

func NewPlayer(game *Game) *Player {
	sprite := assets.PlayerSprite

	// Center player on screen
	bounds := sprite.Bounds()
	bCenterX := float64(bounds.Dx()) / 2
	bCenterY := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - bCenterX,
		Y: ScreenHeight/2 - bCenterY,
	}
	p := &Player{
		sprite:   sprite,
		game:     game,
		position: pos,
	}
	return p
}

func (p *Player) Draw(screen *ebiten.Image) {
	bounds := p.sprite.Bounds()
	halfWidth := float64(bounds.Dx()) / 2
	halfHeight := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfWidth, -halfHeight)
	op.GeoM.Rotate(p.rotation)
	op.GeoM.Translate(halfWidth, halfHeight)

	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

func (p *Player) Update() {
	speed := rotationPerSecond / float64(ebiten.TPS())

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}

	p.accelerate()
}

func (p *Player) accelerate() {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.keepOnScreen()

		if curtAcceleration < maxAcceleration {
			curtAcceleration = p.velocity + 4
		}
		if curtAcceleration > maxAcceleration {
			curtAcceleration = maxAcceleration
		}

		p.velocity = curtAcceleration

		// Move in the direction we are pointing
		dx := math.Sin(p.rotation) * curtAcceleration
		dy := -math.Cos(p.rotation) * curtAcceleration

		// Move the player on the screen
		p.position.X += dx
		p.position.Y += dy
	}
}

func (p *Player) keepOnScreen() {
	if p.position.X >= float64(ScreenWidth) {
		p.position.X = 0
	}
	if p.position.X < 0 {
		p.position.X = float64(ScreenWidth)
	}
	if p.position.Y >= float64(ScreenHeight) {
		p.position.Y = 0
	}
	if p.position.Y < 0 {
		p.position.Y = float64(ScreenHeight)
	}
}
