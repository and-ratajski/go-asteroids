package goasteroids

import (
	"fmt"
	"go-asteroids/assets"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/solarlune/resolv"
)

const (
	baseAsteroidVelocity  = 0.25
	asteroidSpawnTime     = 100 * time.Millisecond
	asteroidSpeedUpAmount = 0.1
	asteroidSpeedUpTime   = 1000 * time.Millisecond
	cleanupExplosionTime  = 200 * time.Millisecond
	playerDyingFrames     = 12
	baseBeatWaitTime      = 1600
	numberOfStars         = 750
	maxNumberOfAliens     = 1
	alienAttackTime       = 3 * time.Second
	alienSpawnTime        = 6 * time.Second
	baseAlienVelocity     = 0.5
)

type GameScene struct {
	player               *Player
	baseVelocity         float64
	stars                []*Star
	asteroidCount        int
	asteroidSpawnTimer   *Timer
	asteroids            map[int]*Asteroid
	asteroidForLevel     int
	velocityTimer        *Timer
	collisionSpace       *resolv.Space
	lasers               map[int]*Laser
	laserCount           int
	score                int
	currentLevel         int
	explosionSmallSprite *ebiten.Image
	explosionSprite      *ebiten.Image
	explosionFrames      []*ebiten.Image
	cleanUpTimer         *Timer
	_isPlayerDead        bool
	audioContext         *audio.Context
	thrustPlayer         *audio.Player
	exhaust              *Exhaust // added dynamically when keyUp pressed
	laserPlayerOne       *audio.Player
	laserPlayerTwo       *audio.Player
	laserPlayerThree     *audio.Player
	explosionPlayer      *audio.Player
	beatPlayerOne        *audio.Player
	beatPlayerTwo        *audio.Player
	beatTimer            *Timer
	beatWaitTime         int
	playBeatOne          bool
	shield               *Shield
	shieldsUpPlayer      *audio.Player
	aliens               map[int]*Alien
	alienCount           int
	alienAttackTimer     *Timer
	alienSoundPlayer     *audio.Player
	alienSpawnTimer      *Timer
	alienLasers          map[int]*Laser
	alienLaserCount      int
	alienLaserPlayer     *audio.Player
	// Cache for text rendering to reduce string formatting overhead
	cachedScoreText     string
	cachedHighScoreText string
	cachedLevelText     string
	lastScore           int
	lastHighScore       int
	lastLevel           int
}

func NewGameScene() *GameScene {
	g := &GameScene{
		asteroidSpawnTimer:   NewTimer(asteroidSpawnTime),
		baseVelocity:         baseAsteroidVelocity,
		velocityTimer:        NewTimer(asteroidSpeedUpTime),
		stars:                GenerateStars(numberOfStars),
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
		beatTimer:            NewTimer(2 * time.Second),
		beatWaitTime:         baseBeatWaitTime,
		currentLevel:         1,
		aliens:               make(map[int]*Alien),
		alienCount:           0,
		alienLasers:          make(map[int]*Laser),
		alienLaserCount:      0,
		alienSpawnTimer:      NewTimer(alienSpawnTime),
		alienAttackTimer:     NewTimer(alienAttackTime),
	}
	g.player = NewPlayer(g)
	g.collisionSpace.Add(g.player.collisionObj)

	g.explosionFrames = assets.Explosion

	// Load audio
	g.audioContext = audio.NewContext(48000) // see docs
	thrustPlayer, _ := g.audioContext.NewPlayer(assets.ThrustSound)
	g.thrustPlayer = thrustPlayer

	laserPlayerOne, _ := g.audioContext.NewPlayer(assets.LaserSoundOne)
	g.laserPlayerOne = laserPlayerOne

	laserPlayerTwo, _ := g.audioContext.NewPlayer(assets.LaserSoundTwo)
	g.laserPlayerTwo = laserPlayerTwo

	laserPlayerThree, _ := g.audioContext.NewPlayer(assets.LaserSoundThree)
	g.laserPlayerThree = laserPlayerThree

	explosionPlayer, _ := g.audioContext.NewPlayer(assets.ExplosionSound)
	g.explosionPlayer = explosionPlayer

	beatPlayerOne, _ := g.audioContext.NewPlayer(assets.BeatSoundOne)
	g.beatPlayerOne = beatPlayerOne
	beatPlayerTwo, _ := g.audioContext.NewPlayer(assets.BeatSoundTwo)
	g.beatPlayerTwo = beatPlayerTwo

	shieldsUpPlayer, _ := g.audioContext.NewPlayer(assets.ShieldSound)
	g.shieldsUpPlayer = shieldsUpPlayer

	alienSoundPlayer, _ := g.audioContext.NewPlayer(assets.AlienSound)
	alienSoundPlayer.SetVolume(0.5)
	g.alienSoundPlayer = alienSoundPlayer

	alienLaserPlayer, _ := g.audioContext.NewPlayer(assets.AlienLaserSound)
	g.alienLaserPlayer = alienLaserPlayer

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()

	g.updateExhaust() // works exactly the same without it..
	g.updateShield()

	g.isPlayerDying()
	g.isPlayerDead(state)

	g.spawnAsteroids()
	for _, a := range g.asteroids {
		a.Update()
	}

	g.spawnAliens()
	for _, a := range g.aliens {
		a.Update()
	}
	g.letAliensAttack()

	for _, l := range g.lasers {
		l.Update()
	}
	for _, al := range g.alienLasers {
		al.Update()
	}

	g.speedUpAsteroids()
	g.isPlayerCollidingWithAsteroid()
	g.isAsteroidHitByPlayerLaser()

	g.isPlayerCollidingWithAlien()
	g.isPlayerHitByAlienLaser()
	g.isAlienHitByPlayerLaser()

	g.beatSound()
	g.isLevelCompleted(state)

	g.cleanUpAsteroidsAndAliens()
	g.removeOffscreenAliens()
	g.removeOffscreenLasers()

	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	for _, s := range g.stars {
		s.Draw(screen)
	}
	g.player.Draw(screen)

	for _, a := range g.asteroids {
		a.Draw(screen)
	}
	for _, l := range g.lasers {
		l.Draw(screen)
	}
	if g.exhaust != nil {
		g.exhaust.Draw(screen)
	}

	if g.shield != nil {
		g.shield.Draw(screen)
	}

	if len(g.player.lifeIndicators) > 0 {
		for _, li := range g.player.lifeIndicators {
			li.Draw(screen)
		}
	}
	if len(g.player.shieldIndicators) > 0 {
		for _, si := range g.player.shieldIndicators {
			si.Draw(screen)
		}
	}
	if g.player.hyperSpaceTimer == nil || g.player.hyperSpaceTimer.IsReady() {
		g.player.hyperSpaceIndicator.Draw(screen)
	}

	for _, a := range g.aliens {
		a.Draw(screen)
	}
	for _, al := range g.alienLasers {
		al.Draw(screen)
	}

	// Draw the score and the high score (with caching to reduce string formatting overhead)
	if g.score != g.lastScore {
		g.cachedScoreText = fmt.Sprintf("%06d", g.score)
		g.lastScore = g.score
	}
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, 40)
	text.Draw(screen, g.cachedScoreText, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   24,
	}, op)

	if g.score >= highScore {
		highScore = g.score
	}
	if highScore != g.lastHighScore {
		g.cachedHighScoreText = fmt.Sprintf("HIGH SCORE %06d", highScore)
		g.lastHighScore = highScore
	}
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, 75)
	text.Draw(screen, g.cachedHighScoreText, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   16,
	}, op)

	// Draw current level (with caching)
	if g.currentLevel != g.lastLevel {
		g.cachedLevelText = fmt.Sprintf("LEVEL %d", g.currentLevel)
		g.lastLevel = g.currentLevel
	}
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, ScreenHeight-40)
	text.Draw(screen, g.cachedLevelText, &text.GoTextFace{
		Source: assets.LevelFont,
		Size:   16,
	}, op)
}

func (g *GameScene) Layout(outsideWidth, outsideHeight int) (ScreenWidth, ScreenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *GameScene) spawnAliens() {
	g.alienSpawnTimer.Update()
	if g.alienSpawnTimer.IsReady() && g.alienCount < maxNumberOfAliens {
		g.alienSpawnTimer.Reset()
		rnd := rand.Intn(100-1) + 1
		if rnd > 50 {
			a := NewAlien(baseAlienVelocity, g)
			g.collisionSpace.Add(a.collisionObj)
			g.alienCount++
			g.aliens[g.alienCount] = a
		}
	}
}

func (g *GameScene) isLevelCompleted(state *State) {
	if g.asteroidCount >= g.asteroidForLevel && len(g.asteroids) == 0 {
		g.baseVelocity = baseAsteroidVelocity
		g.currentLevel++

		// Give extra live after surviving each 5 levels
		if g.currentLevel%5 == 0 {
			if g.player.livesRemaining < 6 {
				g.player.livesRemaining++
				liX := float64(20 + len(g.player.lifeIndicators)*50.0)
				liY := 20.0
				g.player.lifeIndicators = append(g.player.lifeIndicators, NewLifeIndicator(Vector{liX, liY}))
			}
		}

		g.beatWaitTime = baseBeatWaitTime

		state.SceneManager.GoToScene(&LevelStartsScene{
			gameScene:      g,
			stars:          GenerateStars(numberOfStars),
			nextLevelTimer: NewTimer(2 * time.Second),
		})
	}
}

func (g *GameScene) beatSound() {
	g.beatTimer.Update()
	if g.beatTimer.IsReady() {
		if g.playBeatOne {
			_ = g.beatPlayerOne.Rewind()
			g.beatPlayerOne.Play()
			g.beatTimer.Reset()
		} else {
			_ = g.beatPlayerTwo.Rewind()
			g.beatPlayerTwo.Play()
			g.beatTimer.Reset()
		}
		g.playBeatOne = !g.playBeatOne

		// Speed up the Timer for playing beats
		if g.beatWaitTime > 400 {
			g.beatWaitTime = g.beatWaitTime - 25
			g.beatTimer = NewTimer(time.Duration(g.beatWaitTime) * time.Millisecond)
		}
	}
}

func (g *GameScene) updateExhaust() {
	if g.exhaust != nil {
		g.exhaust.Update()
	}
}

func (g *GameScene) updateShield() {
	if g.shield != nil {
		g.shield.Update()
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
					g.playExplosionSound()
				} else {
					// Laser hit Large asteroid
					oldPos := a.position
					a.sprite = g.explosionSprite

					g.score++
					g.playExplosionSound()
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
			// DEAD
			if g.score > originalHighScore {
				err := updateHighScore(g.score)
				if err != nil {
					log.Println(err)
				}
			}
			state.SceneManager.GoToScene(&GameOverScene{
				gameScene:     g,
				asteroids:     make(map[int]*Asteroid),
				asteroidCount: 5,
				stars:         GenerateStars(numberOfStars),
			})
		} else {
			// NOT DEAD
			score := g.score
			livesRemaining := g.player.livesRemaining
			lifeSlice := g.player.lifeIndicators[:len(g.player.lifeIndicators)-1]
			stars := g.stars
			shieldsRemaining := g.player.shieldsRemaining
			shieldSlice := g.player.shieldIndicators

			g.Reset()

			g.player.livesRemaining = livesRemaining
			g.score = score
			g.player.lifeIndicators = lifeSlice
			g.stars = stars
			g.player.shieldsRemaining = shieldsRemaining
			g.player.shieldIndicators = shieldSlice
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
				g.playExplosionSound()
				break
			} else {
				g.bounceAsteroid(a)
			}
		}
	}
}

func (g *GameScene) bounceAsteroid(a *Asteroid) {
	direction := Vector{
		X: -1 * (ScreenWidth/2 - a.position.X),
		Y: -1 * (ScreenHeight/2 - a.position.Y),
	}
	normalizedDirection := direction.Normalize()
	velocity := g.baseVelocity // todo multiply velocity after bounce

	movement := Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}
	a.movement = movement
}

func (g *GameScene) isPlayerCollidingWithAlien() {
	for _, a := range g.aliens {
		if a.collisionObj.IsIntersecting(g.player.collisionObj) {
			if !g.player.isShielded {
				if !a.gameScene.explosionPlayer.IsPlaying() {
					_ = a.gameScene.explosionPlayer.Rewind()
					a.gameScene.explosionPlayer.Play()
				}
				a.gameScene.player.isDying = true
			}
		}
	}
}

func (g *GameScene) isPlayerHitByAlienLaser() {
	for _, l := range g.alienLasers {
		if l.collisionObj.IsIntersecting(g.player.collisionObj) {
			if !g.player.isShielded {
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
				g.player.isDying = true
			}
		}
	}
}

func (g *GameScene) isAlienHitByPlayerLaser() {
	for _, alien := range g.aliens {
		for _, l := range g.lasers {
			if alien.collisionObj.IsIntersecting(l.collisionObj) {
				laserData := l.collisionObj.Data().(*ObjectData)
				delete(g.alienLasers, laserData.index)
				g.collisionSpace.Remove(l.collisionObj)
				alien.sprite = g.explosionSprite
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
				g.score += 50
			}
		}
	}
}

func (g *GameScene) letAliensAttack() {
	if len(g.aliens) > 0 {
		if !g.alienSoundPlayer.IsPlaying() {
			_ = g.alienSoundPlayer.Rewind()
			g.alienSoundPlayer.Play()
		}

		g.alienAttackTimer.Update()
		if g.alienAttackTimer.IsReady() {
			g.alienAttackTimer.Reset()
			for _, a := range g.aliens {
				bounds := a.sprite.Bounds()
				halfW := float64(bounds.Dx() / 2)
				halfH := float64(bounds.Dy() / 2)

				var degreesRadian float64

				if !a.isIntelligent {
					// Fire in a random direction
					degreesRadian = rand.Float64() * (2 * math.Pi)
				} else {
					// Fire with some accuracy
					degreesRadian = math.Atan2(
						g.player.position.Y-a.position.Y,
						g.player.position.X-a.position.X,
					)
					degreesRadian += 0.5 * math.Pi // decrease accuracy
				}

				offsetX := float64(bounds.Dx() - int(halfW))
				offsetY := float64(bounds.Dy() - int(halfH))
				spawnPos := Vector{
					X: a.position.X + halfW + math.Sin(degreesRadian) - offsetX,
					Y: a.position.Y + halfW + math.Cos(degreesRadian) - offsetY,
				}

				laser := NewLaser(g, spawnPos, degreesRadian, g.alienLaserCount)
				laser.sprite = assets.AlienLaserSprite // todo update constructor
				g.alienLaserCount++
				g.alienLasers[g.alienLaserCount] = laser
				if !g.alienLaserPlayer.IsPlaying() {
					_ = g.alienLaserPlayer.Rewind()
					g.alienLaserPlayer.Play()
				}
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
		for i, alien := range g.aliens {
			if alien.sprite == g.explosionSprite {
				delete(g.aliens, i)
				g.collisionSpace.Remove(alien.collisionObj)
			}
		}
		g.cleanUpTimer.Reset()
	}
}

func (g *GameScene) removeOffscreenAliens() {
	for i, a := range g.aliens {
		outOfRangeX := a.position.X > ScreenWidth+200 || a.position.X < -200 // Add 200 to make sure it's outside the screen
		outOfRangeY := a.position.Y > ScreenWidth+200 || a.position.Y < -200
		if outOfRangeX || outOfRangeY {
			delete(g.aliens, i)
			g.alienCount--
			g.collisionSpace.Remove(a.collisionObj)
		}

	}
}

func (g *GameScene) removeOffscreenLasers() {
	lasers2Remove := [](map[int]*Laser){
		g.lasers,
		g.alienLasers,
	}

	for _, lasers := range lasers2Remove {
		for i, l := range lasers {
			outOfRangeX := l.position.X > ScreenWidth+200 || l.position.X < -200
			outOfRangeY := l.position.Y > ScreenWidth+200 || l.position.Y < -200
			if outOfRangeX || outOfRangeY {
				g.collisionSpace.Remove(l.collisionObj)
				delete(lasers, i)
			}
		}
	}
}

func (g *GameScene) playExplosionSound() {
	if !g.explosionPlayer.IsPlaying() {
		_ = g.explosionPlayer.Rewind()
		g.explosionPlayer.Play()
	}
}

func (g *GameScene) Reset() {
	// Clear all but audio
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
	g.player.isShielded = false // to be super safe, when user dies just after opening a shield
	g.aliens = make(map[int]*Alien)
	g.alienCount = 0
	g.alienLasers = make(map[int]*Laser)
	g.alienLaserCount = 0

	// Add fresh player obj
	g.collisionSpace.Add(g.player.collisionObj)
	g.stars = GenerateStars(numberOfStars)
	g.player.shieldsRemaining = numberOfShields
}
