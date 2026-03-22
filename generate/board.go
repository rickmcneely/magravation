package generate

import (
	"fmt"
	"math"
)

// HoleType identifies the purpose of a hole on the board.
type HoleType int

const (
	HoleTrack   HoleType = iota // main playing track
	HoleBase                    // starting base (5 per player)
	HoleHomeRow                 // safe path to center (4 per player)
	HoleCenter                  // center hole
	HoleStart                   // start/entry position on track (marked)
)

func (h HoleType) String() string {
	switch h {
	case HoleTrack:
		return "track"
	case HoleBase:
		return "base"
	case HoleHomeRow:
		return "homerow"
	case HoleCenter:
		return "center"
	case HoleStart:
		return "start"
	default:
		return "unknown"
	}
}

// Hole represents a single marble hole on the board.
type Hole struct {
	X, Y     float64  // physical position (inches), origin at board center
	Type     HoleType // hole purpose
	Player   int      // player index (0-based), -1 for shared
	Diameter float64  // cutting diameter
	Depth    float64  // cutting depth
}

// TextItem represents text to be engraved on the board.
type TextItem struct {
	X, Y     float64
	Text     string
	Height   float64
	Angle    float64
	CenterOn bool
}

// Board holds the complete board layout.
type Board struct {
	Params    Params
	Holes     []Hole
	TextItems []TextItem
}

// GenerateBoard creates the full board layout from parameters.
func GenerateBoard(p Params) (*Board, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	b := &Board{Params: p}

	if p.NumPlayers == 4 {
		b.generate4Player()
	} else {
		b.generate6Player()
	}

	return b, nil
}

// gridToPhys converts a grid position (col, row) on the 15×15 board map
// to physical coordinates with origin at board center.
// Grid center is at (7,7). Y is flipped (row 0 = top = +Y).
func gridToPhys(col, row int, d float64) (float64, float64) {
	x := float64(col-7) * d
	y := float64(7-row) * d
	return x, y
}

// rotatePoint rotates (x, y) by angleDeg degrees CCW around the origin.
func rotatePoint(x, y, angleDeg float64) (float64, float64) {
	rad := angleDeg * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	return x*cos - y*sin, x*sin + y*cos
}

func (b *Board) addHole(x, y float64, htype HoleType, player int) {
	b.Holes = append(b.Holes, Hole{
		X: x, Y: y,
		Type:     htype,
		Player:   player,
		Diameter: b.Params.HoleDiameter(),
		Depth:    b.Params.HoleDepth(),
	})
}

// ─────────────────────────────────────────────────────────────
// 4-Player Board — canonical layout from the character map
// ─────────────────────────────────────────────────────────────
//
// The board is a 15×15 character grid, center at (7,7).
// Each cell is GridSpacing() apart.  89 marble positions total.
//
//   X    XXXXX    X      row  0
//    X   X X X   X       row  1
//     X  X X X  X        row  2
//      X X X X X         row  3
//        X X X           row  4
//   XXXXXX   XXXXXX      row  5
//   X             X      row  6
//   XXXXX  X  XXXXX      row  7
//   X             X      row  8
//   XXXXXX   XXXXXX      row  9
//        X X X           row 10
//      X X X X X         row 11
//     X  X X X  X        row 12
//    X   X X X   X       row 13
//   X    XXXXX    X      row 14
//
// Station structure (one of four, this one pointing UP):
//   Base (5):     top row XXXXX at cols 5-9
//   Home row (4): center column x=7, rows 1-4
//   Left track:   column x=5, rows 1-4
//   Right track:  column x=9, rows 1-4 (row 1 = start position)
//   Diagonals:    4 holes at 45° connecting to adjacent stations

func (b *Board) generate4Player() {
	d := b.Params.GridSpacing()

	type hpos struct {
		col, row int
		htype    HoleType
		player   int
	}

	positions := []hpos{
		// ── Center ──
		{7, 7, HoleCenter, -1},

		// ── Center connectors (inner corners of the cross) ──
		{5, 5, HoleTrack, -1},
		{9, 5, HoleTrack, -1},
		{5, 9, HoleTrack, -1},
		{9, 9, HoleTrack, -1},

		// ── Top station (player 0) ──
		// Base (5 across the top)
		{5, 0, HoleBase, 0}, {6, 0, HoleBase, 0}, {7, 0, HoleBase, 0},
		{8, 0, HoleBase, 0}, {9, 0, HoleBase, 0},
		// Home row (center column, rows 1-4)
		{7, 1, HoleHomeRow, 0}, {7, 2, HoleHomeRow, 0},
		{7, 3, HoleHomeRow, 0}, {7, 4, HoleHomeRow, 0},
		// Left track column (rows 1-4)
		{5, 1, HoleTrack, -1}, {5, 2, HoleTrack, -1},
		{5, 3, HoleTrack, -1}, {5, 4, HoleTrack, -1},
		// Right track column (row 1 = start, rows 2-4 = track)
		{9, 1, HoleStart, 0},
		{9, 2, HoleTrack, -1}, {9, 3, HoleTrack, -1}, {9, 4, HoleTrack, -1},

		// ── Right station (player 1) ──
		// Base (5 down the right edge)
		{14, 5, HoleBase, 1}, {14, 6, HoleBase, 1}, {14, 7, HoleBase, 1},
		{14, 8, HoleBase, 1}, {14, 9, HoleBase, 1},
		// Home row (center row y=7, cols 10-13)
		{13, 7, HoleHomeRow, 1}, {12, 7, HoleHomeRow, 1},
		{11, 7, HoleHomeRow, 1}, {10, 7, HoleHomeRow, 1},
		// Top track row (y=5, cols 10-13)
		{13, 5, HoleTrack, -1}, {12, 5, HoleTrack, -1},
		{11, 5, HoleTrack, -1}, {10, 5, HoleTrack, -1},
		// Bottom track row (col 13 = start, cols 10-12 = track)
		{13, 9, HoleStart, 1},
		{12, 9, HoleTrack, -1}, {11, 9, HoleTrack, -1}, {10, 9, HoleTrack, -1},

		// ── Bottom station (player 2) ──
		// Base (5 across the bottom)
		{9, 14, HoleBase, 2}, {8, 14, HoleBase, 2}, {7, 14, HoleBase, 2},
		{6, 14, HoleBase, 2}, {5, 14, HoleBase, 2},
		// Home row (center column, rows 10-13)
		{7, 13, HoleHomeRow, 2}, {7, 12, HoleHomeRow, 2},
		{7, 11, HoleHomeRow, 2}, {7, 10, HoleHomeRow, 2},
		// Right track column (rows 10-13)
		{9, 13, HoleTrack, -1}, {9, 12, HoleTrack, -1},
		{9, 11, HoleTrack, -1}, {9, 10, HoleTrack, -1},
		// Left track column (row 13 = start, rows 10-12 = track)
		{5, 13, HoleStart, 2},
		{5, 12, HoleTrack, -1}, {5, 11, HoleTrack, -1}, {5, 10, HoleTrack, -1},

		// ── Left station (player 3) ──
		// Base (5 down the left edge)
		{0, 9, HoleBase, 3}, {0, 8, HoleBase, 3}, {0, 7, HoleBase, 3},
		{0, 6, HoleBase, 3}, {0, 5, HoleBase, 3},
		// Home row (center row y=7, cols 1-4)
		{1, 7, HoleHomeRow, 3}, {2, 7, HoleHomeRow, 3},
		{3, 7, HoleHomeRow, 3}, {4, 7, HoleHomeRow, 3},
		// Bottom track row (y=9, cols 1-4)
		{1, 9, HoleTrack, -1}, {2, 9, HoleTrack, -1},
		{3, 9, HoleTrack, -1}, {4, 9, HoleTrack, -1},
		// Top track row (col 1 = start, cols 2-4 = track)
		{1, 5, HoleStart, 3},
		{2, 5, HoleTrack, -1}, {3, 5, HoleTrack, -1}, {4, 5, HoleTrack, -1},

		// ── Diagonal connectors (4 holes each, linking adjacent stations) ──
		// Top-left diagonal (top station ↔ left station)
		{3, 3, HoleTrack, -1}, {2, 2, HoleTrack, -1},
		{1, 1, HoleTrack, -1}, {0, 0, HoleTrack, -1},
		// Top-right diagonal (top station ↔ right station)
		{11, 3, HoleTrack, -1}, {12, 2, HoleTrack, -1},
		{13, 1, HoleTrack, -1}, {14, 0, HoleTrack, -1},
		// Bottom-right diagonal (right station ↔ bottom station)
		{11, 11, HoleTrack, -1}, {12, 12, HoleTrack, -1},
		{13, 13, HoleTrack, -1}, {14, 14, HoleTrack, -1},
		// Bottom-left diagonal (bottom station ↔ left station)
		{3, 11, HoleTrack, -1}, {2, 12, HoleTrack, -1},
		{1, 13, HoleTrack, -1}, {0, 14, HoleTrack, -1},
	}

	for _, p := range positions {
		x, y := gridToPhys(p.col, p.row, d)
		b.addHole(x, y, p.htype, p.player)
	}

	// ── Text ──
	r := b.Params.BoardDiameter/2 - b.Params.EdgeMargin()*0.6
	b.TextItems = append(b.TextItems,
		TextItem{X: 0, Y: r, Text: "AGGRAVATION", Height: b.Params.TextHeight, CenterOn: true},
		TextItem{X: 0, Y: -r, Text: "4 PLAYER", Height: b.Params.TextHeight * 0.5, CenterOn: true},
	)

	// Player labels near bases
	labels := []struct {
		col, row int
		text     string
	}{
		{7, -1, "P1"}, // above top base
		{16, 7, "P2"}, // right of right base
		{7, 16, "P3"}, // below bottom base
		{-2, 7, "P4"}, // left of left base
	}
	for _, lbl := range labels {
		lx, ly := gridToPhys(lbl.col, lbl.row, d)
		b.TextItems = append(b.TextItems, TextItem{
			X: lx, Y: ly, Text: lbl.text,
			Height: b.Params.TextHeight * 0.4, CenterOn: true,
		})
	}
}

// ─────────────────────────────────────────────────────────────
// 6-Player Board — same station structure rotated at 60° intervals
// ─────────────────────────────────────────────────────────────
// Uses the same station template (base + 3 columns + 4 rows)
// rotated for each of 6 arms. Diagonal connectors at 30° angles
// with 2 holes each (shorter path between 60°-separated arms).
// Center + 6 connectors.

func (b *Board) generate6Player() {
	d := b.Params.GridSpacing()

	// Center
	b.addHole(0, 0, HoleCenter, -1)

	// Arm angles (pointing outward from center)
	armAngles := []float64{90, 30, -30, -90, -150, 150}

	// Station template positions in arm-local coordinates.
	// Arm points along +Y. Units are grid cells (multiply by d).
	// Base is at y=7 (farthest from center), inner track at y=3.

	type tpos struct {
		x, y  float64
		htype HoleType
		isStart bool
	}

	// Base (5 holes across)
	base := []tpos{
		{-2, 7, HoleBase, false}, {-1, 7, HoleBase, false}, {0, 7, HoleBase, false},
		{1, 7, HoleBase, false}, {2, 7, HoleBase, false},
	}
	// Home row (center column, 4 rows)
	home := []tpos{
		{0, 6, HoleHomeRow, false}, {0, 5, HoleHomeRow, false},
		{0, 4, HoleHomeRow, false}, {0, 3, HoleHomeRow, false},
	}
	// Left track column (4 rows)
	leftTrack := []tpos{
		{-2, 6, HoleTrack, false}, {-2, 5, HoleTrack, false},
		{-2, 4, HoleTrack, false}, {-2, 3, HoleTrack, false},
	}
	// Right track column (row 6 = start, rows 3-5 = track)
	rightTrack := []tpos{
		{2, 6, HoleTrack, true}, // start position
		{2, 5, HoleTrack, false}, {2, 4, HoleTrack, false}, {2, 3, HoleTrack, false},
	}

	// All station template positions
	stationTemplate := make([]tpos, 0, 17)
	stationTemplate = append(stationTemplate, base...)
	stationTemplate = append(stationTemplate, home...)
	stationTemplate = append(stationTemplate, leftTrack...)
	stationTemplate = append(stationTemplate, rightTrack...)

	for player, angle := range armAngles {
		for _, tp := range stationTemplate {
			rx, ry := rotatePoint(tp.x*d, tp.y*d, angle-90)
			ht := tp.htype
			pl := player
			if tp.htype == HoleBase {
				pl = player
			} else if tp.htype == HoleHomeRow {
				pl = player
			} else {
				pl = -1
			}
			if tp.isStart {
				ht = HoleStart
				pl = player
			}
			b.addHole(rx, ry, ht, pl)
		}
	}

	// Center connectors: 6 positions, one between each pair of adjacent arms.
	// Each connector is at the midpoint angle, at radius = 2√2 × d from center
	// (matching the 4-player center connectors at grid distance √(2²+2²)=2√2).
	connR := 2 * math.Sqrt(2) * d
	for i := 0; i < 6; i++ {
		midAngle := (armAngles[i] + armAngles[(i+1)%6]) / 2
		// Handle wrap-around for last pair
		if i == 5 {
			midAngle = (armAngles[5] + armAngles[0] + 360) / 2
			if midAngle >= 360 {
				midAngle -= 360
			}
		}
		cx := connR * math.Cos(midAngle*math.Pi/180)
		cy := connR * math.Sin(midAngle*math.Pi/180)
		b.addHole(cx, cy, HoleTrack, -1)
	}

	// Diagonal connectors between adjacent stations (2 holes each).
	// Each diagonal goes from one arm's inner-left track to the next arm's
	// inner-right track, at the bisecting angle.
	for i := 0; i < 6; i++ {
		a0 := armAngles[i] - 90
		a1 := armAngles[(i+1)%6] - 90

		// Arm i left-track inner position: (-2, 3) rotated
		lx0, ly0 := rotatePoint(-2*d, 3*d, a0)
		// Arm i+1 right-track inner position: (2, 3) rotated
		rx1, ry1 := rotatePoint(2*d, 3*d, a1)

		// Two evenly-spaced diagonal holes between them
		for k := 1; k <= 2; k++ {
			t := float64(k) / 3.0
			hx := lx0 + t*(rx1-lx0)
			hy := ly0 + t*(ry1-ly0)
			b.addHole(hx, hy, HoleTrack, -1)
		}
	}

	// Text
	r := b.Params.BoardDiameter/2 - b.Params.EdgeMargin()*0.6
	b.TextItems = append(b.TextItems,
		TextItem{X: 0, Y: r, Text: "AGGRAVATION", Height: b.Params.TextHeight, CenterOn: true},
		TextItem{X: 0, Y: -r, Text: "6 PLAYER", Height: b.Params.TextHeight * 0.5, CenterOn: true},
	)

	for player, angle := range armAngles {
		lx, ly := rotatePoint(0, 8*d, angle-90+90)
		b.TextItems = append(b.TextItems, TextItem{
			X: lx, Y: ly, Text: fmt.Sprintf("P%d", player+1),
			Height: b.Params.TextHeight * 0.35, CenterOn: true,
		})
	}
}

// Bounds returns the min/max X,Y extents of all features.
func (b *Board) Bounds() (minX, minY, maxX, maxY float64) {
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64
	for _, h := range b.Holes {
		r := h.Diameter / 2
		if h.X-r < minX {
			minX = h.X - r
		}
		if h.Y-r < minY {
			minY = h.Y - r
		}
		if h.X+r > maxX {
			maxX = h.X + r
		}
		if h.Y+r > maxY {
			maxY = h.Y + r
		}
	}
	return
}
