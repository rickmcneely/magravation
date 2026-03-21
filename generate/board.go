package generate

import (
	"fmt"
	"math"
)

// HoleType identifies the purpose of a hole on the board.
type HoleType int

const (
	HoleTrack   HoleType = iota // main playing track
	HoleBase                    // starting base (home area)
	HoleHomeRow                 // safe path to center
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
	X, Y     float64 // position
	Text     string
	Height   float64 // character height
	Angle    float64 // rotation in degrees
	CenterOn bool    // if true, X,Y is center of text
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

// rotatePoint rotates (x, y) by angleDeg degrees CCW around the origin.
func rotatePoint(x, y, angleDeg float64) (float64, float64) {
	rad := angleDeg * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	return x*cos - y*sin, x*sin + y*cos
}

// addHoleXY adds a hole at physical coordinates.
func (b *Board) addHoleXY(x, y float64, htype HoleType, player int) {
	b.Holes = append(b.Holes, Hole{
		X:        x,
		Y:        y,
		Type:     htype,
		Player:   player,
		Diameter: b.Params.HoleDiameter(),
		Depth:    b.Params.HoleDepth(),
	})
}

// addHoleGrid adds a hole at grid coordinates scaled by spacing.
func (b *Board) addHoleGrid(gx, gy float64, htype HoleType, player int) {
	s := b.Params.GridSpacing()
	b.addHoleXY(gx*s, gy*s, htype, player)
}

// ─────────────────────────────────────────────
// 4-Player Board  (traditional cross/plus)
// ─────────────────────────────────────────────
// Arms at 90° intervals (N, E, S, W).
// Each arm is 3 holes wide, extending 7 units from center.
// Track runs along outer two columns; home row along center.
// Bases (5 holes each, in a line) in the concave corners between arms.

func (b *Board) generate4Player() {
	s := b.Params.GridSpacing()

	// Center hole
	b.addHoleGrid(0, 0, HoleCenter, -1)

	// Arm angles: North=90°, East=0°, South=-90°(270°), West=180°
	armAngles := []float64{90, 0, -90, 180}

	// Arm template (pointing along +Y from origin):
	// Home row
	homeRow := [][2]float64{{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}}
	// Left track (x = -1)
	leftTrack := [][2]float64{{-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {-1, 5}, {-1, 6}}
	// Right track (x = +1)
	rightTrack := [][2]float64{{1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}}
	// Tip
	tip := [][2]float64{{-1, 7}, {0, 7}, {1, 7}}
	// Start position: right column at row 6 (just before tip)
	startPos := [2]float64{1, 6}

	// Base template: 5 holes in a column in the CW corner (between this arm
	// and the next CW arm). Column runs parallel to this arm at x=3.
	basePositions := [][2]float64{{3, 2}, {3, 3}, {3, 4}, {3, 5}, {3, 6}}

	for player, angle := range armAngles {
		rot := angle - 90 // rotation from +Y template to arm direction

		// Home row
		for _, p := range homeRow {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleHomeRow, player)
		}

		// Track (left + right + tip)
		for _, p := range leftTrack {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleTrack, -1)
		}
		for _, p := range rightTrack {
			x, y := rotatePoint(p[0], p[1], rot)
			// Check if this is the start position
			if p == startPos {
				b.addHoleGrid(x, y, HoleStart, player)
			} else {
				b.addHoleGrid(x, y, HoleTrack, -1)
			}
		}
		for _, p := range tip {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleTrack, -1)
		}

		// Base holes
		for _, p := range basePositions {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleBase, player)
		}
	}

	// Text
	titleR := 7.5 * s
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: titleR,
		Text: "AGGRAVATION", Height: b.Params.TextHeight,
		CenterOn: true,
	})
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: -titleR,
		Text: "4 PLAYER", Height: b.Params.TextHeight * 0.5,
		CenterOn: true,
	})

	// Player labels near bases
	playerNames := []string{"P1", "P2", "P3", "P4"}
	for player, angle := range armAngles {
		rot := angle - 90
		// Label near base center (3.5, 3.5) rotated
		lx, ly := rotatePoint(3.5, 3.5, rot)
		b.TextItems = append(b.TextItems, TextItem{
			X: lx * s, Y: ly * s,
			Text:     playerNames[player],
			Height:   b.Params.TextHeight * 0.4,
			CenterOn: true,
		})
	}

	// "HOME" label near center
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: -0.8 * s,
		Text:     "HOME",
		Height:   b.Params.TextHeight * 0.35,
		CenterOn: true,
	})
}

// ─────────────────────────────────────────────
// 6-Player Board  (star with 6 arms at 60°)
// ─────────────────────────────────────────────
// Arms at 60° intervals. Inner portion (r=1..2) is home-row-only
// to avoid overlap. Flanking track columns start at r=3.
// Connecting holes bridge adjacent arms at the inner track radius.

func (b *Board) generate6Player() {
	s := b.Params.GridSpacing()

	// Center hole
	b.addHoleGrid(0, 0, HoleCenter, -1)

	// Arm angles (CW from top): 90°, 30°, -30°, -90°, -150°, 150°
	armAngles := []float64{90, 30, -30, -90, -150, 150}

	// ── Per-arm template (pointing along +Y) ──

	// Home row: 5 holes along center column
	homeRow := [][2]float64{{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}}

	// Flanking track columns start at r=3 to avoid 60° overlap
	leftTrack := [][2]float64{{-1, 3}, {-1, 4}, {-1, 5}, {-1, 6}}
	rightTrack := [][2]float64{{1, 3}, {1, 4}, {1, 5}, {1, 6}}

	// Tip: 3 holes at r=7
	tip := [][2]float64{{-1, 7}, {0, 7}, {1, 7}}

	// Start position: right column at row 6
	startPos := [2]float64{1, 6}

	// Base: 5 holes in a column between this arm and the next CW arm.
	// Column at x=3.5 to clear adjacent arm track holes at 60° spacing.
	basePositions := [][2]float64{{3.5, 2}, {3.5, 3}, {3.5, 4}, {3.5, 5}, {3.5, 6}}

	for player, angle := range armAngles {
		rot := angle - 90

		// Home row
		for _, p := range homeRow {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleHomeRow, player)
		}

		// Track left
		for _, p := range leftTrack {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleTrack, -1)
		}

		// Track right
		for _, p := range rightTrack {
			x, y := rotatePoint(p[0], p[1], rot)
			if p == startPos {
				b.addHoleGrid(x, y, HoleStart, player)
			} else {
				b.addHoleGrid(x, y, HoleTrack, -1)
			}
		}

		// Tip
		for _, p := range tip {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleTrack, -1)
		}

		// Base holes
		for _, p := range basePositions {
			x, y := rotatePoint(p[0], p[1], rot)
			b.addHoleGrid(x, y, HoleBase, player)
		}
	}

	// ── Connecting holes between adjacent arms ──
	// Bridge the gap between one arm's right-track inner (1,3)
	// and the next CW arm's left-track inner (-1,3).

	for i := 0; i < 6; i++ {
		rot0 := armAngles[i] - 90
		rot1 := armAngles[(i+1)%6] - 90

		// Arm i right-track inner
		rx, ry := rotatePoint(1, 3, rot0)
		// Arm i+1 left-track inner
		lx, ly := rotatePoint(-1, 3, rot1)

		// Midpoint connecting hole
		mx := (rx + lx) / 2
		my := (ry + ly) / 2
		b.addHoleGrid(mx, my, HoleTrack, -1)
	}

	// Text
	titleR := 7.5 * s
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: titleR,
		Text: "AGGRAVATION", Height: b.Params.TextHeight,
		CenterOn: true,
	})
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: -titleR,
		Text: "6 PLAYER", Height: b.Params.TextHeight * 0.5,
		CenterOn: true,
	})

	// Player labels near bases
	for player, angle := range armAngles {
		rot := angle - 90
		lx, ly := rotatePoint(2.5, 2, rot)
		b.TextItems = append(b.TextItems, TextItem{
			X: lx * s, Y: ly * s,
			Text:     fmt.Sprintf("P%d", player+1),
			Height:   b.Params.TextHeight * 0.35,
			CenterOn: true,
		})
	}

	// "HOME" label
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: -0.8 * s,
		Text:     "HOME",
		Height:   b.Params.TextHeight * 0.35,
		CenterOn: true,
	})
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
