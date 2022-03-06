package main

// --- 実行時にすること ---
// + Golangの環境を構築する
// + main.go のフォルダでコマンドプロンプトを開く
// + go mod init go-example
// + go mod tidy
//
// + go run main.go

import (
	"image"
	_ "image/png"

	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	ScreenWidth  = 640
	ScreenHeight = 480
)

type Game struct {
	playerImage *ebiten.Image
}

func NewGame() *Game {
	g := new(Game)
	var err error
	var tmpImage *ebiten.Image

	tmpImage, _, err = ebitenutil.NewImageFromFile("./image/EmugenIdle.png")
	if err != nil {
		log.Fatal(err)
	}
	g.playerImage = tmpImage.SubImage(image.Rect(0, 0, 48, 48)).(*ebiten.Image)

	return g
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) PlayerMove() {
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xff, 0xff, 0xff, 0xff})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64((ScreenWidth-48)/2), float64((ScreenHeight-48)/2))
	screen.DrawImage(g.playerImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	var g = NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
