package main

// --- 実行時にすること ---
// + Golangの環境を構築する
// + main.go のフォルダでコマンドプロンプトを開く
// + go mod init go-example
// + go mod tidy
//
// + go run main.go

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth          = 640
	ScreenHeight         = 480
	SqSize               = 32
	PlayerMoveTimerMax   = 8
	PlayerActionTimerMax = 20
)

type KeyInputResult int

const (
	KeyInputResultNone KeyInputResult = iota
	KeyInputResultDown
	KeyInputResultUp
	KeyInputResultLeft
	KeyInputResultRight
	KeyInputResultExit
)

type PlayerStatus struct {
	x, y        int
	preX, preY  int
	moveTimer   int
	actionTimer int
}

type Game struct {
	keyInputResult KeyInputResult
	playerStatus   PlayerStatus
}

func NewGame() *Game {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Golang Sammple 01")

	g := new(Game)
	g.playerStatus.x = 0
	g.playerStatus.y = 0
	g.playerStatus.preX = g.playerStatus.x
	g.playerStatus.preY = g.playerStatus.y
	g.playerStatus.moveTimer = 0
	g.playerStatus.actionTimer = 0
	return g
}

func (g *Game) UpdateInput() {
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys[:0])

	g.keyInputResult = KeyInputResultNone
	for _, v := range keys {

		switch v {
		case ebiten.KeyW:
			g.keyInputResult = KeyInputResultUp
		case ebiten.KeyS:
			g.keyInputResult = KeyInputResultDown
		case ebiten.KeyA:
			g.keyInputResult = KeyInputResultLeft
		case ebiten.KeyD:
			g.keyInputResult = KeyInputResultRight
		case ebiten.KeyEscape:
			g.keyInputResult = KeyInputResultExit
		}
	}
}

func (g *Game) UpdateDatas() {
	switch g.keyInputResult {
	case KeyInputResultUp:
		if g.playerStatus.moveTimer == 0 {
			g.playerStatus.y--
			g.playerStatus.moveTimer = PlayerMoveTimerMax
		}
	case KeyInputResultDown:
		if g.playerStatus.moveTimer == 0 {
			g.playerStatus.y++
			g.playerStatus.moveTimer = PlayerMoveTimerMax
		}
	case KeyInputResultLeft:
		if g.playerStatus.moveTimer == 0 {
			g.playerStatus.x--
			g.playerStatus.moveTimer = PlayerMoveTimerMax
		}
	case KeyInputResultRight:
		if g.playerStatus.moveTimer == 0 {
			g.playerStatus.x++
			g.playerStatus.moveTimer = PlayerMoveTimerMax
		}
	case KeyInputResultExit:
		os.Exit(0)
	}

	g.playerStatus.actionTimer++
	if g.playerStatus.actionTimer > PlayerActionTimerMax {
		g.playerStatus.actionTimer = 0
	}

	if g.playerStatus.moveTimer > 0 {
		g.playerStatus.moveTimer--
		if g.playerStatus.moveTimer == 0 {
			g.playerStatus.preX = g.playerStatus.x
			g.playerStatus.preY = g.playerStatus.y
		}
	}
}

func (g *Game) Update() error {
	g.UpdateInput()
	g.UpdateDatas()

	return nil
}

func (g *Game) DrawDebugLog(screen *ebiten.Image) {
	str := fmt.Sprintf("Key %d\nPos %d, %d", g.keyInputResult, g.playerStatus.x, g.playerStatus.y)
	ebitenutil.DebugPrint(screen, str)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xff, 0xff, 0xff, 0xff})

	playerColor := color.RGBA{0xff, uint8(128 + g.playerStatus.actionTimer*3), uint8(128 + g.playerStatus.actionTimer*3), 0xff}

	if g.playerStatus.moveTimer == 0 {
		ebitenutil.DrawRect(screen, float64(g.playerStatus.x*SqSize), float64(g.playerStatus.y*SqSize), SqSize, SqSize, playerColor)
	} else {
		moveX := g.playerStatus.x - g.playerStatus.preX
		moveY := g.playerStatus.y - g.playerStatus.preY

		x := g.playerStatus.preX*SqSize + moveX*SqSize*(PlayerMoveTimerMax-g.playerStatus.moveTimer)/PlayerMoveTimerMax
		y := g.playerStatus.preY*SqSize + moveY*SqSize*(PlayerMoveTimerMax-g.playerStatus.moveTimer)/PlayerMoveTimerMax

		ebitenutil.DrawRect(screen, float64(x), float64(y), SqSize, SqSize, playerColor)
	}

	g.DrawDebugLog(screen)
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
