package generate

import (
	"fmt"
	"math"
)

// Params holds all configuration for board generation.
type Params struct {
	BoardDiameter  float64 // diameter of round board (inches)
	MarbleDiameter float64 // marble diameter (inches)
	NumPlayers     int     // 3 to 6

	// Tool definitions
	BallEndDiameter float64 // ball end mill diameter (inches)
	VBitAngle       float64 // V-bit included angle (degrees)

	// Speeds and feeds
	SpindleSpeed int     // RPM
	FeedRateXY   float64 // horizontal feed (ipm)
	FeedRateZ    float64 // plunge feed (ipm)
	DepthPerPass float64 // max depth per pass (inches)

	// Text
	TextDepth  float64 // depth for V-bit text engraving (inches)
	TextHeight float64 // character height for text (inches)

	// Heights
	SafeZ      float64 // safe retract height (inches)
	ClearanceZ float64 // clearance plane above work (inches)
}

// DefaultParams returns sensible defaults for hardwood aggravation board.
func DefaultParams() Params {
	return Params{
		BoardDiameter:  20.0,
		MarbleDiameter: 0.625, // 5/8"
		NumPlayers:     4,

		BallEndDiameter: 0.25,
		VBitAngle:       60.0,

		SpindleSpeed: 18000,
		FeedRateXY:   80.0,
		FeedRateZ:    30.0,
		DepthPerPass: 0.125,

		TextDepth:  0.03,
		TextHeight: 0.5,

		SafeZ:      0.5,
		ClearanceZ: 0.1,
	}
}

// GridSpacing returns the distance between adjacent grid cells (1.75 * marble diameter).
func (p Params) GridSpacing() float64 {
	return 1.75 * p.MarbleDiameter
}

// EdgeMargin returns the distance from outermost marble center to board edge (2.5 * marble diameter).
func (p Params) EdgeMargin() float64 {
	return 2.5 * p.MarbleDiameter
}

// HoleDiameter returns the pocket opening diameter (matches marble).
func (p Params) HoleDiameter() float64 {
	return p.MarbleDiameter
}

// HoleDepth returns the pocket depth (marbleRadius / 2 = marbleDiameter / 4).
func (p Params) HoleDepth() float64 {
	return p.MarbleDiameter / 4.0
}

// MaxRadius returns the distance from board center to the farthest hole.
// The farthest holes are base row corner positions at (±2, 7) in arm-local
// grid coordinates = sqrt(4+49) ≈ 7.28 grid cells from center.
// This is the same for all player counts.
func (p Params) MaxRadius() float64 {
	return math.Sqrt(4+49) * p.GridSpacing()
}

// MinBoardDiameterForPlayers returns the minimum board diameter for a given
// player count and marble diameter.
func MinBoardDiameterForPlayers(numPlayers int, marbleDiameter float64) float64 {
	d := 1.75 * marbleDiameter
	em := 2.5 * marbleDiameter
	maxR := math.Sqrt(4+49) * d // base row corners at sqrt(53) grid cells
	return 2 * (maxR + em)
}

// MinBoardDiameter returns the minimum board diameter for this config.
func (p Params) MinBoardDiameter() float64 {
	return MinBoardDiameterForPlayers(p.NumPlayers, p.MarbleDiameter)
}

// Validate checks that parameters are physically reasonable.
func (p Params) Validate() error {
	if p.BoardDiameter < 10 {
		return fmt.Errorf("board diameter %.1f\" is too small (min 10\")", p.BoardDiameter)
	}
	if p.MarbleDiameter < 0.25 || p.MarbleDiameter > 2.0 {
		return fmt.Errorf("marble diameter %.3f\" out of range (0.25-2.0\")", p.MarbleDiameter)
	}
	if p.NumPlayers < 3 || p.NumPlayers > 6 {
		return fmt.Errorf("number of players must be 3 to 6, got %d", p.NumPlayers)
	}

	minDiam := p.MinBoardDiameter()
	if p.BoardDiameter < minDiam-0.1 {
		return fmt.Errorf("board diameter %.1f\" too small for %.3f\" marbles (need at least %.1f\")",
			p.BoardDiameter, p.MarbleDiameter, minDiam)
	}

	return nil
}
