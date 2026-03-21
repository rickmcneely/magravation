package generate

import "fmt"

// Params holds all configuration for board generation.
type Params struct {
	BoardDiameter   float64 // diameter of round board (inches)
	BoardThickness  float64 // board thickness (inches) — for double-sided depth check
	MarbleDiameter  float64 // marble diameter (inches)
	NumPlayers      int     // 4 or 6
	Clearance       float64 // extra clearance for marble holes (inches)
	WallMin         float64 // minimum wall between holes (inches)
	SafeZ           float64 // safe retract height (inches)
	ClearanceZ      float64 // clearance plane above work (inches)
	MarbleDepthFrac float64 // fraction of marble diameter for hole depth
	TextDepth       float64 // depth for V-bit text engraving (inches)
	TextHeight      float64 // character height for text (inches)
	BoardMargin     float64 // margin from board edge (inches)

	// Tool definitions
	BallEndDiameter float64 // 1/4" ball end mill diameter
	VBitAngle       float64 // V-bit included angle (degrees)

	// Speeds and feeds
	SpindleSpeed int     // RPM
	FeedRateXY   float64 // horizontal feed (ipm)
	FeedRateZ    float64 // plunge feed (ipm)
	DepthPerPass float64 // max depth per pass (inches)

	// Origin is always center of round board
}

// DefaultParams returns sensible defaults for hardwood aggravation board.
func DefaultParams() Params {
	return Params{
		BoardDiameter:   24.0,
		BoardThickness:  0.75, // 3/4"
		MarbleDiameter:  0.625, // 5/8"
		NumPlayers:      4,
		Clearance:       0.03125, // 1/32"
		WallMin:         0.1875,  // 3/16"
		SafeZ:           0.5,
		ClearanceZ:      0.1,
		MarbleDepthFrac: 0.50,
		TextDepth:       0.03,
		TextHeight:      0.5,
		BoardMargin:     0.75,

		BallEndDiameter: 0.25,
		VBitAngle:       60.0,

		SpindleSpeed: 18000,
		FeedRateXY:   80.0,
		FeedRateZ:    30.0,
		DepthPerPass: 0.125,
	}
}

// Validate checks that parameters are physically reasonable.
func (p Params) Validate() error {
	if p.BoardDiameter < 10 {
		return fmt.Errorf("board diameter %.1f\" is too small (min 10\")", p.BoardDiameter)
	}
	if p.MarbleDiameter < 0.25 || p.MarbleDiameter > 2.0 {
		return fmt.Errorf("marble diameter %.3f\" out of range (0.25-2.0\")", p.MarbleDiameter)
	}
	if p.NumPlayers != 4 && p.NumPlayers != 6 {
		return fmt.Errorf("number of players must be 4 or 6, got %d", p.NumPlayers)
	}
	if p.BallEndDiameter > p.MarbleDiameter+p.Clearance {
		return fmt.Errorf("ball end mill (%.3f\") is larger than marble hole (%.3f\")", p.BallEndDiameter, p.MarbleDiameter+p.Clearance)
	}

	// Check that holes fit within the round board
	spacing := p.GridSpacing()
	maxRadius := 7.1 * spacing // arm tips at ~7.07 grid units from center
	usableRadius := p.BoardDiameter/2 - p.BoardMargin
	if maxRadius > usableRadius {
		return fmt.Errorf("board too small: need %.1f\" radius for holes but only %.1f\" usable radius",
			maxRadius, usableRadius)
	}

	// Check double-sided depth (center hole + some home rows overlap between sides)
	holeDepth := p.HoleDepth()
	if 2*holeDepth > p.BoardThickness-0.0625 {
		return fmt.Errorf("double-sided conflict: hole depth %.3f\" x 2 = %.3f\" exceeds board thickness %.3f\" minus 1/16\" floor; increase thickness or reduce marble depth fraction",
			holeDepth, 2*holeDepth, p.BoardThickness)
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

// GridSpacing returns the distance between adjacent hole centers.
func (p Params) GridSpacing() float64 {
	return p.MarbleDiameter + p.Clearance + p.WallMin
}
