package goasteroids

import (
	"go-asteroids/assets"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type TitleScene struct {
	asteroids     map[int]*Asteroid
	asteroidCount int
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
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

	for _, a := range t.asteroids {
		a.Draw(screen)
	}
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
