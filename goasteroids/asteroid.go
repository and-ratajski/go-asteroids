package goasteroids

import (
	"go-asteroids/assets"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	rotationSpeedMin                = -0.02
	rotationSpeedMax                = 0.02
	noOfSmallAsteroidsFromLargerOne = 4
)

type Asteroid struct {
	game          *GameScene
	position      Vector
	rotation      float64
	movement      Vector
	angle         float64
	rotationSpeed float64
	sprite        *ebiten.Image
	collisionObj  *resolv.Circle
}

// NewAsteroid is a factory method to create new (random) asteroid - a pointer to it
func NewAsteroid(baseVelocity float64, g *GameScene, index int) *Asteroid {
	// Target the center of the screen
	target := Vector{
		X: ScreenWidth / 2,
		Y: ScreenHeight / 2,
	}

	// Pick a random angle for the asteroid
	angle := rand.Float64() * 2 * math.Pi

	// Distance from the center that asteroid should spawn at.
	// Half the width, add some arbitrary distance
	radius := float64(ScreenWidth/2 + 500)

	// Create the position vector, using the angle and simple math
	position := Vector{
		X: target.X + math.Cos(angle)*radius,
		Y: target.Y + math.Sin(angle)*radius,
	}

	// Keep the asteroid moving towards the center of the screen
	velocity := baseVelocity + rand.Float64()*1.5
	direction := Vector{
		X: target.X - position.X,
		Y: target.Y - position.Y,
	}
	normalizedDirection := direction.Normalize()
	movement := Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	// Assign a sprite to the asteroid
	sprite := assets.AsteroidsSprites[rand.Intn(len(assets.AsteroidsSprites))]

	// Create the collision object
	collisionObj := resolv.NewCircle(position.X, position.Y, float64(sprite.Bounds().Dx()/2))

	// Create an asteroid object and return
	a := &Asteroid{
		game:          g,
		position:      position,
		movement:      movement,
		rotationSpeed: rotationSpeedMin + rand.Float64()*(rotationSpeedMax-rotationSpeedMin),
		sprite:        sprite,
		angle:         angle,
		collisionObj:  collisionObj,
	}

	// Fill collision object data
	a.collisionObj.SetPosition(position.X, position.Y)
	a.collisionObj.Tags().Set(TagAsteroid | TagLarge)
	a.collisionObj.SetData(&ObjectData{index: index})

	return a
}

func (a *Asteroid) Update() {
	dx := a.movement.X
	dy := a.movement.Y

	a.position.X += dx
	a.position.Y += dy
	a.rotation += a.rotationSpeed

	a.keepOnScreen()
	a.collisionObj.SetPosition(a.position.X, a.position.Y)
}

func (a *Asteroid) Draw(screem *ebiten.Image) {
	bounds := a.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(a.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(a.position.X, a.position.Y)

	screem.DrawImage(a.sprite, op)
}

func (a *Asteroid) keepOnScreen() {
	if a.position.X >= float64(ScreenWidth) {
		a.position.X = 0
		a.collisionObj.SetPosition(0, a.position.Y)
	}
	if a.position.X < 0 {
		a.position.X = ScreenWidth
		a.collisionObj.SetPosition(ScreenWidth, a.position.Y)
	}
	if a.position.Y >= float64(ScreenHeight) {
		a.position.Y = 0
		a.collisionObj.SetPosition(a.position.X, 0)
	}
	if a.position.Y < 0 {
		a.position.Y = ScreenHeight
		a.collisionObj.SetPosition(a.position.X, ScreenHeight)
	}
}
