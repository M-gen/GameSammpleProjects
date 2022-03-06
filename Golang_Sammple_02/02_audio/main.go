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
	_ "image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	ScreenWidth  = 640
	ScreenHeight = 420
	SampleRate   = 32000
)

type KeyInputResult int

type Game struct {
	timer        int
	timerMax     int
	audioContext *audio.Context
	audioPlayer  *AudioPlayer
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
}

func NewPlayer(game *Game, audioContext *audio.Context, musicType musicType, path string) (*AudioPlayer, error) {
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

var ()

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Golang Sammple 01")

	g := new(Game)
	g.timer = 0
	g.timerMax = 20

	g.audioContext = audio.NewContext(SampleRate)

	var err error
	g.audioPlayer, err = NewPlayer(g, g.audioContext, typeOgg, "./sound/coin_01.ogg")
	if err != nil {
		log.Fatal(err)
	}

	return g
}

func (g *Game) Update() error {
	g.timer++
	if g.timer >= g.timerMax {
		g.audioPlayer.Play(0.2)
		g.timer = 0
	}

	return nil
}

func (g *Game) DrawDebugLog(screen *ebiten.Image) {
	str := fmt.Sprintf("timer %d\n", g.timer)
	ebitenutil.DebugPrint(screen, str)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xff, 0xff, 0xff, 0xff})
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
