package main

// --- 実行時にすること ---
// + Golangの環境を構築する
// + main.go のフォルダでコマンドプロンプトを開く
// + go mod init go-example
// + go mod tidy
//
// + go run main.go

import (
	"bytes"
	"image"
	_ "image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth          = 640
	ScreenHeight         = 480
	SqSize               = 32
	PlayerMoveTimerMax   = 8
	PlayerActionTimerMax = 20
	SampleRate           = 32000
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

type Position struct {
	x, y int
}

type Sq struct {
	position Position
	isIn     bool
}

type MapData struct {
	Sq   []Sq
	Size Position
}

func (mapData *MapData) GetSq(x int, y int) Sq {
	var i = x + y*mapData.Size.x
	return mapData.Sq[i]
}

type Camera struct {
	pos           Position
	drawOffsetPos Position
}

func (camera *Camera) Update(player PlayerStatus) {
	mx := camera.pos.x - player.x*SqSize - SqSize/2
	my := camera.pos.y - player.y*SqSize - SqSize/2
	camera.pos.x = camera.pos.x - mx*1/10
	camera.pos.y = camera.pos.y - my*1/10
	camera.drawOffsetPos.x = camera.pos.x - ScreenWidth/2
	camera.drawOffsetPos.y = camera.pos.y - ScreenHeight/2
}

type Coin struct {
	pos Position
}

type Game struct {
	keyInputResult  KeyInputResult
	playerStatus    PlayerStatus
	tileImageA      *ebiten.Image
	tileImageB      *ebiten.Image
	tileImageWall   *ebiten.Image
	playerImage     *ebiten.Image
	mapData         MapData
	camera          Camera
	coin            []Coin
	coinImage       *ebiten.Image
	audioContext    *audio.Context
	audioPlayerCoin *AudioPlayer
	audioPlayerMove *AudioPlayer
}

type musicType int

const (
	typeOgg musicType = iota
	typeMP3
)

func (t musicType) String() string {
	switch t {
	case typeOgg:
		return "Ogg"
	case typeMP3:
		return "MP3"
	default:
		panic("not reached")
	}
}

// Player represents the current audio state.
type AudioPlayer struct {
	game         *Game
	audioContext *audio.Context
	audioPlayer  *audio.Player
	current      time.Duration
	total        time.Duration
	seBytes      []byte
	seCh         chan []byte
	volume128    int
	musicType    musicType

	playButtonPosition  image.Point
	alertButtonPosition image.Point
}

func NewAudioPlayer(game *Game, audioContext *audio.Context, musicType musicType, path string) (*AudioPlayer, error) {
	type audioStream interface {
		io.ReadSeeker
		Length() int64
	}

	const bytesPerSample = 4 // TODO: This should be defined in audio package

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var s audioStream

	switch musicType {
	case typeOgg:
		var err error
		s, err = vorbis.Decode(audioContext, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case typeMP3:
		var err error
		s, err = mp3.Decode(audioContext, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		panic("not reached")
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return nil, err
	}
	player := &AudioPlayer{
		game:         game,
		audioContext: audioContext,
		audioPlayer:  p,
		total:        time.Second * time.Duration(s.Length()) / bytesPerSample / SampleRate,
		volume128:    128,
		seCh:         make(chan []byte),
		musicType:    musicType,
	}
	if player.total == 0 {
		player.total = 1
	}

	return player, nil
}

func (p *AudioPlayer) Play(volume float64) {
	p.audioPlayer.SetVolume(volume)
	p.audioPlayer.Seek(0)
	for p.audioPlayer.Current() != 0 {
		// 再生開始一を、先頭にシークできていない場合がある
		// このとき、再度Seek(0)を呼ばないと、変化しないようなので、再度呼び出して強行する
		p.audioPlayer.Seek(0)
	}
	p.audioPlayer.Play()
}

func (p *AudioPlayer) Close() error {
	return p.audioPlayer.Close()
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Golang Sammple 01")

	g := new(Game)
	g.playerStatus.x = 5
	g.playerStatus.y = 5
	g.playerStatus.preX = g.playerStatus.x
	g.playerStatus.preY = g.playerStatus.y
	g.playerStatus.moveTimer = 0
	g.playerStatus.actionTimer = 0

	g.audioContext = audio.NewContext(SampleRate)

	var err error
	var tmpImage *ebiten.Image
	tmpImage, _, err = ebitenutil.NewImageFromFile("./image/Tiles.png")
	if err != nil {
		log.Fatal(err)
	}
	g.tileImageWall = tmpImage.SubImage(image.Rect(0, 0, SqSize, SqSize)).(*ebiten.Image)
	g.tileImageA = tmpImage.SubImage(image.Rect(32, 0, 32+SqSize, SqSize)).(*ebiten.Image)
	g.tileImageB = tmpImage.SubImage(image.Rect(32, 32, 32+SqSize, 32*SqSize)).(*ebiten.Image)

	tmpImage, _, err = ebitenutil.NewImageFromFile("./image/EmugenIdle.png")
	if err != nil {
		log.Fatal(err)
	}
	g.playerImage = tmpImage.SubImage(image.Rect(0, 0, 48, 48)).(*ebiten.Image)

	g.audioPlayerCoin, err = NewAudioPlayer(g, g.audioContext, typeOgg, "./sound/coin_01.ogg")
	if err != nil {
		log.Fatal(err)
	}

	g.audioPlayerMove, err = NewAudioPlayer(g, g.audioContext, typeOgg, "./sound/Jump-6.ogg")
	if err != nil {
		log.Fatal(err)
	}

	// マップの初期化
	g.mapData.Size.x = 15
	g.mapData.Size.y = 10
	g.mapData.Sq = make([]Sq, g.mapData.Size.x*g.mapData.Size.y, g.mapData.Size.x*g.mapData.Size.y)

	for i := 0; i < len(g.mapData.Sq); i++ {
		var v = &g.mapData.Sq[i]
		var x = i % g.mapData.Size.x
		var y = i / g.mapData.Size.x
		if x == 0 || y == 0 || x == g.mapData.Size.x-1 || y == g.mapData.Size.y-1 {
			v.isIn = false
		} else {
			v.isIn = true
		}
		v.position.x = x
		v.position.y = y
	}

	// コインの初期化
	g.coinImage, _, _ = ebitenutil.NewImageFromFile("./image/Coin.png")
	coinNum := 10
	g.coin = make([]Coin, coinNum, coinNum)
	for i := 0; i < len(g.coin); i++ {
		g.coin[i].pos.x = rand.Intn(g.mapData.Size.x-2) + 1
		g.coin[i].pos.y = rand.Intn(g.mapData.Size.y-2) + 1
	}

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
			if g.mapData.GetSq(g.playerStatus.x, g.playerStatus.y-1).isIn {
				g.playerStatus.y--
				g.playerStatus.moveTimer = PlayerMoveTimerMax
				g.PlayerMove()
			}
		}
	case KeyInputResultDown:
		if g.playerStatus.moveTimer == 0 {
			if g.mapData.GetSq(g.playerStatus.x, g.playerStatus.y+1).isIn {
				g.playerStatus.y++
				g.playerStatus.moveTimer = PlayerMoveTimerMax
				g.PlayerMove()
			}
		}
	case KeyInputResultLeft:
		if g.playerStatus.moveTimer == 0 {
			if g.mapData.GetSq(g.playerStatus.x-1, g.playerStatus.y).isIn {
				g.playerStatus.x--
				g.playerStatus.moveTimer = PlayerMoveTimerMax
				g.PlayerMove()
			}
		}
	case KeyInputResultRight:
		if g.playerStatus.moveTimer == 0 {
			if g.mapData.GetSq(g.playerStatus.x+1, g.playerStatus.y).isIn {
				g.playerStatus.x++
				g.playerStatus.moveTimer = PlayerMoveTimerMax
				g.PlayerMove()
			}
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

	g.camera.Update(g.playerStatus)
}

func (g *Game) Update() error {
	g.UpdateInput()
	g.UpdateDatas()

	return nil
}

func (g *Game) PlayerMove() {
	g.audioPlayerMove.Play(1.0)

	for i := 0; i < len(g.coin); i++ {
		if (g.coin[i].pos.x == g.playerStatus.x) && (g.coin[i].pos.y == g.playerStatus.y) {
			g.coin[i].pos.x = -1 // 削除が大変なので、マップ外に配置する
			g.coin[i].pos.y = -1
			g.audioPlayerCoin.Play(1.0)
		}
	}
}

func (g *Game) DrawDebugLog(screen *ebiten.Image) {
	str := fmt.Sprintf("Key %d\nPos %d, %d", g.keyInputResult, g.playerStatus.x, g.playerStatus.y)
	ebitenutil.DebugPrint(screen, str)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xff, 0xff, 0xff, 0xff})

	for _, v := range g.mapData.Sq {
		if v.isIn {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(SqSize*v.position.x-g.camera.drawOffsetPos.x), float64(SqSize*v.position.y-g.camera.drawOffsetPos.y))
			if (v.position.x+v.position.y)%2 == 0 {
				screen.DrawImage(g.tileImageA, op)
			} else {
				screen.DrawImage(g.tileImageB, op)
			}

		} else {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(SqSize*v.position.x-g.camera.drawOffsetPos.x), float64(SqSize*v.position.y-g.camera.drawOffsetPos.y))
			screen.DrawImage(g.tileImageWall, op)
		}
	}

	for _, v := range g.coin {
		if v.pos.x == -1 {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(SqSize*v.pos.x-g.camera.drawOffsetPos.x+8), float64(SqSize*v.pos.y-g.camera.drawOffsetPos.y+8))
		screen.DrawImage(g.coinImage, op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(SqSize*g.playerStatus.x-8-g.camera.drawOffsetPos.x), float64(SqSize*g.playerStatus.y-14-g.camera.drawOffsetPos.y))
	screen.DrawImage(g.playerImage, op)

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
