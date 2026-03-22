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

	BallEndDiameter float64 // ball end mill diameter (inches)
	VBitAngle       float64 // V-bit included angle (degrees)

	SpindleSpeed int
	FeedRateXY   float64
	FeedRateZ    float64
	DepthPerPass float64

	TextDepth  float64
	TextHeight float64

	SafeZ      float64
	ClearanceZ float64
}

func DefaultParams() Params {
	return Params{
		BoardDiameter:  20.0,
		MarbleDiameter: 0.625,
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

func (p Params) GridSpacing() float64  { return 1.75 * p.MarbleDiameter }
func (p Params) EdgeMargin() float64   { return 2.5 * p.MarbleDiameter }
func (p Params) HoleDiameter() float64 { return p.MarbleDiameter }
func (p Params) HoleDepth() float64    { return p.MarbleDiameter / 4.0 }

// ConnectorCircleRadius returns the radius (in grid cells) of the circle
// formed by the C (center connector) positions. Sized so the chord between
// adjacent C positions equals 4 grid cells (the c-to-c distance in a station).
//   R = 2 / sin(π/N)
func ConnectorCircleRadius(n int) float64 {
	return 2.0 / math.Sin(math.Pi/float64(n))
}

// StationBaseY returns the y-coordinate (in grid cells from center) of the
// base row for an arm in an N-player layout. The station extends from the
// connector circle outward: C → c (1 cell) → 3 rows → base row (4 cells).
//   baseY = 2/tan(π/N) + 5
func StationBaseY(n int) float64 {
	return 2.0/math.Tan(math.Pi/float64(n)) + 5.0
}

// MinBoardDiameterForPlayers returns the minimum board diameter:
//   2 * (distance_0_to_i + 2 * marbleDiameter * 1.5)
// where distance_0_to_i = baseY * gridSpacing.
func MinBoardDiameterForPlayers(numPlayers int, marbleDiameter float64) float64 {
	d := 1.75 * marbleDiameter
	distToI := StationBaseY(numPlayers) * d
	return 2 * (distToI + 2*marbleDiameter*1.5)
}

func (p Params) MinBoardDiameter() float64 {
	return MinBoardDiameterForPlayers(p.NumPlayers, p.MarbleDiameter)
}

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
		return fmt.Errorf("board diameter %.1f\" too small for %d players with %.3f\" marbles (need at least %.1f\")",
			p.BoardDiameter, p.NumPlayers, p.MarbleDiameter, minDiam)
	}

	return nil
}
