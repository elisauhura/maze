package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	. "github.com/elUrso/maze/levels"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var start time.Time
var finish time.Time
var last time.Time
var freeze bool

type game struct {
	started   bool
	completed bool
	l         int
	s         *stage
	objs      []object
	player    object
	exit      object
	hold      *object
}

type stage struct {
	name  string
	cells [][]cell
	board *ebiten.Image
	x, y  int
}

type object struct {
	x, y int
	t, o string
}

type cell struct {
	valid, wu, wd, wl, wr bool
}

func (g *game) load() {
	if g.l == len(Levels) {
		g.completed = true
		finish = time.Now()
		return
	}

	if g.completed {
		return
	}

	last = time.Now()

	s := &stage{}

	s.name = Levels[g.l].Name

	buff := strings.NewReader(Levels[g.l].Data)

	header := ""
	fmt.Fscanf(buff, "%v\n", &header)
	if header != "mazemap" {
		print("Invalid map")
		os.Exit(1)
	}

	var x, y int
	fmt.Fscanf(buff, "%v %v\n", &x, &y)

	s.x = x
	s.y = y

	s.cells = make([][]cell, y)
	for i := 0; i < y; i++ {
		line := make([]cell, x)
		s.cells[i] = line
	}

	s.readLevel(buff, x, y)

	s.renderBoard()

	// read elements
	var e int
	var t, o string

	g.objs = []object{}
	g.hold = nil

	fmt.Fscanf(buff, "%v\n", &e)

	for i := 0; i < e; i++ {

		fmt.Fscanf(buff, "%v %v %v %v\n", &x, &y, &t, &o)

		g.objs = append(g.objs, object{x, y, t, o})
	}

	fmt.Fscanf(buff, "%v %v %v\n", &x, &y, &o)
	g.player = object{x, y, "p", o}

	fmt.Fscanf(buff, "%v %v\n", &x, &y)
	g.exit = object{x, y, "e", "u"}

	g.s = s
}

func (s *stage) readLevel(buff io.Reader, x, y int) {
	for i := 0; i < y; i++ {
		line := ""
		fmt.Fscanf(buff, "%v\n", &line)
		for j := 0; j < x; j++ {
			if line[j] != 'x' {
				s.cells[i][j].valid = true
			} else {
				continue
			}
			// wall up
			if i == 0 || !s.cells[i-1][j].valid || s.cells[i-1][j].wd {
				s.cells[i][j].wu = true
			}
			// wall left
			if j == 0 || !s.cells[i][j-1].valid || s.cells[i][j-1].wr {
				s.cells[i][j].wl = true
			}
			// wall down
			if line[j] == 'd' || line[j] == 'b' {
				s.cells[i][j].wd = true
			}
			// wall right
			if line[j] == 'r' || line[j] == 'b' {
				s.cells[i][j].wr = true
			}
		}
	}
}

func (s *stage) renderBoard() {
	s.board = ebiten.NewImage(s.x*32, s.y*32)
	for i := 0; i < s.y; i++ {
		for j := 0; j < s.x; j++ {
			if s.cells[i][j].valid {
				ebitenutil.DrawRect(s.board, float64(j*32), float64(i*32), 32, 32, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
			}
			if s.cells[i][j].wu {
				ebitenutil.DrawRect(s.board, float64(j*32), float64(i*32), 32, 4, color.RGBA{0x40, 0x40, 0x40, 0xFF})
			}
			if s.cells[i][j].wl {
				ebitenutil.DrawRect(s.board, float64(j*32), float64(i*32), 4, 32, color.RGBA{0x40, 0x40, 0x40, 0xFF})
			}
			if s.cells[i][j].wr {
				ebitenutil.DrawRect(s.board, float64(j*32+28), float64(i*32), 4, 32, color.RGBA{0x40, 0x40, 0x40, 0xFF})
			}
			if s.cells[i][j].wd {
				ebitenutil.DrawRect(s.board, float64(j*32), float64(i*32+28), 32, 4, color.RGBA{0x40, 0x40, 0x40, 0xFF})
			}
		}
	}
}

func (g *game) Update() error {
	if freeze {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			freeze = false
			g.load()
		}

		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && (!g.started || g.completed) {
		g.started = true
		g.completed = false
		g.l = 0
		g.load()
		start = time.Now()
	}

	if g.completed {
		return nil
	}

	if !g.started {
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.load()
		return nil
	}

	now := time.Now()
	if now.Sub(last).Seconds() > 1 {
		last = now
		g.updateObjs()
	}

	var move bool

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.player.o = "u"
		move = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.player.o = "d"
		move = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.player.o = "r"
		move = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.player.o = "l"
		move = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.hold == nil {
			for i, o := range g.objs {
				if o.x == g.player.x && o.y == g.player.y && o.t == "g" && o.o == g.player.o {
					g.hold = &g.objs[i]
					g.hold.x = -1
				}
			}
		} else {
			switch g.player.o {
			case "u":
				if !g.s.cells[g.player.y][g.player.x].wu && g.noWall(g.player.x, g.player.y, g.player.o) {
					g.hold.x = g.player.x
					g.hold.y = g.player.y
					g.hold.o = g.player.o
					g.hold = nil
				}
			case "l":
				if !g.s.cells[g.player.y][g.player.x].wl && g.noWall(g.player.x, g.player.y, g.player.o) {
					g.hold.x = g.player.x
					g.hold.y = g.player.y
					g.hold.o = g.player.o
					g.hold = nil
				}
			case "r":
				if !g.s.cells[g.player.y][g.player.x].wr && g.noWall(g.player.x, g.player.y, g.player.o) {
					g.hold.x = g.player.x
					g.hold.y = g.player.y
					g.hold.o = g.player.o
					g.hold = nil
				}
			case "d":
				if !g.s.cells[g.player.y][g.player.x].wd && g.noWall(g.player.x, g.player.y, g.player.o) {
					g.hold.x = g.player.x
					g.hold.y = g.player.y
					g.hold.o = g.player.o
					g.hold = nil
				}
			}
		}
	}

	if move {
		g.movePlayer()
	}

	for _, o := range g.objs {
		if o.t != "g" {
			// check if player is visible
			ox := o.x
			oy := o.y
			ok := true
			for ok {
				if g.player.x == ox && g.player.y == oy {
					g.l = 0
					freeze = true
					return nil
				}
				ox, oy, ok = g.next(ox, oy, o.o)
			}

		}
	}

	if g.player.x == g.exit.x && g.player.y == g.exit.y && g.hold == nil {
		g.l = g.l + 1
		g.load()
	}

	return nil
}

func (g *game) movePlayer() {
	p := &g.player
	c := g.s.cells

	switch g.player.o {
	case "u":
		if p.y > 0 && c[p.y-1][p.x].valid && !c[p.y][p.x].wu && g.noWall(p.x, p.y, "u") && g.noWall(p.x, p.y-1, "d") {
			p.y = p.y - 1
		}
	case "l":
		if p.x > 0 && c[p.y][p.x-1].valid && !c[p.y][p.x].wl && g.noWall(p.x, p.y, "l") && g.noWall(p.x-1, p.y, "r") {
			p.x = p.x - 1
		}
	case "r":
		if p.x < len(c[0])-1 && c[p.y][p.x+1].valid && !c[p.y][p.x].wr && g.noWall(p.x, p.y, "r") && g.noWall(p.x+1, p.y, "l") {
			p.x = p.x + 1
		}
	case "d":
		if p.y < len(c)-1 && c[p.y+1][p.x].valid && !c[p.y][p.x].wd && g.noWall(p.x, p.y, "d") && g.noWall(p.x, p.y+1, "u") {
			p.y = p.y + 1
		}
	}
}

func (g *game) next(ox, oy int, o string) (int, int, bool) {
	c := g.s.cells
	for _, ob := range g.objs {
		if ob.x == ox && ob.y == oy && ob.t == "g" && ob.o == o {
			return 0, 0, false
		}
	}
	switch o {
	case "u":
		if oy > 0 && c[oy-1][ox].valid && !c[oy][ox].wu {
			for _, o := range g.objs {
				if o.x == ox && o.y == oy-1 {
					if o.t == "g" && o.o != "d" {

					} else {
						return 0, 0, false
					}
				}
			}
			if g.exit.x == ox && g.exit.y == oy-1 {
				return 0, 0, false
			}
			return ox, oy - 1, true
		}
	case "l":
		if ox > 0 && c[oy][ox-1].valid && !c[oy][ox].wl {
			for _, o := range g.objs {
				if o.x == ox-1 && o.y == oy {
					if o.t == "g" && o.o != "r" {

					} else {
						return 0, 0, false
					}
				}
			}
			if g.exit.x == ox-1 && g.exit.y == oy {
				return 0, 0, false
			}
			return ox - 1, oy, true
		}
	case "r":
		if ox < len(c[0])-1 && c[oy][ox+1].valid && !c[oy][ox].wr {
			for _, o := range g.objs {
				if o.x == ox+1 && o.y == oy {
					if o.t == "g" && o.o != "l" {

					} else {
						return 0, 0, false
					}
				}
			}
			if g.exit.x == ox+1 && g.exit.y == oy {
				return 0, 0, false
			}
			return ox + 1, oy, true
		}
	case "d":
		if oy < len(c)-1 && c[oy+1][ox].valid && !c[oy][ox].wd {
			for _, o := range g.objs {
				if o.x == ox && o.y == oy+1 {
					if o.t == "g" && o.o != "u" {

					} else {
						return 0, 0, false
					}
				}
			}
			if g.exit.x == ox && g.exit.y == oy+1 {
				return 0, 0, false
			}
			return ox, oy + 1, true
		}
	}
	return 0, 0, false
}

func (g *game) updateObjs() {
	for i, o := range g.objs {
		if o.t != "g" {
			x, y, ok := g.next(o.x, o.y, o.o)
			if ok {
				g.objs[i].x = x
				g.objs[i].y = y
			} else {
				g.objs[i].o = nextO(o.t, o.o)
			}
		}
	}
}

func nextO(t, o string) string {
	a := make(map[string]string)
	switch t {
	case "r":
		a["u"] = "r"
		a["l"] = "u"
		a["r"] = "d"
		a["d"] = "l"
	case "l":
		a["u"] = "l"
		a["l"] = "d"
		a["r"] = "u"
		a["d"] = "r"
	case "b":
		a["u"] = "d"
		a["l"] = "r"
		a["r"] = "l"
		a["d"] = "u"
	}
	return a[o]
}

func (g *game) noWall(x, y int, or string) bool {
	for _, o := range g.objs {
		if o.x == x && o.y == y && o.t == "g" && o.o == or {
			return false
		}
	}
	return true
}

func (g *game) Draw(screen *ebiten.Image) {
	if freeze {
		ebitenutil.DrawRect(screen, 0, 0, 400, 400, color.RGBA{0xFF, 0x00, 0x00, 0xFF})
	}

	if !g.started {
		ebitenutil.DebugPrint(screen, intro)
		return
	}

	if g.completed {
		ebitenutil.DebugPrint(screen, end+fmt.Sprint(finish.Sub(start).Seconds())+" seconds")
		return
	}

	mapg := ebiten.GeoM{}
	mapg.Translate(184, 184)
	mapg.Translate(-float64(g.player.x*32), -float64(g.player.y*32))
	screen.DrawImage(g.s.board, &ebiten.DrawImageOptions{GeoM: mapg})

	for _, o := range g.objs {
		if o.x == -1 {
			continue
		}
		objg := ebiten.GeoM{}
		rotatePlayer(&objg, o.o)
		objg.Translate(184, 184)
		objg.Translate(-float64(g.player.x*32), -float64(g.player.y*32))
		objg.Translate(float64(32*o.x), float64(32*o.y))
		var img *ebiten.Image
		switch o.t {
		case "g":
			img = wallImg()
		case "l":
			img = elImg()
		case "r":
			img = erImg()
		case "b":
			img = ebImg()
		}
		screen.DrawImage(img, &ebiten.DrawImageOptions{GeoM: objg})
	}

	exitg := mapg
	exitg.Translate(float64(32*g.exit.x), float64(32*g.exit.y))
	screen.DrawImage(exitImg(), &ebiten.DrawImageOptions{GeoM: exitg})

	geom := ebiten.GeoM{}
	rotatePlayer(&geom, g.player.o)
	geom.Translate(184, 184)
	if g.hold == nil {
		screen.DrawImage(playerImg(), &ebiten.DrawImageOptions{GeoM: geom})
	} else {
		screen.DrawImage(playerHImg(), &ebiten.DrawImageOptions{GeoM: geom})
	}

	ebitenutil.DebugPrint(screen, g.s.name)

	if freeze {
		ebitenutil.DebugPrint(screen, `
They've got you!
Press <enter> to respawn!`)
	}
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 400, 400
}

func main() {
	ebiten.SetWindowSize(800, 800)
	ebiten.SetWindowTitle("GGJ 2021 - MAZE")
	if err := ebiten.RunGame(&game{}); err != nil {
		log.Fatal(err)
	}
}

var player *ebiten.Image

func playerImg() *ebiten.Image {
	if player != nil {
		return player
	}

	player = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(player, 8, 8, 16, 16, color.RGBA{0xDD, 0x34, 0x10, 0xFF})
	ebitenutil.DrawRect(player, 10, 8, 12, 4, color.RGBA{0xDD, 0xDD, 0xDD, 0xFF})

	return player
}

var playerH *ebiten.Image

func playerHImg() *ebiten.Image {
	if playerH != nil {
		return playerH
	}

	playerH = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(playerH, 8, 8, 16, 16, color.RGBA{0xDD, 0x34, 0x10, 0xFF})
	ebitenutil.DrawRect(playerH, 10, 8, 12, 4, color.RGBA{0xDD, 0xDD, 0xDD, 0xFF})
	ebitenutil.DrawRect(playerH, 2, 5, 28, 3, color.RGBA{0x00, 0xDF, 0x00, 0xFF})

	return playerH
}

var er *ebiten.Image

func erImg() *ebiten.Image {
	if er != nil {
		return er
	}

	er = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(er, 8, 8, 16, 16, color.RGBA{0x10, 0xF0, 0x10, 0xFF})
	ebitenutil.DrawRect(er, 10, 8, 12, 4, color.RGBA{0x83, 0x83, 0x83, 0xFF})

	return er
}

var el *ebiten.Image

func elImg() *ebiten.Image {
	if el != nil {
		return el
	}

	el = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(el, 8, 8, 16, 16, color.RGBA{0xCA, 0x10, 0xCA, 0xFF})
	ebitenutil.DrawRect(el, 10, 8, 12, 4, color.RGBA{0x83, 0x83, 0x83, 0xFF})

	return el
}

var eb *ebiten.Image

func ebImg() *ebiten.Image {
	if eb != nil {
		return eb
	}

	eb = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(eb, 8, 8, 16, 16, color.RGBA{0x10, 0x34, 0xDD, 0xFF})
	ebitenutil.DrawRect(eb, 10, 8, 12, 4, color.RGBA{0x83, 0x83, 0x83, 0xFF})

	return eb
}

var exit *ebiten.Image

func exitImg() *ebiten.Image {
	if exit != nil {
		return exit
	}

	exit = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(exit, 6, 6, 20, 20, color.RGBA{0x00, 0x00, 0x00, 0xFF})
	ebitenutil.DrawRect(exit, 8, 8, 16, 16, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	ebitenutil.DrawRect(exit, 10, 10, 12, 12, color.RGBA{0x00, 0x00, 0x00, 0xFF})
	ebitenutil.DrawRect(exit, 12, 12, 8, 8, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	ebitenutil.DrawRect(exit, 14, 14, 4, 4, color.RGBA{0x00, 0x00, 0x00, 0xFF})

	return exit
}

var wall *ebiten.Image

func wallImg() *ebiten.Image {
	if wall != nil {
		return wall
	}

	wall = ebiten.NewImage(32, 32)
	ebitenutil.DrawRect(wall, 0, 0, 32, 4, color.RGBA{0x00, 0xDF, 0x00, 0xFF})
	ebitenutil.DrawRect(wall, 0, 0, 32, 1, color.RGBA{0xDF, 0x00, 0x00, 0xFF})

	return wall
}

func rotatePlayer(geom *ebiten.GeoM, o string) {
	switch o {
	case "u":
		return
	case "d":
		geom.Rotate(math.Pi)
		geom.Translate(32, 32)
	case "l":
		geom.Rotate(math.Pi * 1.5)
		geom.Translate(0, 32)
	case "r":
		geom.Rotate(math.Pi * 0.5)
		geom.Translate(32, 0)
	}
}

const intro = `Welcome to MAZE!

In this game you're lost multi level maze and
the only way out is going upstairs.
But be careful, as much lost as you're,
you do not want to be found!

Press <space> to hold a blue wall and
press again to release it in a valid space.
You can only hold a wall if it is in your side of the room.
You cannot go upstair holding a wall.

Use the arrow keys to move around the map.
To survive, do not let the guards be able to see you!

You can use <esc> to reset a level

Press <enter> to begin!



Made by vitor @ (silva.moe)
using GoLang + Ebiten
`

const end = `
Congratulations, you escaped the MAZE!

Press <enter> to reset the game.

\(^.^)/

It just took you: `
