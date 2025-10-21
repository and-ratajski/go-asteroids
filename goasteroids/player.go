package goasteroids

import (
	"go-asteroids/assets"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	rotationPerSecond = math.Pi
	maxAcceleration   = 8.0
	ScreenWidth       = 1280
	ScreenHeight      = 720 // 16:9 aspect ratio
	shootCoolDown     = 150 * time.Millisecond
	burstCoolDown     = 500 * time.Millisecond
	laserSpawnOffset  = 50.0
	maxShotsPerBurst  = 3
)

var curtAcceleration float64
var shotsFired = 0

type Player struct {
	game          *GameScene
	sprite        *ebiten.Image
	position      Vector
	rotation      float64
	velocity      float64
	collisionObj  *resolv.Circle
	shootCoolDown *Timer
	burstCoolDown *Timer
}

func NewPlayer(game *GameScene) *Player {
	sprite := assets.PlayerSprite

	// Center player on screen
	bounds := sprite.Bounds()
	bCenterX := float64(bounds.Dx()) / 2
	bCenterY := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - bCenterX,
		Y: ScreenHeight/2 - bCenterY,
	}
	collisionObj := resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2))

	p := &Player{
		sprite:        sprite,
		game:          game,
		position:      pos,
		collisionObj:  collisionObj,
		shootCoolDown: NewTimer(shootCoolDown),
		burstCoolDown: NewTimer(burstCoolDown),
	}
	p.collisionObj.SetPosition(pos.X, pos.Y)
	p.collisionObj.Tags().Set(TagPlayer)

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
	p.collisionObj.SetPosition(p.position.X, p.position.Y)

	p.burstCoolDown.Update()
	p.shootCoolDown.Update()
	p.fireLasers()
}

func (p *Player) fireLasers() {
	if p.burstCoolDown.IsReady() {
		if p.shootCoolDown.IsReady() && ebiten.IsKeyPressed(ebiten.KeySpace) {
			p.shootCoolDown.Reset()
			shotsFired++
			if shotsFired <= maxShotsPerBurst {
				bound := p.sprite.Bounds()
				halfW := float64(bound.Dx() / 2)
				halfH := float64(bound.Dy() / 2)

				spawnPos := Vector{
					X: p.position.X + halfW + math.Sin(p.rotation)*laserSpawnOffset,
					Y: p.position.Y + halfH + (-math.Cos(p.rotation))*laserSpawnOffset,
				}
				laser := NewLaser(p.game, spawnPos, p.rotation, p.game.laserCount)
				p.game.lasers[p.game.laserCount] = laser
				p.game.collisionSpace.Add(laser.collisionObj)
				p.game.laserCount++
			} else {
				p.burstCoolDown.Reset()
				shotsFired = 0
			}
		}
	}
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
		p.collisionObj.SetPosition(0, p.position.Y)
	}
	if p.position.X < 0 {
		p.position.X = float64(ScreenWidth)
		p.collisionObj.SetPosition(ScreenWidth, p.position.Y)
	}
	if p.position.Y >= float64(ScreenHeight) {
		p.position.Y = 0
		p.collisionObj.SetPosition(p.position.X, 0)
	}
	if p.position.Y < 0 {
		p.position.Y = float64(ScreenHeight)
		p.collisionObj.SetPosition(p.position.X, ScreenHeight)
	}
}
