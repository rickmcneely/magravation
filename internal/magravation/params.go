package magravation

import "fmt"

// Params holds all configuration for board generation.
type Params struct {
	BoardSize       float64 // side length of square board (inches)
	MarbleDiameter  float64 // marble diameter (inches)
	DiceSize        float64 // dice side length (inches)
	NumPlayers      int     // 4 or 6
	Clearance       float64 // extra clearance for marble holes (inches)
	WallMin         float64 // minimum wall between holes (inches)
	SafeZ           float64 // safe retract height (inches)
	ClearanceZ      float64 // clearance plane above work (inches)
	MarbleDepthFrac float64 // fraction of marble diameter for hole depth
	DiceDepthFrac   float64 // fraction of dice size for pocket depth
	TextDepth       float64 // depth for V-bit text engraving (inches)
	TextHeight      float64 // character height for text (inches)
	BoardMargin     float64 // margin from board edge (inches)

	// Tool definitions
	BallEndDiameter  float64 // 1/4" ball end mill diameter
	StraightDiameter float64 // 1/8" straight bit diameter
	VBitAngle        float64 // V-bit included angle (degrees)

	// Speeds and feeds
	SpindleSpeed    int     // RPM
	FeedRateXY      float64 // horizontal feed (ipm)
	FeedRateZ       float64 // plunge feed (ipm)
	DepthPerPass    float64 // max depth per pass (inches)
	PocketStepover  float64 // stepover fraction for pocketing (0-1)

	// Origin at center of board (true) or bottom-left corner (false)
	CenterOrigin bool
}

// DefaultParams returns sensible defaults for hardwood aggravation board.
func DefaultParams() Params {
	return Params{
		BoardSize:       24.0,
		MarbleDiameter:  0.625, // 5/8"
		DiceSize:        0.625, // 5/8" = ~16mm standard dice
		NumPlayers:      4,
		Clearance:       0.03125, // 1/32"
		WallMin:         0.1875,  // 3/16"
		SafeZ:           0.5,
		ClearanceZ:      0.1,
		MarbleDepthFrac: 0.55,
		DiceDepthFrac:   0.65,
		TextDepth:       0.03,
		TextHeight:      0.5,
		BoardMargin:     1.0,

		BallEndDiameter:  0.25,
		StraightDiameter: 0.125,
		VBitAngle:        60.0,

		SpindleSpeed:    18000,
		FeedRateXY:      80.0,
		FeedRateZ:       30.0,
		DepthPerPass:    0.125,
		PocketStepover:  0.40,

		CenterOrigin: true,
	}
}

// Validate checks that parameters are physically reasonable.
func (p Params) Validate() error {
	if p.BoardSize < 10 {
		return fmt.Errorf("board size %.1f\" is too small (min 10\")", p.BoardSize)
	}
	if p.MarbleDiameter < 0.25 || p.MarbleDiameter > 2.0 {
		return fmt.Errorf("marble diameter %.3f\" out of range (0.25-2.0\")", p.MarbleDiameter)
	}
	if p.DiceSize < 0.25 || p.DiceSize > 2.0 {
		return fmt.Errorf("dice size %.3f\" out of range (0.25-2.0\")", p.DiceSize)
	}
	if p.NumPlayers != 4 && p.NumPlayers != 6 {
		return fmt.Errorf("number of players must be 4 or 6, got %d", p.NumPlayers)
	}
	if p.BallEndDiameter > p.MarbleDiameter+p.Clearance {
		return fmt.Errorf("ball end mill (%.3f\") is larger than marble hole (%.3f\")", p.BallEndDiameter, p.MarbleDiameter+p.Clearance)
	}

	// Check that holes fit on the board
	holeSpacing := p.MarbleDiameter + p.Clearance + p.WallMin
	gridSpan := 14.0 * holeSpacing // 15 grid positions across
	usable := p.BoardSize - 2*p.BoardMargin
	if gridSpan > usable {
		return fmt.Errorf("board too small: need %.1f\" for holes but only %.1f\" usable", gridSpan, usable)
	}

	return nil
}

// HoleDiameter returns the cutting diameter for marble holes.
func (p Params) HoleDiameter() float64 {
	return p.MarbleDiameter + p.Clearance
}

// HoleDepth returns the cutting depth for marble holes.
func (p Params) HoleDepth() float64 {
	return p.MarbleDiameter * p.MarbleDepthFrac
}

// DicePocketDepth returns the depth for dice storage pockets.
func (p Params) DicePocketDepth() float64 {
	return p.DiceSize * p.DiceDepthFrac
}

// GridSpacing returns the distance between adjacent hole centers.
func (p Params) GridSpacing() float64 {
	return p.MarbleDiameter + p.Clearance + p.WallMin
}
