package goasteroids

import (
	"go-asteroids/assets"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
)

const (
	ScreenWidth                 = 1280
	ScreenHeight                = 720 // 16:9 aspect ratio
	rotationPerSecond           = math.Pi
	maxForwardAcceleration      = 8.0
	forwardAccelerationIncrease = 4.0
	maxReverseAcceleration      = 2.0
	shootCoolDown               = 150 * time.Millisecond
	burstCoolDown               = 500 * time.Millisecond
	laserSpawnOffset            = 50.0
	maxShotsPerBurst            = 3
	dyingAnimationOffset        = 50 * time.Millisecond
	numberOfLives               = 3
	numberOfShields             = 3
	shieldDuration              = 6 * time.Second
	hyperSpaceCooldown          = 10 * time.Second
	driftTime                   = 30 * time.Second
)

var curtAcceleration float64
var shotsFired = 0

type Player struct {
	game                *GameScene
	sprite              *ebiten.Image
	position            Vector
	rotation            float64
	velocity            float64
	collisionObj        *resolv.Circle
	shootCoolDown       *Timer
	burstCoolDown       *Timer
	isShielded          bool
	isDying             bool
	isDead              bool
	dyingTimer          *Timer
	dyingCounter        int
	livesRemaining      int
	lifeIndicators      []*LifeIndicator
	shieldTimer         *Timer
	shieldsRemaining    int
	shieldIndicators    []*ShieldIndicator
	hyperSpaceIndicator *HyperSpaceIndicator
	hyperSpaceTimer     *Timer
	driftTimer          *Timer
	driftAngle          float64 // While player is drifting, changing their angle shouldn't change drift angle
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

	// Add live indicators - number of icons displayed one next to another
	var lifeIndicators []*LifeIndicator
	var liStartPositionX = 20.0
	for i := 0; i < numberOfLives; i++ {
		li := NewLifeIndicator(Vector{X: liStartPositionX, Y: 20.0})
		lifeIndicators = append(lifeIndicators, li)
		liStartPositionX += 50.0
	}

	// Add shield indicators - number of icons displayed one next to another
	var shieldIndicators []*ShieldIndicator
	var siStartPositionX = 45.0
	for i := 0; i < numberOfShields; i++ {
		si := NewShieldIndicator(Vector{X: siStartPositionX, Y: 60.0})
		shieldIndicators = append(shieldIndicators, si)
		siStartPositionX += 50.0
	}

	p := &Player{
		sprite:              sprite,
		game:                game,
		position:            pos,
		collisionObj:        collisionObj,
		shootCoolDown:       NewTimer(shootCoolDown),
		burstCoolDown:       NewTimer(burstCoolDown),
		isShielded:          false,
		isDying:             false,
		isDead:              false,
		dyingTimer:          NewTimer(dyingAnimationOffset),
		dyingCounter:        0,
		livesRemaining:      numberOfLives,
		lifeIndicators:      lifeIndicators,
		shieldsRemaining:    numberOfShields,
		shieldIndicators:    shieldIndicators,
		hyperSpaceIndicator: NewHyperSpaceIndicator(Vector{X: 37.0, Y: 95.0}), // found by trail an error
		hyperSpaceTimer:     nil,
		driftTimer:          nil,
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
	p.isPlayerDead()

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}

	p.accelerate()
	p.isDoneAccelerating()

	p.reverse()
	p.isDoneReversing()

	p.isPlayerDrifting()
	p.isDoneDrifting()

	p.updateExhaustSprite()
	p.useShield()

	p.collisionObj.SetPosition(p.position.X, p.position.Y)

	p.burstCoolDown.Update()
	p.shootCoolDown.Update()
	p.fireLasers()

	p.hyperSpace()
	if p.hyperSpaceTimer != nil {
		p.hyperSpaceTimer.Update()
	}
}

func (p *Player) isPlayerDead() {
	if p.isDead {
		p.game._isPlayerDead = true
	}
}

func (p *Player) isPlayerDrifting() {
	if p.driftTimer != nil {
		p.keepOnScreen()

		p.driftTimer.Update()
		decelerationSpeed := p.velocity / float64(ebiten.TPS()) * 2

		p.position.X += math.Sin(p.driftAngle) * decelerationSpeed
		p.position.Y += -math.Cos(p.driftAngle) * decelerationSpeed
		p.collisionObj.SetPosition(p.position.X, p.position.Y)
	}

}

func (p *Player) isDoneDrifting() {
	if p.driftTimer != nil && p.driftTimer.IsReady() {
		p.driftTimer = nil
		p.velocity = 0
	}
}

func (p *Player) hyperSpace() {
	if ebiten.IsKeyPressed(ebiten.KeyH) && (p.hyperSpaceTimer == nil || p.hyperSpaceTimer.IsReady()) {
		var randX, randY int
		for {
			randX = rand.Intn(ScreenWidth)
			randX = rand.Intn(ScreenHeight)

			collision := p.game.checkCollision(p.collisionObj, nil)
			if !collision {
				break
			}
		}

		p.position.X = float64(randX)
		p.position.Y = float64(randY)

		if p.hyperSpaceTimer == nil {
			p.hyperSpaceTimer = NewTimer(hyperSpaceCooldown)
		}
		p.hyperSpaceTimer.Reset()
	}
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
				p.game.laserCount++ // No clearing?

				switch shotsFired {
				case 1:
					if !p.game.laserPlayerOne.IsPlaying() {
						_ = p.game.laserPlayerOne.Rewind()
						p.game.laserPlayerOne.Play()
					}
				case 2:
					if !p.game.laserPlayerTwo.IsPlaying() {
						_ = p.game.laserPlayerTwo.Rewind()
						p.game.laserPlayerTwo.Play()
					}
				case 3:
					if !p.game.laserPlayerThree.IsPlaying() {
						_ = p.game.laserPlayerThree.Rewind()
						p.game.laserPlayerThree.Play()
					}
				}
			} else {
				p.burstCoolDown.Reset()
				shotsFired = 0
			}
		}
	}
}

func (p *Player) accelerate() {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.driftTimer = nil
		p.keepOnScreen()

		if curtAcceleration < maxForwardAcceleration {
			curtAcceleration = p.velocity + forwardAccelerationIncrease
		}
		if curtAcceleration > maxForwardAcceleration {
			curtAcceleration = maxForwardAcceleration
		}

		p.velocity = curtAcceleration

		// Move in the direction we are pointing
		dx := math.Sin(p.rotation) * curtAcceleration
		dy := -math.Cos(p.rotation) * curtAcceleration

		// Show exhaust
		bounds := p.sprite.Bounds()
		halfW := float64(bounds.Dx() / 2)
		halfH := float64(bounds.Dy() / 2)

		exhSpawnPos := Vector{
			p.position.X + halfW + math.Sin(p.rotation)*exhaustSpawnOffest,
			p.position.Y + halfH - math.Cos(p.rotation)*exhaustSpawnOffest,
		}
		p.game.exhaust = NewExhaust(exhSpawnPos, p.rotation+180.0*math.Pi/180.0)

		// Move the player on the screen
		p.position.X += dx
		p.position.Y += dy

		// Check if the sound is not already playing
		if !p.game.thrustPlayer.IsPlaying() {
			_ = p.game.thrustPlayer.Rewind() // Important to rewind
			p.game.thrustPlayer.Play()
		}
	}
}

func (p *Player) isDoneAccelerating() {
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		if p.game.thrustPlayer.IsPlaying() {
			p.game.thrustPlayer.Pause()
		}

		// Figure out velocity
		if p.velocity < curtAcceleration*10 {
			p.velocity = curtAcceleration*10 - 5.0
		}
		if p.velocity < 0 {
			p.velocity = 0
		}

		// Otherwise they would still accelerate?
		curtAcceleration = 0

		p.driftTimer = NewTimer(driftTime)
		p.driftAngle = p.rotation
	}
}

func (p *Player) reverse() {
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.driftTimer = nil
		p.keepOnScreen()

		dx := -math.Sin(p.rotation) * maxReverseAcceleration
		dy := math.Cos(p.rotation) * maxReverseAcceleration

		bounds := p.sprite.Bounds()
		halfW := float64(bounds.Dx() / 2)
		halfH := float64(bounds.Dy() / 2)

		exhSpawnPos := Vector{
			p.position.X + halfW - math.Sin(p.rotation)*exhaustSpawnOffest,
			p.position.Y + halfH + math.Cos(p.rotation)*exhaustSpawnOffest,
		}
		p.game.exhaust = NewExhaust(exhSpawnPos, p.rotation+180.0*math.Pi/180.0)

		p.position.X += dx
		p.position.Y += dy

		p.collisionObj.SetPosition(p.position.X, p.position.Y)

		if !p.game.thrustPlayer.IsPlaying() {
			_ = p.game.thrustPlayer.Rewind()
			p.game.thrustPlayer.Play()
		}
	}
}

func (p *Player) isDoneReversing() {
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		if p.game.thrustPlayer.IsPlaying() {
			p.game.thrustPlayer.Pause()
		}
	}
}

func (p *Player) updateExhaustSprite() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) && !ebiten.IsKeyPressed(ebiten.KeyDown) && p.game.exhaust != nil {
		p.game.exhaust = nil
	}
}

func (p *Player) useShield() {
	if ebiten.IsKeyPressed(ebiten.KeyS) && !p.isShielded && p.shieldsRemaining > 0 {
		if !p.game.shieldsUpPlayer.IsPlaying() {
			_ = p.game.shieldsUpPlayer.Rewind()
			p.game.shieldsUpPlayer.Play()
		}

		p.isShielded = true
		p.shieldTimer = NewTimer(shieldDuration)
		p.game.shield = NewShield(Vector{}, p.rotation, p.game)
		p.shieldsRemaining--
		p.shieldIndicators = p.shieldIndicators[:len(p.shieldIndicators)-1]
	}

	if p.shieldTimer != nil && p.isShielded {
		p.shieldTimer.Update()
	}

	if p.shieldTimer != nil && p.shieldTimer.IsReady() {
		p.shieldTimer = nil
		p.isShielded = false
		p.game.collisionSpace.Remove(p.game.shield.collisionObj)
		p.game.shield = nil
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
