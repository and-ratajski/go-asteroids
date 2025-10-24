package goasteroids

import (
	"go-asteroids/assets"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var highScore int
var originalHighScore int

// init will be called only once (this is how GO works) and read high score from store
func init() {
	hs, err := getHighScore()
	if err != nil {
		log.Println("Error getting high score: ", err)
	}
	highScore = hs
	originalHighScore = hs
}

type TitleScene struct {
	stars         []*Star
	asteroids     map[int]*Asteroid
	asteroidCount int
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	// Draw stars
	for _, s := range t.stars {
		s.Draw(screen)
	}
	// Draw asteroids
	for _, a := range t.asteroids {
		a.Draw(screen)
	}

	// Draw text
	textToDraw := "1 coin 1 play"
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(float64(ScreenWidth/2), float64(ScreenHeight-200))
	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.TitleFont,
		Size:   48,
	}, op)
}

func (t *TitleScene) Update(state *State) error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		state.SceneManager.GoToScene(NewGameScene())
		return nil
	}

	if len(t.asteroids) < 10 {
		a := NewAsteroid(0.25, &GameScene{}, len(t.asteroids)-1)
		t.asteroidCount++
		t.asteroids[t.asteroidCount] = a
	}

	for _, a := range t.asteroids {
		a.Update()
	}

	return nil
}
