package goasteroids

import "github.com/hajimehoshi/ebiten/v2"

type Game struct {
	sceneManager *SceneManager
	input        Input
}

// Input is just a stub, but it's needed for Ebitengine
type Input struct{}

func (i *Input) Update() {}

func (g *Game) Update() error {
	if g.sceneManager == nil {
		g.sceneManager = &SceneManager{}
		g.sceneManager.GoToScene(&TitleScene{})
	}

	g.input.Update() // Allows keystrokes go to various scenes
	if err := g.sceneManager.Update(&g.input); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sceneManager.Draw(screen)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}
