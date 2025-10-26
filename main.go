package main

import (
	"go-asteroids/goasteroids"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("Go Asteroids, Go!")
	ebiten.SetWindowSize(goasteroids.ScreenWidth, goasteroids.ScreenHeight)

	// Hide cursor
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	// Set to full screen
	//ebiten.SetFullscreen(true)

	err := ebiten.RunGame(&goasteroids.Game{})
	if err != nil {
		panic(err)
	}
}
