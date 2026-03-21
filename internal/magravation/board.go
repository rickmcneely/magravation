package magravation

import "math"

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
	X, Y     float64  // physical position (inches)
	GridX    int      // grid position
	GridY    int      // grid position
	Type     HoleType // hole purpose
	Player   int      // player index (0-based), -1 for shared
	Diameter float64  // cutting diameter
	Depth    float64  // cutting depth
}

// DicePocket represents a rectangular pocket for dice storage.
type DicePocket struct {
	CenterX, CenterY float64 // center position
	Width, Height     float64 // pocket dimensions
	Depth             float64 // pocket depth
	CornerRadius      float64 // corner radius (min = tool radius)
	Label             string  // e.g., "DICE"
}

// TextItem represents text to be engraved on the board.
type TextItem struct {
	X, Y     float64 // starting position (bottom-left of text)
	Text     string
	Height   float64 // character height
	Angle    float64 // rotation in degrees (0 = horizontal)
	CenterOn bool    // if true, X,Y is center of text
}

// Board holds the complete board layout.
type Board struct {
	Params      Params
	Holes       []Hole
	DicePockets []DicePocket
	TextItems   []TextItem
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

// gridToPhysical converts grid coordinates to physical coordinates.
// Grid center (0,0) maps to board center.
func (b *Board) gridToPhysical(gx, gy int) (float64, float64) {
	spacing := b.Params.GridSpacing()
	x := float64(gx) * spacing
	y := float64(gy) * spacing
	if !b.Params.CenterOrigin {
		x += b.Params.BoardSize / 2
		y += b.Params.BoardSize / 2
	}
	return x, y
}

func (b *Board) addHole(gx, gy int, htype HoleType, player int) {
	x, y := b.gridToPhysical(gx, gy)
	b.Holes = append(b.Holes, Hole{
		X:        x,
		Y:        y,
		GridX:    gx,
		GridY:    gy,
		Type:     htype,
		Player:   player,
		Diameter: b.Params.HoleDiameter(),
		Depth:    b.Params.HoleDepth(),
	})
}

func (b *Board) generate4Player() {
	p := b.Params

	// Center hole
	b.addHole(0, 0, HoleCenter, -1)

	// Home rows (5 holes each, leading from track toward center)
	// Player 0 (North): home row along y=1..5 at x=0
	for i := 1; i <= 5; i++ {
		b.addHole(0, i, HoleHomeRow, 0)
	}
	// Player 1 (East): home row along x=1..5 at y=0
	for i := 1; i <= 5; i++ {
		b.addHole(i, 0, HoleHomeRow, 1)
	}
	// Player 2 (South): home row along y=-1..-5 at x=0
	for i := 1; i <= 5; i++ {
		b.addHole(0, -i, HoleHomeRow, 2)
	}
	// Player 3 (West): home row along x=-1..-5 at y=0
	for i := 1; i <= 5; i++ {
		b.addHole(-i, 0, HoleHomeRow, 3)
	}

	// Main track - 56 positions forming a loop around the cross
	// The cross arms extend from -7 to +7, are 3 units wide (columns -1, 0, 1)
	// Track runs along the outside edges of the cross

	type gridPos struct{ x, y int }
	track := []gridPos{}

	// Going clockwise from top-left of north arm:
	// Top of north arm (left to right)
	track = append(track, gridPos{-1, 7}, gridPos{0, 7}, gridPos{1, 7})
	// East side of north arm (going south)
	for y := 6; y >= 2; y-- {
		track = append(track, gridPos{1, y})
	}
	// Inner corner NE, across to east arm
	track = append(track, gridPos{1, 1}, gridPos{2, 1}, gridPos{3, 1}, gridPos{4, 1}, gridPos{5, 1}, gridPos{6, 1})
	// Tip of east arm
	track = append(track, gridPos{7, 1}, gridPos{7, 0}, gridPos{7, -1})
	// South side of east arm (going west)
	track = append(track, gridPos{6, -1}, gridPos{5, -1}, gridPos{4, -1}, gridPos{3, -1}, gridPos{2, -1})
	// Inner corner SE, down south arm
	track = append(track, gridPos{1, -1})
	for y := -2; y >= -6; y-- {
		track = append(track, gridPos{1, y})
	}
	// Bottom of south arm (right to left)
	track = append(track, gridPos{1, -7}, gridPos{0, -7}, gridPos{-1, -7})
	// West side of south arm (going north)
	for y := -6; y <= -2; y++ {
		track = append(track, gridPos{-1, y})
	}
	// Inner corner SW, across to west arm
	track = append(track, gridPos{-1, -1}, gridPos{-2, -1}, gridPos{-3, -1}, gridPos{-4, -1}, gridPos{-5, -1}, gridPos{-6, -1})
	// Tip of west arm
	track = append(track, gridPos{-7, -1}, gridPos{-7, 0}, gridPos{-7, 1})
	// North side of west arm (going east)
	track = append(track, gridPos{-6, 1}, gridPos{-5, 1}, gridPos{-4, 1}, gridPos{-3, 1}, gridPos{-2, 1})
	// Inner corner NW, up north arm
	track = append(track, gridPos{-1, 1})
	for y := 2; y <= 6; y++ {
		track = append(track, gridPos{-1, y})
	}

	// Mark start positions (where players enter the track)
	// Player 0 (North) enters at (1, 6) - just below the top on east side
	// Player 1 (East) enters at (6, -1) - just left of tip on south side
	// Player 2 (South) enters at (-1, -6) - just above bottom on west side
	// Player 3 (West) enters at (-6, 1) - just right of tip on north side
	startPositions := map[gridPos]int{
		{1, 6}:   0,
		{6, -1}:  1,
		{-1, -6}: 2,
		{-6, 1}:  3,
	}

	for _, pos := range track {
		if player, isStart := startPositions[pos]; isStart {
			b.addHole(pos.x, pos.y, HoleStart, player)
		} else {
			b.addHole(pos.x, pos.y, HoleTrack, -1)
		}
	}

	// Home bases (4 holes each, in the concave corners of the cross)
	// Player 0 (North): top-right corner
	bases0 := []gridPos{{3, 5}, {4, 5}, {3, 4}, {4, 4}}
	for _, pos := range bases0 {
		b.addHole(pos.x, pos.y, HoleBase, 0)
	}
	// Player 1 (East): bottom-right corner
	bases1 := []gridPos{{5, -3}, {5, -4}, {4, -3}, {4, -4}}
	for _, pos := range bases1 {
		b.addHole(pos.x, pos.y, HoleBase, 1)
	}
	// Player 2 (South): bottom-left corner
	bases2 := []gridPos{{-3, -5}, {-4, -5}, {-3, -4}, {-4, -4}}
	for _, pos := range bases2 {
		b.addHole(pos.x, pos.y, HoleBase, 2)
	}
	// Player 3 (West): top-left corner
	bases3 := []gridPos{{-5, 3}, {-5, 4}, {-4, 3}, {-4, 4}}
	for _, pos := range bases3 {
		b.addHole(pos.x, pos.y, HoleBase, 3)
	}

	// Dice storage pockets (in remaining corner areas)
	// Place dice pockets in the diagonal corners where there's open space
	spacing := p.GridSpacing()
	diceWidth := p.DiceSize + 0.0625  // 1/16" clearance per side
	diceDepth := p.DicePocketDepth()
	cornerRadius := p.StraightDiameter / 2
	if cornerRadius < 0.0625 {
		cornerRadius = 0.0625
	}

	// Top-left area (between player 3 base and player 0 track)
	dcx1, dcy1 := b.gridToPhysical(-4, 6)
	b.DicePockets = append(b.DicePockets, DicePocket{
		CenterX: dcx1, CenterY: dcy1,
		Width: diceWidth, Height: diceWidth,
		Depth: diceDepth, CornerRadius: cornerRadius,
		Label: "DICE",
	})
	// Bottom-right area
	dcx2, dcy2 := b.gridToPhysical(4, -6)
	b.DicePockets = append(b.DicePockets, DicePocket{
		CenterX: dcx2, CenterY: dcy2,
		Width: diceWidth, Height: diceWidth,
		Depth: diceDepth, CornerRadius: cornerRadius,
		Label: "DICE",
	})

	// Text items
	titleY := float64(7)*spacing + p.TextHeight*1.5
	if p.CenterOrigin {
		b.TextItems = append(b.TextItems, TextItem{
			X: 0, Y: titleY,
			Text: "AGGRAVATION", Height: p.TextHeight,
			CenterOn: true,
		})
	} else {
		b.TextItems = append(b.TextItems, TextItem{
			X: p.BoardSize / 2, Y: p.BoardSize/2 + titleY,
			Text: "AGGRAVATION", Height: p.TextHeight,
			CenterOn: true,
		})
	}

	// Player labels near their bases
	playerNames := []string{"NORTH", "EAST", "SOUTH", "WEST"}
	type labelPos struct {
		gx, gy int
		angle  float64
	}
	labels := []labelPos{
		{3, 3, 0},    // Player 0 - below their base
		{3, -5, 0},   // Player 1 - below their base
		{-4, -3, 0},  // Player 2 - above their base
		{-3, 5, 0},   // Player 3 - below their base
	}
	for i, lbl := range labels {
		lx, ly := b.gridToPhysical(lbl.gx, lbl.gy)
		b.TextItems = append(b.TextItems, TextItem{
			X: lx, Y: ly,
			Text:     playerNames[i],
			Height:   p.TextHeight * 0.6,
			Angle:    lbl.angle,
			CenterOn: true,
		})
	}

	// "HOME" label at center
	cx, cy := b.gridToPhysical(0, 0)
	b.TextItems = append(b.TextItems, TextItem{
		X: cx, Y: cy - spacing*0.7,
		Text:     "HOME",
		Height:   p.TextHeight * 0.5,
		CenterOn: true,
	})
}

func (b *Board) generate6Player() {
	p := b.Params

	// 6-player board uses a hexagonal/star layout
	// 6 arms at 60-degree intervals
	// Main track connects the arm tips

	b.addHole(0, 0, HoleCenter, -1)

	// Arm directions (starting from top, going clockwise)
	// 0°, 60°, 120°, 180°, 240°, 300°
	angles := []float64{90, 30, -30, -90, -150, 150} // degrees from positive X

	spacing := p.GridSpacing()

	// For each player/arm
	for player := 0; player < 6; player++ {
		angle := angles[player] * math.Pi / 180

		dx := math.Cos(angle)
		dy := math.Sin(angle)

		// Home row: 5 positions from center outward along arm
		for i := 1; i <= 5; i++ {
			x := float64(i) * spacing * dx
			y := float64(i) * spacing * dy
			if !p.CenterOrigin {
				x += p.BoardSize / 2
				y += p.BoardSize / 2
			}
			b.Holes = append(b.Holes, Hole{
				X: x, Y: y,
				Type: HoleHomeRow, Player: player,
				Diameter: p.HoleDiameter(), Depth: p.HoleDepth(),
			})
		}

		// Base holes: 4 holes offset to the right of the arm
		perpAngle := angle - math.Pi/2
		perpDx := math.Cos(perpAngle)
		perpDy := math.Sin(perpAngle)
		baseOffsets := [][2]float64{{4, 1.5}, {5, 1.5}, {4, 2.5}, {5, 2.5}}
		for _, off := range baseOffsets {
			x := off[0]*spacing*dx + off[1]*spacing*perpDx
			y := off[0]*spacing*dy + off[1]*spacing*perpDy
			if !p.CenterOrigin {
				x += p.BoardSize / 2
				y += p.BoardSize / 2
			}
			b.Holes = append(b.Holes, Hole{
				X: x, Y: y,
				Type: HoleBase, Player: player,
				Diameter: p.HoleDiameter(), Depth: p.HoleDepth(),
			})
		}
	}

	// Main track: hexagonal ring connecting arm tips
	// Between each pair of adjacent arms, place track holes along the perimeter
	trackRadius := 7.0 * spacing // distance from center to track
	totalTrackHoles := 60        // 10 per segment between arms

	for i := 0; i < totalTrackHoles; i++ {
		angle := (float64(i)/float64(totalTrackHoles))*2*math.Pi + math.Pi/2
		x := trackRadius * math.Cos(angle)
		y := trackRadius * math.Sin(angle)
		if !p.CenterOrigin {
			x += p.BoardSize / 2
			y += p.BoardSize / 2
		}

		// Check if this is a start position (near an arm direction)
		isStart := false
		startPlayer := -1
		for pl := 0; pl < 6; pl++ {
			armAngle := angles[pl] * math.Pi / 180
			diff := math.Abs(angle - armAngle)
			if diff > math.Pi {
				diff = 2*math.Pi - diff
			}
			if diff < 0.15 {
				isStart = true
				startPlayer = pl
				break
			}
		}

		htype := HoleTrack
		if isStart {
			htype = HoleStart
		}
		b.Holes = append(b.Holes, Hole{
			X: x, Y: y,
			Type: htype, Player: startPlayer,
			Diameter: p.HoleDiameter(), Depth: p.HoleDepth(),
		})
	}

	// Dice pockets
	diceWidth := p.DiceSize + 0.0625
	diceDepth := p.DicePocketDepth()
	cornerRadius := p.StraightDiameter / 2

	// Place two dice pockets on opposite sides
	dp1x := 5.0 * spacing * math.Cos(0)
	dp1y := 5.0 * spacing * math.Sin(0)
	dp2x := 5.0 * spacing * math.Cos(math.Pi)
	dp2y := 5.0 * spacing * math.Sin(math.Pi)
	if !p.CenterOrigin {
		dp1x += p.BoardSize / 2
		dp1y += p.BoardSize / 2
		dp2x += p.BoardSize / 2
		dp2y += p.BoardSize / 2
	}
	b.DicePockets = append(b.DicePockets,
		DicePocket{CenterX: dp1x, CenterY: dp1y, Width: diceWidth, Height: diceWidth, Depth: diceDepth, CornerRadius: cornerRadius, Label: "DICE"},
		DicePocket{CenterX: dp2x, CenterY: dp2y, Width: diceWidth, Height: diceWidth, Depth: diceDepth, CornerRadius: cornerRadius, Label: "DICE"},
	)

	// Text
	titleDist := 8.0 * spacing
	b.TextItems = append(b.TextItems, TextItem{
		X: 0, Y: titleDist,
		Text: "AGGRAVATION", Height: p.TextHeight,
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
