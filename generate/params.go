package generate

import (
	"fmt"
	"math"
)

// Params holds all configuration for board generation.
type Params struct {
	BoardDiameter  float64 // diameter of round board (inches)
	BoardThickness float64 // board thickness (inches) — for double-sided depth check
	MarbleDiameter float64 // marble diameter (inches)
	NumPlayers     int     // 4 or 6

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
		BoardDiameter:  26.0,
		BoardThickness: 0.75,
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

// GridSpacing returns the distance between adjacent grid cells.
// Each character cell in the board map is this far apart.
func (p Params) GridSpacing() float64 {
	return 1.75 * p.MarbleDiameter
}

// EdgeMargin returns the distance from outermost marble center to board edge.
func (p Params) EdgeMargin() float64 {
	return 2.5 * p.MarbleDiameter
}

// HoleDiameter returns the pocket opening diameter (matches marble).
func (p Params) HoleDiameter() float64 {
	return p.MarbleDiameter
}

// HoleDepth returns the pocket depth = marbleRadius / 2.
func (p Params) HoleDepth() float64 {
	return p.MarbleDiameter / 4.0
}

// MaxRadius returns the distance from board center to the farthest hole.
// For 4-player: corner diagonals at grid (0,0) = 7√2 cells from center.
// For 6-player: similar extent.
func (p Params) MaxRadius() float64 {
	return 7.0 * math.Sqrt(2) * p.GridSpacing()
}

// MinBoardDiameter returns the minimum board diameter to fit all holes.
func (p Params) MinBoardDiameter() float64 {
	return 2 * (p.MaxRadius() + p.EdgeMargin())
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

	minDiam := p.MinBoardDiameter()
	if p.BoardDiameter < minDiam {
		return fmt.Errorf("board diameter %.1f\" too small for %.3f\" marbles (need at least %.1f\")",
			p.BoardDiameter, p.MarbleDiameter, minDiam)
	}

	// Double-sided depth check
	depth := p.HoleDepth()
	if 2*depth > p.BoardThickness-0.0625 {
		return fmt.Errorf("double-sided conflict: hole depth %.3f\" x 2 = %.3f\" exceeds board thickness %.3f\" minus 1/16\" floor",
			depth, 2*depth, p.BoardThickness)
	}

	return nil
}
