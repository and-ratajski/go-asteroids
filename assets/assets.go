package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed *
var assets embed.FS

var TitleFont = mustLoadFontFace("fonts/title.ttf")
var PlayerSprite = mustLoadImage("images/player.png")
var AsteroidsSprites = mustLoadImages("images/meteors/*.png") // Asteroids imgs are called Meteors...
var AsteroidsSpritesSmall = mustLoadImages("images/meteors-small/*.png")
var LaserSprite = mustLoadImage("images/laser.png")

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		panic(err)
	}

	images := make([]*ebiten.Image, len(matches))
	for i, match := range matches {
		images[i] = mustLoadImage(match)
	}

	return images
}

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		panic(err)
	}
	defer func(f fs.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadFontFace(name string) *text.GoTextFaceSource {
	f, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}

	r := bytes.NewReader(f)
	ts, err := text.NewGoTextFaceSource(r)
	if err != nil {
		panic(err)
	}

	return ts
}
