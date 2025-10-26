package goasteroids

import (
	"fmt"
	"go-asteroids/assets"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type LevelStartsScene struct {
	gameScene      *GameScene
	nextLevelTimer *Timer
	stars          []*Star
}

func (lss *LevelStartsScene) Draw(screen *ebiten.Image) {
	for _, s := range lss.stars {
		s.Draw(screen)
	}
	text2Draw := fmt.Sprintf("LEVEL %d", lss.gameScene.currentLevel)
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, ScreenHeight/2)
	text.Draw(screen, text2Draw, &text.GoTextFace{
		Source: assets.TitleFont,
		Size:   48,
	}, op)

}

func (lss *LevelStartsScene) Update(state *State) error {
	// Move to the next game scene when the timer is ready or spacebar was pressed
	lss.nextLevelTimer.Update()
	if lss.nextLevelTimer.IsReady() || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		lss.gameScene.asteroidForLevel += 2
		lss.gameScene.asteroidCount = 0
		for k, v := range lss.gameScene.lasers {
			delete(lss.gameScene.lasers, k)
			lss.gameScene.collisionSpace.Remove(v.collisionObj)
		}
		state.SceneManager.GoToScene(lss.gameScene)
	}

	return nil
}
