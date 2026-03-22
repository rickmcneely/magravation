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
	HoleStart                   // start/entry position on track
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
	X, Y     float64
	Type     HoleType
	Player   int // 0-based, -1 for shared
	Diameter float64
	Depth    float64
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

// ─────────────────────────────────────────────────────────────
// 4-Player Board — canonical 15×15 grid layout (89 holes)
// ─────────────────────────────────────────────────────────────

func (b *Board) generate4Player() {
	d := b.Params.GridSpacing()

	type hpos struct {
		col, row int
		htype    HoleType
		player   int
	}

	positions := []hpos{
		// Center
		{7, 7, HoleCenter, -1},

		// Center connectors (inner corners of the cross)
		{5, 5, HoleTrack, -1}, {9, 5, HoleTrack, -1},
		{5, 9, HoleTrack, -1}, {9, 9, HoleTrack, -1},

		// ── Top station (player 0) ──
		{5, 0, HoleBase, 0}, {6, 0, HoleBase, 0}, {7, 0, HoleBase, 0},
		{8, 0, HoleBase, 0}, {9, 0, HoleBase, 0},
		{7, 1, HoleHomeRow, 0}, {7, 2, HoleHomeRow, 0},
		{7, 3, HoleHomeRow, 0}, {7, 4, HoleHomeRow, 0},
		{5, 1, HoleTrack, -1}, {5, 2, HoleTrack, -1},
		{5, 3, HoleTrack, -1}, {5, 4, HoleTrack, -1},
		{9, 1, HoleStart, 0},
		{9, 2, HoleTrack, -1}, {9, 3, HoleTrack, -1}, {9, 4, HoleTrack, -1},

		// ── Right station (player 1) ──
		{14, 5, HoleBase, 1}, {14, 6, HoleBase, 1}, {14, 7, HoleBase, 1},
		{14, 8, HoleBase, 1}, {14, 9, HoleBase, 1},
		{13, 7, HoleHomeRow, 1}, {12, 7, HoleHomeRow, 1},
		{11, 7, HoleHomeRow, 1}, {10, 7, HoleHomeRow, 1},
		{13, 5, HoleTrack, -1}, {12, 5, HoleTrack, -1},
		{11, 5, HoleTrack, -1}, {10, 5, HoleTrack, -1},
		{13, 9, HoleStart, 1},
		{12, 9, HoleTrack, -1}, {11, 9, HoleTrack, -1}, {10, 9, HoleTrack, -1},

		// ── Bottom station (player 2) ──
		{9, 14, HoleBase, 2}, {8, 14, HoleBase, 2}, {7, 14, HoleBase, 2},
		{6, 14, HoleBase, 2}, {5, 14, HoleBase, 2},
		{7, 13, HoleHomeRow, 2}, {7, 12, HoleHomeRow, 2},
		{7, 11, HoleHomeRow, 2}, {7, 10, HoleHomeRow, 2},
		{9, 13, HoleTrack, -1}, {9, 12, HoleTrack, -1},
		{9, 11, HoleTrack, -1}, {9, 10, HoleTrack, -1},
		{5, 13, HoleStart, 2},
		{5, 12, HoleTrack, -1}, {5, 11, HoleTrack, -1}, {5, 10, HoleTrack, -1},

		// ── Left station (player 3) ──
		{0, 9, HoleBase, 3}, {0, 8, HoleBase, 3}, {0, 7, HoleBase, 3},
		{0, 6, HoleBase, 3}, {0, 5, HoleBase, 3},
		{1, 7, HoleHomeRow, 3}, {2, 7, HoleHomeRow, 3},
		{3, 7, HoleHomeRow, 3}, {4, 7, HoleHomeRow, 3},
		{1, 9, HoleTrack, -1}, {2, 9, HoleTrack, -1},
		{3, 9, HoleTrack, -1}, {4, 9, HoleTrack, -1},
		{1, 5, HoleStart, 3},
		{2, 5, HoleTrack, -1}, {3, 5, HoleTrack, -1}, {4, 5, HoleTrack, -1},

		// ── Diagonal connectors ──
		{3, 3, HoleTrack, -1}, {2, 2, HoleTrack, -1},
		{1, 1, HoleTrack, -1}, {0, 0, HoleTrack, -1},
		{11, 3, HoleTrack, -1}, {12, 2, HoleTrack, -1},
		{13, 1, HoleTrack, -1}, {14, 0, HoleTrack, -1},
		{11, 11, HoleTrack, -1}, {12, 12, HoleTrack, -1},
		{13, 13, HoleTrack, -1}, {14, 14, HoleTrack, -1},
		{3, 11, HoleTrack, -1}, {2, 12, HoleTrack, -1},
		{1, 13, HoleTrack, -1}, {0, 14, HoleTrack, -1},
	}

	for _, p := range positions {
		x, y := gridToPhys(p.col, p.row, d)
		b.addHole(x, y, p.htype, p.player)
	}

	b.addBoardText(d)
}

// ─────────────────────────────────────────────────────────────
// N-Player Board (3, 5, or 6) — rotation-based layout
// ─────────────────────────────────────────────────────────────
// Same station structure as 4-player but rotated.
// Flanking track columns start at a safe radius to avoid
// collision with adjacent arms:
//   3 players (120°): r ≥ 2
//   5 players  (72°): r ≥ 4
//   6 players  (60°): r ≥ 5
// Center column (home row) is safe at all radii.
// One connector hole per gap at the midpoint between adjacent
// arm track inners.

func minFlankRadius(n int) int {
	switch n {
	case 3:
		return 2
	case 5:
		return 4
	case 6:
		return 5
	default:
		return 3
	}
}

func (b *Board) generateNPlayer() {
	n := b.Params.NumPlayers
	d := b.Params.GridSpacing()
	minR := minFlankRadius(n)

	// Arm angles (equally spaced, first arm pointing up)
	armAngles := make([]float64, n)
	for i := 0; i < n; i++ {
		armAngles[i] = 90.0 - float64(i)*360.0/float64(n)
	}

	// Center hole
	b.addHole(0, 0, HoleCenter, -1)

	// ── Station template (arm pointing along +Y) ──

	type tpos struct {
		x, y    float64
		htype   HoleType
		isStart bool
	}

	// Base: 5 holes at y=7
	base := []tpos{
		{-2, 7, HoleBase, false}, {-1, 7, HoleBase, false}, {0, 7, HoleBase, false},
		{1, 7, HoleBase, false}, {2, 7, HoleBase, false},
	}

	// Home row: center column y=3..6 (safe at all player counts)
	home := []tpos{
		{0, 6, HoleHomeRow, false}, {0, 5, HoleHomeRow, false},
		{0, 4, HoleHomeRow, false}, {0, 3, HoleHomeRow, false},
	}

	// Flanking track columns: y = minR .. 6
	var leftTrack, rightTrack []tpos
	for r := 6; r >= minR; r-- {
		leftTrack = append(leftTrack, tpos{-2, float64(r), HoleTrack, false})
	}
	// Right track: row 6 is start position
	rightTrack = append(rightTrack, tpos{2, 6, HoleTrack, true})
	for r := 5; r >= minR; r-- {
		rightTrack = append(rightTrack, tpos{2, float64(r), HoleTrack, false})
	}

	// Combine all station template positions
	station := make([]tpos, 0, len(base)+len(home)+len(leftTrack)+len(rightTrack))
	station = append(station, base...)
	station = append(station, home...)
	station = append(station, leftTrack...)
	station = append(station, rightTrack...)

	// Place each station
	for player := 0; player < n; player++ {
		rot := armAngles[player] - 90

		for _, tp := range station {
			rx, ry := rotatePoint(tp.x*d, tp.y*d, rot)
			ht := tp.htype
			pl := -1
			if tp.htype == HoleBase || tp.htype == HoleHomeRow {
				pl = player
			}
			if tp.isStart {
				ht = HoleStart
				pl = player
			}
			b.addHole(rx, ry, ht, pl)
		}
	}

	// ── Gap connector holes ──
	// One hole per gap at the midpoint between adjacent arm track inners.
	for i := 0; i < n; i++ {
		rot0 := armAngles[i] - 90
		rot1 := armAngles[(i+1)%n] - 90

		// Arm i right-track inner: (2, minR) rotated
		rx, ry := rotatePoint(2*d, float64(minR)*d, rot0)
		// Arm i+1 left-track inner: (-2, minR) rotated
		lx, ly := rotatePoint(-2*d, float64(minR)*d, rot1)

		// Midpoint connector
		b.addHole((rx+lx)/2, (ry+ly)/2, HoleTrack, -1)
	}

	b.addBoardText(d)
}

func (b *Board) addBoardText(d float64) {
	r := b.Params.BoardDiameter/2 - b.Params.EdgeMargin()*0.6
	n := b.Params.NumPlayers

	b.TextItems = append(b.TextItems,
		TextItem{X: 0, Y: r, Text: "AGGRAVATION", Height: b.Params.TextHeight, CenterOn: true},
		TextItem{X: 0, Y: -r, Text: fmt.Sprintf("%d PLAYER", n), Height: b.Params.TextHeight * 0.5, CenterOn: true},
	)

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
		armAngles := make([]float64, n)
		for i := 0; i < n; i++ {
			armAngles[i] = 90.0 - float64(i)*360.0/float64(n)
		}
		for i, angle := range armAngles {
			lx, ly := rotatePoint(0, 8*d, angle-90+90)
			b.TextItems = append(b.TextItems, TextItem{
				X: lx, Y: ly, Text: fmt.Sprintf("P%d", i+1),
				Height: b.Params.TextHeight * 0.35, CenterOn: true,
			})
		}
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
