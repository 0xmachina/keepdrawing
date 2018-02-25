package main

// TODO: Split this into several packages.

import (
	"log"
	"strings"
	"fmt"
	"os"
	"sort"
	
	gc "github.com/rthornton128/goncurses"
)

// Our basic map tiles.
const (
	MAP_NONE = ' '
	MAP_ROOM = '.'
	MAP_DOOR_CLOSED = '#'
	MAP_DOOR_OPEN   = '='
	MAP_STAIRS_UP   = '^'
	MAP_STAIRS_DOWN = 'v'
)

// One map (or level, if you will).
type Map struct {
	Level int
	Map [][]byte
	Height int
	Width int
}

// Our main map container, which also contains where to draw the maps.
// TODO: Perhaps consider splitting the model from the view.
type MapContainer struct {
	points map[rune]string // These are our points of interest, although
	                       // not implemented yet.
	maps map[int]Map       // Our levels, but there is no way to change level
	                       // right now.
	curlevel int           // Our current level
	win *gc.Window         // The ncurses window we can draw in
	height int             // Height of our ncurses window
	width int              // Width of our ncurses window
	cy int                 // Cursor position
	cx int
	drawMode bool          // Whether we are in room drawing mode
}

// The map sort is a different container tha the map[] above, to allow for
// sorting when we are writing the container to a file.
type MapSort []Map

func (ms MapSort) Len() int           { return len(ms) }
func (ms MapSort) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }
func (ms MapSort) Less(i, j int) bool { return ms[i].Level < ms[j].Level }

func (mc *MapContainer) WriteToFile(filename string) {
	// We overwrite whatever was there.
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("save: ", err)
	}
	defer f.Close()
	
	// First we write the points of interest
	for p, s := range mc.points {
		_, err := f.WriteString(fmt.Sprintf("%s: %s\n", p, s))
		if err != nil {
			log.Fatal("save: ", err)
		}
	}
	f.Write([]byte("\n"))
	f.Sync()
	
	var m MapSort
	for _, mp := range mc.maps {
		m = append(m, mp)
	}
	sort.Sort(m)
	
	for _, mp := range m {
		// Each level has its number above
		// Follow by its width and height, so when the map is read again
		// the system will know how much to read.
		f.WriteString(fmt.Sprintf("Level %d\n", mp.Level))
		f.WriteString(fmt.Sprintf("%d × %d\n", mp.Width, mp.Height))
		for _, line := range mp.Map {
			f.Write(line)
			f.Write([]byte("\n"))
		}
		f.Write([]byte("\n"))
	}
}

func (mc *MapContainer) DrawMap() {
	mc.win.Clear()
	mc.win.Box(0, 0)
	mp := mc.maps[mc.curlevel].Map
	for i := 0; i < len(mp); i++ {
		mc.win.MovePrint(i+1, 1, string(mp[i]))
	}
	mc.win.AttrOn(gc.A_REVERSE)
	mc.win.MovePrint(mc.cy+1, mc.cx+1, string(mp[mc.cy][mc.cx]))
	mc.win.AttrOff(gc.A_REVERSE)
	mc.win.Refresh()
}

func (mc *MapContainer) MoveCursor(cy, cx int) (int, int) {
	if cy < 0 || cy >= mc.height ||
		cx < 0 || cx >= mc.width {
		return mc.cy, mc.cx // outside map, don't do anything
	}
	mc.cy = cy
	mc.cx = cx
	if mc.drawMode {
		mc.maps[mc.curlevel].Map[mc.cy][mc.cx] = MAP_ROOM
	}
	mc.DrawMap()
	return mc.cy, mc.cx
}

func (mc *MapContainer) CursorPosition() (int, int) {
	return mc.cy, mc.cx
}

func (mc *MapContainer) SetDrawMode() {
	mc.drawMode = !mc.drawMode
}

func (mc *MapContainer) DrawMode() bool {
	return mc.drawMode
}

func (mc *MapContainer) CurrentLevel() int {
	return mc.curlevel
}

func (mc *MapContainer) PlaceStairs() {
	var c byte
	mp := mc.maps[mc.curlevel].Map
	// Allow one to flip the stairs and remove again
	switch mp[mc.cy][mc.cx] {
	case MAP_STAIRS_DOWN:
		c = MAP_STAIRS_UP
	case MAP_STAIRS_UP:
		c = MAP_ROOM
	default:
		c = MAP_STAIRS_DOWN
	}
	mc.maps[mc.curlevel].Map[mc.cy][mc.cx] = c
	mc.DrawMap()
}

func (mc *MapContainer) PlaceDoor() {
	var c byte
	mp := mc.maps[mc.curlevel].Map
	switch mp[mc.cy][mc.cx] {
	case MAP_DOOR_CLOSED:
		c = MAP_DOOR_OPEN
	case MAP_DOOR_OPEN:
		c = MAP_ROOM
	default:
		c = MAP_DOOR_CLOSED
	}
	mc.maps[mc.curlevel].Map[mc.cy][mc.cx] = c
	mc.DrawMap()
}

func NewMapContainer(height, width int, win *gc.Window) *MapContainer {
	var maps map[int]Map
	maps = make(map[int]Map, 1)
	var curlevelmap [][]byte
	curlevelmap = make([][]byte, height)
	for i := 0; i < height; i++ {
		curlevelmap[i] = []byte(strings.Repeat(" ", width))
	}
	curlevel := 1
	maps[curlevel] = Map{
		curlevel,
		curlevelmap,
		height,
		width,
	}
	
	return &MapContainer{
		make(map[rune]string),
		maps,
		curlevel,
		win,
		height,
		width,
		0,
		0,
		false,
	}
}

type MenuItem struct {
	Key rune
	Text string
	PosX int
	Len int
}

func main() {
	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal("init:", err)
	}
	defer gc.End()
	
	// Set up
	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)
	stdscr.Keypad(true)
	
	rows, cols := stdscr.MaxYX()
	height := rows - 4
	width := cols
	
	win, err := gc.NewWindow(height, width, 2, 0)
	if err != nil {
		log.Fatal("new window:", err)
	}
	defer win.Delete()
	win.Keypad(true)
	
	m := NewMapContainer(height - 2, width - 2, win)
	
	// current position
	cy, cx := m.CursorPosition()
	
	menu := []MenuItem{
		MenuItem{
			'r',
			"Room draw mode",
			0, 0,
		},
		MenuItem{
			'd',
			"Door",
			0, 0,
		},
		MenuItem{
			't',
			"Stairs",
			0, 0,
		},
		MenuItem{
			'p',
			"Point of interest",
			0, 0,
		},
		MenuItem{
			's',
			"Save",
			0, 0,
		},
	}
	
	x := 0
	for i, m := range menu {
		t := fmt.Sprintf("%s = %s", string(m.Key), m.Text)
		stdscr.MovePrint(0, x, t)
		menu[i].PosX = x
		menu[i].Len = len(t)
		x += len(t) + 2
	}
	stdscr.MovePrint(1, 0, fmt.Sprintf("Level %d", m.CurrentLevel()))
	stdscr.Refresh()
	
	if gc.MouseOk() {
		stdscr.MovePrint(1, width-18, "No mouse support")
	}
	
	gc.MouseInterval(50)
	
	// We want to detect when the left mouse button is clicked and released
	// and also whether the mouse is moving
	gc.MouseMask(gc.M_B1_PRESSED | gc.M_B1_RELEASED | gc.M_B1_CLICKED | gc.M_POSITION, nil)

	mMoving := false
	m.DrawMap()
main:
	for {
		key := win.GetChar()
		switch key {
		case 'q':
			// quit
			break main
		case 's':
			// save
			m.WriteToFile("keep.map") // TODO: Allow the user to input a filename
		case 'r':
			// enter/leave room drawing mode
			mi := menu[0]
			if !m.DrawMode() {
				stdscr.AttrOn(gc.A_REVERSE)
			}
			stdscr.MovePrint(0, mi.PosX, fmt.Sprintf("%s = %s", string(mi.Key), mi.Text))
			if !m.DrawMode() {
				stdscr.AttrOff(gc.A_REVERSE)
			}
			stdscr.Refresh()
			m.SetDrawMode()
		case 't':
			// place stairs
			m.PlaceStairs()
		case 'd':
			// place door
			m.PlaceDoor()
		case gc.KEY_UP, gc.KEY_DOWN, gc.KEY_LEFT, gc.KEY_RIGHT:
			// move cursor
			switch key {
			case gc.KEY_UP:
				cy--
			case gc.KEY_DOWN:
				cy++
			case gc.KEY_LEFT:
				cx--
			case gc.KEY_RIGHT:
				cx++
			}
			cy, cx = m.MoveCursor(cy, cx)
		case gc.KEY_MOUSE:
			if md := gc.GetMouse(); md != nil {
				if md.State == gc.M_B1_PRESSED {
					mMoving = true
				}
				if md.State == gc.M_B1_RELEASED && mMoving {
					mMoving = false
				}
				if mMoving {
					cy, cx = md.Y - 2, md.X
					cy, cx = m.MoveCursor(cy-1, cx-1)
				}
			}
		}
	}
}

