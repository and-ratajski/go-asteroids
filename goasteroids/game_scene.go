package goasteroids

import (
	"go-asteroids/assets"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/solarlune/resolv"
)

const (
	baseAsteroidVelocity  = 0.25
	asteroidSpawnTime     = 100 * time.Millisecond
	asteroidSpeedUpAmount = 0.1
	asteroidSpeedUpTime   = 1000 * time.Millisecond
	cleanupExplosionTime  = 200 * time.Millisecond
	playerDyingFrames     = 12
)

type GameScene struct {
	player               *Player
	baseVelocity         float64
	asteroidCount        int
	asteroidSpawnTimer   *Timer
	asteroids            map[int]*Asteroid
	asteroidForLevel     int
	velocityTimer        *Timer
	collisionSpace       *resolv.Space
	lasers               map[int]*Laser
	laserCount           int
	score                int
	explosionSmallSprite *ebiten.Image
	explosionSprite      *ebiten.Image
	explosionFrames      []*ebiten.Image
	cleanUpTimer         *Timer
	_isPlayerDead        bool
	audioContext         *audio.Context
	thrustPlayer         *audio.Player
	exhaust              *Exhaust // added dynamically when keyUp pressed
}

func NewGameScene() *GameScene {
	g := &GameScene{
		asteroidSpawnTimer:   NewTimer(asteroidSpawnTime),
		baseVelocity:         baseAsteroidVelocity,
		velocityTimer:        NewTimer(asteroidSpeedUpTime),
		asteroids:            make(map[int]*Asteroid),
		asteroidCount:        0,
		asteroidForLevel:     2,
		collisionSpace:       resolv.NewSpace(ScreenWidth, ScreenHeight, 16, 16), // simple math gave 16?
		lasers:               make(map[int]*Laser),
		laserCount:           0,
		explosionSprite:      assets.ExplosionSprite,
		explosionSmallSprite: assets.ExplosionSmallSprite,
		cleanUpTimer:         NewTimer(cleanupExplosionTime),
		_isPlayerDead:        false,
	}
	g.player = NewPlayer(g)
	g.collisionSpace.Add(g.player.collisionObj)

	g.explosionFrames = assets.Explosion

	// Load audio
	g.audioContext = audio.NewContext(48000) // see docs
	thrustPlayer, _ := g.audioContext.NewPlayer(assets.ThrustSound)
	g.thrustPlayer = thrustPlayer

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()
	g.isPlayerDying()
	g.isPlayerDead(state)

	g.spawnAsteroids()
	for _, a := range g.asteroids {
		a.Update()
	}
	for _, l := range g.lasers {
		l.Update()
	}

	g.speedUpAsteroids()
	g.isPlayerCollidingWithAsteroid()
	g.isAsteroidHitByPlayerLaser()
	g.cleanUpAsteroidsAndAliens()
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	if g.exhaust != nil {
		g.exhaust.Draw(screen)
	}

	for _, a := range g.asteroids {
		a.Draw(screen)
	}
	for _, l := range g.lasers {
		l.Draw(screen)
	}
}

func (g *GameScene) Layout(outsideWidth, outsideHeight int) (ScreenWidth, ScreenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *GameScene) updateExhaust() {
	if g.exhaust != nil {
		g.exhaust.Update()
	}
}

func (g *GameScene) isAsteroidHitByPlayerLaser() {
	for _, a := range g.asteroids {
		for _, l := range g.lasers {
			if a.collisionObj.IsIntersecting(l.collisionObj) {
				if a.collisionObj.Tags().Has(TagSmall) {
					// Laser hit small asteroid
					a.sprite = g.explosionSmallSprite
					g.score++
				} else {
					// Laser hit Large asteroid
					oldPos := a.position

					a.sprite = g.explosionSprite

					g.score++

					numToSpawn := rand.Intn(noOfSmallAsteroidsFromLargerOne)
					for i := 0; i < numToSpawn; i++ {
						_asteroid := NewAsteroid(baseAsteroidVelocity, g, len(a.game.asteroids)-1, "small")
						_asteroid.position = Vector{oldPos.X + float64(rand.Intn(100-50)+50), oldPos.Y + float64(rand.Intn(100-50)+50)}
						_asteroid.collisionObj.SetPosition(_asteroid.position.X, _asteroid.position.Y)
						g.collisionSpace.Add(_asteroid.collisionObj)
						g.asteroidCount++
						g.asteroids[a.game.asteroidCount] = _asteroid
					}
				}
			}
		}
	}
}

func (g *GameScene) isPlayerDying() {
	if g.player.isDying {
		g.player.dyingTimer.Update()
		if g.player.dyingTimer.IsReady() {
			g.player.dyingTimer.Reset()
			g.player.dyingCounter++
			if g.player.dyingCounter == playerDyingFrames { // let the player dye for few frames...
				g.player.isDying = false
				g.player.isDead = true
			} else if g.player.dyingCounter < playerDyingFrames {
				g.player.sprite = g.explosionFrames[g.player.dyingCounter]
			} else {
				// do nothing - might be executed on a tick just before ^ finishes and throw indexOutOfRange error
			}
		}
	}
}

func (g *GameScene) isPlayerDead(state *State) {
	if g._isPlayerDead {
		g.player.livesRemaining--
		if g.player.livesRemaining == 0 {
			g.Reset() // Reset current scene rather than creating a new one because of audio bug
			state.SceneManager.GoToScene(g)
		}
	}
}

func (g *GameScene) spawnAsteroids() {
	g.asteroidSpawnTimer.Update()
	if g.asteroidSpawnTimer.IsReady() {
		g.asteroidSpawnTimer.Reset()
		if len(g.asteroids) < g.asteroidForLevel && g.asteroidCount < g.asteroidForLevel {
			a := NewAsteroid(g.baseVelocity, g, len(g.asteroids)-1)
			g.collisionSpace.Add(a.collisionObj)
			g.asteroidCount++
			g.asteroids[g.asteroidCount] = a
		}
	}
}

func (g *GameScene) speedUpAsteroids() {
	g.velocityTimer.Update()
	if g.velocityTimer.IsReady() {
		g.velocityTimer.Reset()
		g.baseVelocity += asteroidSpeedUpAmount
	}
}

func (g *GameScene) isPlayerCollidingWithAsteroid() {
	for _, a := range g.asteroids {
		if a.collisionObj.IsIntersecting(g.player.collisionObj) {
			if !g.player.isShielded {
				a.game.player.isDying = true
				break
			} else {
				// Bounce the asteroid
			}
		}
	}
}

func (g *GameScene) cleanUpAsteroidsAndAliens() {
	g.cleanUpTimer.Update()
	if g.cleanUpTimer.IsReady() {
		for i, a := range g.asteroids {
			if a.sprite == g.explosionSprite || a.sprite == g.explosionSmallSprite {
				delete(g.asteroids, i)
				g.collisionSpace.Remove(a.collisionObj)
			}
		}
		g.cleanUpTimer.Reset()
	}
}

func (g *GameScene) Reset() {
	// Clear all
	g.player = NewPlayer(g)
	g.asteroids = make(map[int]*Asteroid)
	g.asteroidCount = 0
	g.lasers = make(map[int]*Laser)
	g.laserCount = 0
	g.score = 0
	g.asteroidSpawnTimer.Reset()
	g.baseVelocity = baseAsteroidVelocity
	g.velocityTimer.Reset()
	g._isPlayerDead = false
	g.exhaust = nil
	g.collisionSpace.RemoveAll()

	// Add fresh player obj
	g.collisionSpace.Add(g.player.collisionObj)
}
