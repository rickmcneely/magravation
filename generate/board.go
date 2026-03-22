package generate

import (
	"fmt"
	"math"
)

// HoleType identifies the purpose of a hole on the board.
type HoleType int

const (
	HoleTrack   HoleType = iota // main playing track (X, c, C, i)
	HoleBase                    // start positions for marbles ($, S, s) — 4 per player
	HoleHomeRow                 // ending positions (E, m) — 4 per player
	HoleCenter                  // center hole (0)
	HoleStart                   // track entry point where marbles enter play
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

// startPositions computes the 4 radial start positions for a player.
// They lie along the bisector angle between the player's arm and the
// previous arm (CCW), at distances 7, 6, 5, 4 grid cells from center.
// The outermost ($) is at the same radius as the intersection (i) at r=7.
func startPositions(armAngle float64, numPlayers int, d float64) [4][2]float64 {
	bisectorAngle := armAngle + 180.0/float64(numPlayers) // bisect toward CCW neighbor
	rad := bisectorAngle * math.Pi / 180
	dx, dy := math.Cos(rad), math.Sin(rad)

	var positions [4][2]float64
	for i := 0; i < 4; i++ {
		r := float64(7-i) * d // 7d, 6d, 5d, 4d from center
		positions[i] = [2]float64{r * dx, r * dy}
	}
	return positions
}

// ─────────────────────────────────────────────────────────────
// 4-Player Board — canonical grid layout
// ─────────────────────────────────────────────────────────────
//
// Station map (one of four, pointing UP):
//
//   $    XXiXX        $ S S s = start positions (radial, 4 per player)
//    S   X E X        X = track, i = intersection, c = connector
//     S  X E X        E = ending position, m = innermost ending
//      s X E X        C = center connector (shared between stations)
//        c m c        0 = center
//        C   C
//
//          0
//
//        C   C
//
// Track path: X positions form the perimeter loop.
// The track connects between stations via c → C → next station's C → c.
// No diagonal track holes — start positions fill the corner areas instead.
// 89 holes total: 4×17 station + 4×4 start + 4 center connectors + 1 center.

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

		// ── Center connectors (C) — inner corners of the cross ──
		{5, 5, HoleTrack, -1}, {9, 5, HoleTrack, -1},
		{5, 9, HoleTrack, -1}, {9, 9, HoleTrack, -1},

		// ── Top station (player 0) ──
		// Base row: X X i X X (all track)
		{5, 0, HoleTrack, -1}, {6, 0, HoleTrack, -1}, {7, 0, HoleTrack, -1},
		{8, 0, HoleTrack, -1}, {9, 0, HoleTrack, -1},
		// Left track column (rows 1-3) + station connector c (row 4)
		{5, 1, HoleTrack, -1}, {5, 2, HoleTrack, -1},
		{5, 3, HoleTrack, -1}, {5, 4, HoleTrack, -1},
		// Ending/home column: E E E m (rows 1-4)
		{7, 1, HoleHomeRow, 0}, {7, 2, HoleHomeRow, 0},
		{7, 3, HoleHomeRow, 0}, {7, 4, HoleHomeRow, 0},
		// Right track column: start (row 1), track (rows 2-3), c (row 4)
		{9, 1, HoleStart, 0},
		{9, 2, HoleTrack, -1}, {9, 3, HoleTrack, -1}, {9, 4, HoleTrack, -1},

		// ── Right station (player 1) ──
		{14, 5, HoleTrack, -1}, {14, 6, HoleTrack, -1}, {14, 7, HoleTrack, -1},
		{14, 8, HoleTrack, -1}, {14, 9, HoleTrack, -1},
		{13, 5, HoleTrack, -1}, {12, 5, HoleTrack, -1},
		{11, 5, HoleTrack, -1}, {10, 5, HoleTrack, -1},
		{13, 7, HoleHomeRow, 1}, {12, 7, HoleHomeRow, 1},
		{11, 7, HoleHomeRow, 1}, {10, 7, HoleHomeRow, 1},
		{13, 9, HoleStart, 1},
		{12, 9, HoleTrack, -1}, {11, 9, HoleTrack, -1}, {10, 9, HoleTrack, -1},

		// ── Bottom station (player 2) ──
		{9, 14, HoleTrack, -1}, {8, 14, HoleTrack, -1}, {7, 14, HoleTrack, -1},
		{6, 14, HoleTrack, -1}, {5, 14, HoleTrack, -1},
		{9, 13, HoleTrack, -1}, {9, 12, HoleTrack, -1},
		{9, 11, HoleTrack, -1}, {9, 10, HoleTrack, -1},
		{7, 13, HoleHomeRow, 2}, {7, 12, HoleHomeRow, 2},
		{7, 11, HoleHomeRow, 2}, {7, 10, HoleHomeRow, 2},
		{5, 13, HoleStart, 2},
		{5, 12, HoleTrack, -1}, {5, 11, HoleTrack, -1}, {5, 10, HoleTrack, -1},

		// ── Left station (player 3) ──
		{0, 9, HoleTrack, -1}, {0, 8, HoleTrack, -1}, {0, 7, HoleTrack, -1},
		{0, 6, HoleTrack, -1}, {0, 5, HoleTrack, -1},
		{1, 9, HoleTrack, -1}, {2, 9, HoleTrack, -1},
		{3, 9, HoleTrack, -1}, {4, 9, HoleTrack, -1},
		{1, 7, HoleHomeRow, 3}, {2, 7, HoleHomeRow, 3},
		{3, 7, HoleHomeRow, 3}, {4, 7, HoleHomeRow, 3},
		{1, 5, HoleStart, 3},
		{2, 5, HoleTrack, -1}, {3, 5, HoleTrack, -1}, {4, 5, HoleTrack, -1},
	}

	// Add grid-based station positions
	for _, p := range positions {
		x, y := gridToPhys(p.col, p.row, d)
		b.addHole(x, y, p.htype, p.player)
	}

	// Add radial start positions (4 per player)
	armAngles := []float64{90, 0, -90, 180}
	for player, armAngle := range armAngles {
		starts := startPositions(armAngle, 4, d)
		for _, pos := range starts {
			b.addHole(pos[0], pos[1], HoleBase, player)
		}
	}

	b.addBoardText(d)
}

// ─────────────────────────────────────────────────────────────
// N-Player Board (3, 5, or 6) — rotation-based layout
// ─────────────────────────────────────────────────────────────
// Same station structure rotated for each arm.
// Flanking track columns start at safe radius:
//   3 players (120°): r ≥ 2   →  5 track rows/side
//   5 players  (72°): r ≥ 4   →  3 track rows/side
//   6 players  (60°): r ≥ 5   →  2 track rows/side
// 4 radial start positions per player along bisector angle.
// One connector hole per gap at midpoint of adjacent arm track inners.

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

	armAngles := make([]float64, n)
	for i := 0; i < n; i++ {
		armAngles[i] = 90.0 - float64(i)*360.0/float64(n)
	}

	// Center hole
	b.addHole(0, 0, HoleCenter, -1)

	// ── Station template (arm pointing along +Y) ──
	type tpos struct {
		x, y  float64
		htype HoleType
		local bool // true = player-owned (home), false = shared (track)
		start bool // true = mark as HoleStart
	}

	// Base row: X X i X X at y=7 (all track)
	baseRow := []tpos{
		{-2, 7, HoleTrack, false, false}, {-1, 7, HoleTrack, false, false},
		{0, 7, HoleTrack, false, false}, // i (intersection)
		{1, 7, HoleTrack, false, false}, {2, 7, HoleTrack, false, false},
	}

	// Home/ending column: E at y=6,5,4 and m at y=3 (4 positions)
	// (home row goes from base inward toward center)
	home := []tpos{
		{0, 6, HoleHomeRow, true, false}, {0, 5, HoleHomeRow, true, false},
		{0, 4, HoleHomeRow, true, false}, {0, 3, HoleHomeRow, true, false},
	}

	// Flanking track columns: y = minR .. 6
	var leftTrack, rightTrack []tpos
	for r := 6; r >= minR; r-- {
		leftTrack = append(leftTrack, tpos{-2, float64(r), HoleTrack, false, false})
	}
	// Right track: first row (y=6) is the start/entry position
	rightTrack = append(rightTrack, tpos{2, 6, HoleTrack, false, true})
	for r := 5; r >= minR; r-- {
		rightTrack = append(rightTrack, tpos{2, float64(r), HoleTrack, false, false})
	}

	station := make([]tpos, 0, 20)
	station = append(station, baseRow...)
	station = append(station, home...)
	station = append(station, leftTrack...)
	station = append(station, rightTrack...)

	// Place stations
	for player := 0; player < n; player++ {
		rot := armAngles[player] - 90

		for _, tp := range station {
			rx, ry := rotatePoint(tp.x*d, tp.y*d, rot)
			ht := tp.htype
			pl := -1
			if tp.local {
				pl = player
			}
			if tp.start {
				ht = HoleStart
				pl = player
			}
			b.addHole(rx, ry, ht, pl)
		}

		// Radial start positions (4 per player)
		starts := startPositions(armAngles[player], n, d)
		for _, pos := range starts {
			b.addHole(pos[0], pos[1], HoleBase, player)
		}
	}

	// ── Gap connector holes ──
	for i := 0; i < n; i++ {
		rot0 := armAngles[i] - 90
		rot1 := armAngles[(i+1)%n] - 90

		rx, ry := rotatePoint(2*d, float64(minR)*d, rot0)
		lx, ly := rotatePoint(-2*d, float64(minR)*d, rot1)

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
			rad := angle * math.Pi / 180
			b.TextItems = append(b.TextItems, TextItem{
				X: 8 * d * math.Cos(rad), Y: 8 * d * math.Sin(rad),
				Text: fmt.Sprintf("P%d", i+1),
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
