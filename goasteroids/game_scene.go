package goasteroids

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	baseAsteroidVelocity  = 0.25
	asteroidSpawnTime     = 100 * time.Millisecond
	asteroidSpeedUpAmount = 0.1
	asteroidSpeedUpTime   = 1000 * time.Millisecond
)

type GameScene struct {
	player             *Player
	baseVelocity       float64
	asteroidCount      int
	asteroidSpawnTimer *Timer
	asteroids          map[int]*Asteroid
	asteroidForLevel   int
	velocityTimer      *Timer
	collisionSpace     *resolv.Space
	lasers             map[int]*Laser
	laserCount         int
}

func NewGameScene() *GameScene {
	g := &GameScene{
		asteroidSpawnTimer: NewTimer(asteroidSpawnTime),
		baseVelocity:       baseAsteroidVelocity,
		velocityTimer:      NewTimer(asteroidSpeedUpTime),
		asteroids:          make(map[int]*Asteroid),
		asteroidCount:      0,
		asteroidForLevel:   2,
		collisionSpace:     resolv.NewSpace(ScreenWidth, ScreenHeight, 16, 16), // simple math gave 16?
		lasers:             make(map[int]*Laser),
		laserCount:         0,
	}
	g.player = NewPlayer(g)
	g.collisionSpace.Add(g.player.collisionObj)

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()
	g.spawnAsteroids()

	for _, a := range g.asteroids {
		a.Update()
	}
	for _, l := range g.lasers {
		l.Update()
	}

	g.speedUpAsteroids()
	g.isPlayerCollidingWithAsteroid()
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)
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
			data := a.collisionObj.Data().(*ObjectData)
			fmt.Println("Player collided with asteroid of index: ", data.index)
		}
	}
}
