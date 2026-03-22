package generate

import (
	"fmt"
	"math"
)

type HoleType int

const (
	HoleTrack   HoleType = iota
	HoleBase             // start positions ($,S,s) — 4 per player
	HoleHomeRow          // ending positions (E,m) — 4 per player
	HoleCenter           // center hole (0)
	HoleStart            // track entry point
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

type Hole struct {
	X, Y     float64
	Type     HoleType
	Player   int
	Diameter float64
	Depth    float64
}

type TextItem struct {
	X, Y     float64
	Text     string
	Height   float64
	Angle    float64
	CenterOn bool
}

type Board struct {
	Params    Params
	Holes     []Hole
	TextItems []TextItem
}

func GenerateBoard(p Params) (*Board, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	b := &Board{Params: p}
	if p.NumPlayers == 4 {
		b.generate4Player()
	} else {
		b.generateNPlayer()
	}
	return b, nil
}

func gridToPhys(col, row int, d float64) (float64, float64) {
	return float64(col-7) * d, float64(7-row) * d
}

func rotatePoint(x, y, angleDeg float64) (float64, float64) {
	rad := angleDeg * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	return x*cos - y*sin, x*sin + y*cos
}

func (b *Board) addHole(x, y float64, htype HoleType, player int) {
	b.Holes = append(b.Holes, Hole{
		X: x, Y: y, Type: htype, Player: player,
		Diameter: b.Params.HoleDiameter(),
		Depth:    b.Params.HoleDepth(),
	})
}

// startPositions returns 4 radial start positions for a player.
// Along the bisector angle between this arm and the CCW neighbor.
// Outermost ($) at same radius as i (baseY * d from center).
// Each subsequent marble is 1.4 * marbleDiameter closer to center.
func startPositions(armAngle float64, numPlayers int, d, marbleDiam float64) [4][2]float64 {
	bisector := armAngle + 180.0/float64(numPlayers)
	rad := bisector * math.Pi / 180
	dx, dy := math.Cos(rad), math.Sin(rad)
	baseY := StationBaseY(numPlayers)
	startSpacing := 1.4 * marbleDiam

	var pos [4][2]float64
	for i := 0; i < 4; i++ {
		r := baseY*d - float64(i)*startSpacing
		pos[i] = [2]float64{r * dx, r * dy}
	}
	return pos
}

// ─────────────────────────────────────────────────────────────
// 4-Player Board — canonical 15×15 grid (89 holes)
// ─────────────────────────────────────────────────────────────
// The 4-player grid layout exactly matches the connector circle
// formula: R = 2√2 ≈ 2.83, C at grid (5,5)(9,5)(5,9)(9,9),
// c at grid row 4, base at row 0. Kept as hardcoded grid for
// pixel-perfect positioning.

func (b *Board) generate4Player() {
	d := b.Params.GridSpacing()

	type hpos struct {
		col, row int
		htype    HoleType
		player   int
	}

	positions := []hpos{
		{7, 7, HoleCenter, -1},

		// Center connectors (C)
		{5, 5, HoleTrack, -1}, {9, 5, HoleTrack, -1},
		{5, 9, HoleTrack, -1}, {9, 9, HoleTrack, -1},

		// Top station (player 0)
		// Base row XXiXX (all track)
		{5, 0, HoleTrack, -1}, {6, 0, HoleTrack, -1}, {7, 0, HoleTrack, -1},
		{8, 0, HoleTrack, -1}, {9, 0, HoleTrack, -1},
		// Left track: X(1-3) + c(4)
		{5, 1, HoleTrack, -1}, {5, 2, HoleTrack, -1},
		{5, 3, HoleTrack, -1}, {5, 4, HoleTrack, -1},
		// Home/ending: E(1-3) + m(4)
		{7, 1, HoleHomeRow, 0}, {7, 2, HoleHomeRow, 0},
		{7, 3, HoleHomeRow, 0}, {7, 4, HoleHomeRow, 0},
		// Right track: start(1) + X(2-3) + c(4)
		{9, 1, HoleStart, 0},
		{9, 2, HoleTrack, -1}, {9, 3, HoleTrack, -1}, {9, 4, HoleTrack, -1},

		// Right station (player 1)
		{14, 5, HoleTrack, -1}, {14, 6, HoleTrack, -1}, {14, 7, HoleTrack, -1},
		{14, 8, HoleTrack, -1}, {14, 9, HoleTrack, -1},
		{13, 5, HoleTrack, -1}, {12, 5, HoleTrack, -1},
		{11, 5, HoleTrack, -1}, {10, 5, HoleTrack, -1},
		{13, 7, HoleHomeRow, 1}, {12, 7, HoleHomeRow, 1},
		{11, 7, HoleHomeRow, 1}, {10, 7, HoleHomeRow, 1},
		{13, 9, HoleStart, 1},
		{12, 9, HoleTrack, -1}, {11, 9, HoleTrack, -1}, {10, 9, HoleTrack, -1},

		// Bottom station (player 2)
		{9, 14, HoleTrack, -1}, {8, 14, HoleTrack, -1}, {7, 14, HoleTrack, -1},
		{6, 14, HoleTrack, -1}, {5, 14, HoleTrack, -1},
		{9, 13, HoleTrack, -1}, {9, 12, HoleTrack, -1},
		{9, 11, HoleTrack, -1}, {9, 10, HoleTrack, -1},
		{7, 13, HoleHomeRow, 2}, {7, 12, HoleHomeRow, 2},
		{7, 11, HoleHomeRow, 2}, {7, 10, HoleHomeRow, 2},
		{5, 13, HoleStart, 2},
		{5, 12, HoleTrack, -1}, {5, 11, HoleTrack, -1}, {5, 10, HoleTrack, -1},

		// Left station (player 3)
		{0, 9, HoleTrack, -1}, {0, 8, HoleTrack, -1}, {0, 7, HoleTrack, -1},
		{0, 6, HoleTrack, -1}, {0, 5, HoleTrack, -1},
		{1, 9, HoleTrack, -1}, {2, 9, HoleTrack, -1},
		{3, 9, HoleTrack, -1}, {4, 9, HoleTrack, -1},
		{1, 7, HoleHomeRow, 3}, {2, 7, HoleHomeRow, 3},
		{3, 7, HoleHomeRow, 3}, {4, 7, HoleHomeRow, 3},
		{1, 5, HoleStart, 3},
		{2, 5, HoleTrack, -1}, {3, 5, HoleTrack, -1}, {4, 5, HoleTrack, -1},
	}

	for _, p := range positions {
		x, y := gridToPhys(p.col, p.row, d)
		b.addHole(x, y, p.htype, p.player)
	}

	// Radial start positions (4 per player)
	armAngles := []float64{90, 0, -90, 180}
	for player, armAngle := range armAngles {
		for _, pos := range startPositions(armAngle, 4, d, b.Params.MarbleDiameter) {
			b.addHole(pos[0], pos[1], HoleBase, player)
		}
	}

	b.addBoardText(d)
}

// ─────────────────────────────────────────────────────────────
// N-Player Board (3, 5, or 6) — rotation-based layout
// ─────────────────────────────────────────────────────────────
//
// The connector circle (C positions) determines station placement:
//   R_connector = 2 / sin(π/N)          (grid cells)
//   Adjacent C chord distance = 4       (same as c-to-c in station)
//   C_y (arm-local) = 2 / tan(π/N)     (grid cells from center)
//   c_y = C_y + 1                       (1 cell outward from C)
//   base_y = c_y + 4 = C_y + 5         (4 more cells to the tip)
//
// Each station has: base row (5) + 3 track/home rows (9) + inner row c,m,c (3)
//                 = 17 positions + 4 radial starts = 21 per player.
// Plus N center connectors (C) + 1 center = N+1 shared positions.

func (b *Board) generateNPlayer() {
	n := b.Params.NumPlayers
	d := b.Params.GridSpacing()

	cLocalY := 2.0 / math.Tan(math.Pi/float64(n)) // C position y in arm-local grid
	cY := cLocalY + 1.0                             // c position y (1 cell outward)

	armAngles := make([]float64, n)
	for i := 0; i < n; i++ {
		armAngles[i] = 90.0 - float64(i)*360.0/float64(n)
	}

	// Center
	b.addHole(0, 0, HoleCenter, -1)

	// Center connectors (C) on the connector circle
	connR := ConnectorCircleRadius(n) * d
	for i := 0; i < n; i++ {
		midAngle := armAngles[i] + 180.0/float64(n) // bisector toward CCW neighbor
		cx := connR * math.Cos(midAngle*math.Pi/180)
		cy := connR * math.Sin(midAngle*math.Pi/180)
		b.addHole(cx, cy, HoleTrack, -1)
	}

	// Station template (arm pointing +Y, all y values in grid cells)
	// Row 0 (inner): c, m, c  at y = cY
	// Row 1:         X, E, X  at y = cY+1
	// Row 2:         X, E, X  at y = cY+2
	// Row 3:         X, E, X  at y = cY+3
	// Row 4 (base):  XXiXX    at y = cY+4

	type tpos struct {
		x, y  float64
		htype HoleType
		owned bool // player-owned (home) vs shared (track)
		start bool
	}

	var station []tpos

	// Inner row: c, m, c
	station = append(station,
		tpos{-2, cY, HoleTrack, false, false},     // c left
		tpos{0, cY, HoleHomeRow, true, false},      // m
		tpos{2, cY, HoleTrack, false, false},       // c right
	)

	// 3 middle rows: X, E, X
	for row := 1; row <= 3; row++ {
		y := cY + float64(row)
		station = append(station,
			tpos{-2, y, HoleTrack, false, false},   // X left
			tpos{0, y, HoleHomeRow, true, false},    // E
			tpos{2, y, HoleTrack, false, false},     // X right
		)
	}
	// Mark right track row 1 as start position
	station[4].start = true // (2, cY+1) = right track, first row above c

	// Base row: X X i X X
	baseY := cY + 4.0
	station = append(station,
		tpos{-2, baseY, HoleTrack, false, false},
		tpos{-1, baseY, HoleTrack, false, false},
		tpos{0, baseY, HoleTrack, false, false},    // i (intersection)
		tpos{1, baseY, HoleTrack, false, false},
		tpos{2, baseY, HoleTrack, false, false},
	)

	// Place each station
	for player := 0; player < n; player++ {
		rot := armAngles[player] - 90

		for _, tp := range station {
			rx, ry := rotatePoint(tp.x*d, tp.y*d, rot)
			ht := tp.htype
			pl := -1
			if tp.owned {
				pl = player
			}
			if tp.start {
				ht = HoleStart
				pl = player
			}
			b.addHole(rx, ry, ht, pl)
		}

		// Radial start positions
		for _, pos := range startPositions(armAngles[player], n, d, b.Params.MarbleDiameter) {
			b.addHole(pos[0], pos[1], HoleBase, player)
		}
	}

	b.addBoardText(d)
}

func (b *Board) addBoardText(d float64) {
	r := b.Params.BoardDiameter/2 - b.Params.EdgeMargin()*0.6
	n := b.Params.NumPlayers

	b.TextItems = append(b.TextItems,
		TextItem{X: 0, Y: r, Text: "WAHOO!", Height: b.Params.TextHeight, CenterOn: true},
		TextItem{X: 0, Y: -r, Text: fmt.Sprintf("%d PLAYER", n), Height: b.Params.TextHeight * 0.5, CenterOn: true},
	)

	armAngles := make([]float64, n)
	for i := 0; i < n; i++ {
		armAngles[i] = 90.0 - float64(i)*360.0/float64(n)
	}

	if n == 4 {
		labels := [][2]int{{7, -1}, {16, 7}, {7, 16}, {-2, 7}}
		for i, lbl := range labels {
			lx, ly := gridToPhys(lbl[0], lbl[1], d)
			b.TextItems = append(b.TextItems, TextItem{
				X: lx, Y: ly, Text: fmt.Sprintf("P%d", i+1),
				Height: b.Params.TextHeight * 0.4, CenterOn: true,
			})
		}
	} else {
		baseY := StationBaseY(n)
		for i, angle := range armAngles {
			rad := angle * math.Pi / 180
			r := (baseY + 1) * d
			b.TextItems = append(b.TextItems, TextItem{
				X: r * math.Cos(rad), Y: r * math.Sin(rad),
				Text: fmt.Sprintf("P%d", i+1),
				Height: b.Params.TextHeight * 0.35, CenterOn: true,
			})
		}
	}
}

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
