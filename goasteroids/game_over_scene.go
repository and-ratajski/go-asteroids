package goasteroids

import (
	"go-asteroids/assets"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type GameOverScene struct {
	gameScene     *GameScene
	asteroids     map[int]*Asteroid
	asteroidCount int
	stars         []*Star
}

func (o *GameOverScene) Draw(screen *ebiten.Image) {
	for _, s := range o.stars {
		s.Draw(screen)
	}
	for _, a := range o.asteroids {
		a.Draw(screen)
	}

	text2Draw := "Game Over"
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, ScreenHeight/2+100)
	text.Draw(screen, text2Draw, &text.GoTextFace{
		Source: assets.TitleFont,
		Size:   48,
	}, op)

	// Draw new high score
	if o.gameScene.score > originalHighScore {
		text2Draw = "New High Score!"
		op := &text.DrawOptions{
			LayoutOptions: text.LayoutOptions{
				PrimaryAlign: text.AlignCenter,
			},
		}
		op.ColorScale.ScaleWithColor(color.White)
		op.GeoM.Translate(ScreenWidth/2, ScreenHeight/2-200)
		text.Draw(screen, text2Draw, &text.GoTextFace{
			Source: assets.TitleFont,
			Size:   48,
		}, op)
	}
}

func (o *GameOverScene) Update(state *State) error {
	// Spawn asteroids
	if len(o.asteroids) < 10 {
		a := NewAsteroid(0.25, &GameScene{}, len(o.asteroids)-1)
		o.asteroidCount++
		o.asteroids[o.asteroidCount] = a
	}

	// Update asteroids
	for _, a := range o.asteroids {
		a.Update()
	}

	// Check to see if spacebar is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		o.gameScene.Reset()
		o.gameScene.currentLevel = 1
		state.SceneManager.GoToScene(o.gameScene)
	}

	// Check to see if time to quit
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	return nil
}
