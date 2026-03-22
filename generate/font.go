package generate

import "math"

// Stroke font for CNC text engraving.
// Each glyph is defined as a series of strokes (polylines).
// Coordinates are normalized: width varies per character, height is 1.0.
// A nil entry in a stroke list means "pen up" (lift between segments).

type point struct {
	X, Y float64
}

type glyph struct {
	Width   float64   // character width (normalized)
	Strokes [][]point // each []point is a continuous polyline
}

var strokeFont map[rune]glyph

func init() {
	strokeFont = map[rune]glyph{
		'A': {Width: 0.7, Strokes: [][]point{
			{{0, 0}, {0.35, 1}, {0.7, 0}},
			{{0.15, 0.4}, {0.55, 0.4}},
		}},
		'B': {Width: 0.65, Strokes: [][]point{
			{{0, 0}, {0, 1}, {0.45, 1}, {0.6, 0.85}, {0.6, 0.65}, {0.45, 0.5}, {0, 0.5}},
			{{0.45, 0.5}, {0.65, 0.35}, {0.65, 0.15}, {0.45, 0}, {0, 0}},
		}},
		'C': {Width: 0.65, Strokes: [][]point{
			{{0.65, 0.85}, {0.45, 1}, {0.2, 1}, {0, 0.8}, {0, 0.2}, {0.2, 0}, {0.45, 0}, {0.65, 0.15}},
		}},
		'D': {Width: 0.65, Strokes: [][]point{
			{{0, 0}, {0, 1}, {0.4, 1}, {0.65, 0.8}, {0.65, 0.2}, {0.4, 0}, {0, 0}},
		}},
		'E': {Width: 0.6, Strokes: [][]point{
			{{0.6, 1}, {0, 1}, {0, 0}, {0.6, 0}},
			{{0, 0.5}, {0.45, 0.5}},
		}},
		'F': {Width: 0.55, Strokes: [][]point{
			{{0.55, 1}, {0, 1}, {0, 0}},
			{{0, 0.5}, {0.4, 0.5}},
		}},
		'G': {Width: 0.7, Strokes: [][]point{
			{{0.65, 0.85}, {0.45, 1}, {0.2, 1}, {0, 0.8}, {0, 0.2}, {0.2, 0}, {0.5, 0}, {0.7, 0.2}, {0.7, 0.5}, {0.4, 0.5}},
		}},
		'H': {Width: 0.65, Strokes: [][]point{
			{{0, 0}, {0, 1}},
			{{0.65, 0}, {0.65, 1}},
			{{0, 0.5}, {0.65, 0.5}},
		}},
		'I': {Width: 0.3, Strokes: [][]point{
			{{0, 1}, {0.3, 1}},
			{{0.15, 1}, {0.15, 0}},
			{{0, 0}, {0.3, 0}},
		}},
		'J': {Width: 0.5, Strokes: [][]point{
			{{0.1, 1}, {0.5, 1}},
			{{0.35, 1}, {0.35, 0.15}, {0.2, 0}, {0.05, 0}, {0, 0.15}},
		}},
		'K': {Width: 0.6, Strokes: [][]point{
			{{0, 0}, {0, 1}},
			{{0.6, 1}, {0, 0.4}},
			{{0.15, 0.55}, {0.6, 0}},
		}},
		'L': {Width: 0.55, Strokes: [][]point{
			{{0, 1}, {0, 0}, {0.55, 0}},
		}},
		'M': {Width: 0.8, Strokes: [][]point{
			{{0, 0}, {0, 1}, {0.4, 0.4}, {0.8, 1}, {0.8, 0}},
		}},
		'N': {Width: 0.65, Strokes: [][]point{
			{{0, 0}, {0, 1}, {0.65, 0}, {0.65, 1}},
		}},
		'O': {Width: 0.7, Strokes: [][]point{
			{{0.2, 0}, {0, 0.2}, {0, 0.8}, {0.2, 1}, {0.5, 1}, {0.7, 0.8}, {0.7, 0.2}, {0.5, 0}, {0.2, 0}},
		}},
		'P': {Width: 0.6, Strokes: [][]point{
			{{0, 0}, {0, 1}, {0.4, 1}, {0.6, 0.85}, {0.6, 0.65}, {0.4, 0.5}, {0, 0.5}},
		}},
		'Q': {Width: 0.7, Strokes: [][]point{
			{{0.2, 0}, {0, 0.2}, {0, 0.8}, {0.2, 1}, {0.5, 1}, {0.7, 0.8}, {0.7, 0.2}, {0.5, 0}, {0.2, 0}},
			{{0.45, 0.2}, {0.7, 0}},
		}},
		'R': {Width: 0.65, Strokes: [][]point{
			{{0, 0}, {0, 1}, {0.4, 1}, {0.6, 0.85}, {0.6, 0.65}, {0.4, 0.5}, {0, 0.5}},
			{{0.35, 0.5}, {0.65, 0}},
		}},
		'S': {Width: 0.6, Strokes: [][]point{
			{{0.6, 0.85}, {0.45, 1}, {0.15, 1}, {0, 0.85}, {0, 0.65}, {0.15, 0.5}, {0.45, 0.5}, {0.6, 0.35}, {0.6, 0.15}, {0.45, 0}, {0.15, 0}, {0, 0.15}},
		}},
		'T': {Width: 0.6, Strokes: [][]point{
			{{0, 1}, {0.6, 1}},
			{{0.3, 1}, {0.3, 0}},
		}},
		'U': {Width: 0.65, Strokes: [][]point{
			{{0, 1}, {0, 0.2}, {0.15, 0}, {0.5, 0}, {0.65, 0.2}, {0.65, 1}},
		}},
		'V': {Width: 0.7, Strokes: [][]point{
			{{0, 1}, {0.35, 0}, {0.7, 1}},
		}},
		'W': {Width: 0.9, Strokes: [][]point{
			{{0, 1}, {0.2, 0}, {0.45, 0.6}, {0.7, 0}, {0.9, 1}},
		}},
		'X': {Width: 0.65, Strokes: [][]point{
			{{0, 0}, {0.65, 1}},
			{{0, 1}, {0.65, 0}},
		}},
		'Y': {Width: 0.65, Strokes: [][]point{
			{{0, 1}, {0.325, 0.5}},
			{{0.65, 1}, {0.325, 0.5}, {0.325, 0}},
		}},
		'Z': {Width: 0.6, Strokes: [][]point{
			{{0, 1}, {0.6, 1}, {0, 0}, {0.6, 0}},
		}},
		'0': {Width: 0.65, Strokes: [][]point{
			{{0.2, 0}, {0, 0.2}, {0, 0.8}, {0.2, 1}, {0.45, 1}, {0.65, 0.8}, {0.65, 0.2}, {0.45, 0}, {0.2, 0}},
			{{0.1, 0.15}, {0.55, 0.85}},
		}},
		'1': {Width: 0.4, Strokes: [][]point{
			{{0.1, 0.8}, {0.25, 1}, {0.25, 0}},
			{{0.05, 0}, {0.4, 0}},
		}},
		'2': {Width: 0.6, Strokes: [][]point{
			{{0, 0.8}, {0.15, 1}, {0.45, 1}, {0.6, 0.85}, {0.6, 0.65}, {0, 0}, {0.6, 0}},
		}},
		'3': {Width: 0.6, Strokes: [][]point{
			{{0, 0.85}, {0.15, 1}, {0.45, 1}, {0.6, 0.85}, {0.6, 0.65}, {0.4, 0.5}},
			{{0.4, 0.5}, {0.6, 0.35}, {0.6, 0.15}, {0.45, 0}, {0.15, 0}, {0, 0.15}},
		}},
		'4': {Width: 0.65, Strokes: [][]point{
			{{0.5, 0}, {0.5, 1}, {0, 0.35}, {0.65, 0.35}},
		}},
		'5': {Width: 0.6, Strokes: [][]point{
			{{0.55, 1}, {0, 1}, {0, 0.55}, {0.35, 0.55}, {0.55, 0.4}, {0.55, 0.15}, {0.4, 0}, {0.1, 0}, {0, 0.1}},
		}},
		'6': {Width: 0.6, Strokes: [][]point{
			{{0.5, 1}, {0.2, 1}, {0, 0.75}, {0, 0.15}, {0.15, 0}, {0.45, 0}, {0.6, 0.15}, {0.6, 0.4}, {0.45, 0.55}, {0.15, 0.55}, {0, 0.4}},
		}},
		'7': {Width: 0.6, Strokes: [][]point{
			{{0, 1}, {0.6, 1}, {0.2, 0}},
		}},
		'8': {Width: 0.6, Strokes: [][]point{
			{{0.15, 0.5}, {0, 0.65}, {0, 0.85}, {0.15, 1}, {0.45, 1}, {0.6, 0.85}, {0.6, 0.65}, {0.45, 0.5}, {0.15, 0.5}},
			{{0.15, 0.5}, {0, 0.35}, {0, 0.15}, {0.15, 0}, {0.45, 0}, {0.6, 0.15}, {0.6, 0.35}, {0.45, 0.5}},
		}},
		'9': {Width: 0.6, Strokes: [][]point{
			{{0.6, 0.6}, {0.45, 0.45}, {0.15, 0.45}, {0, 0.6}, {0, 0.85}, {0.15, 1}, {0.45, 1}, {0.6, 0.85}, {0.6, 0.25}, {0.45, 0}, {0.15, 0}},
		}},
		' ': {Width: 0.4, Strokes: nil},
		'-': {Width: 0.4, Strokes: [][]point{
			{{0.05, 0.5}, {0.35, 0.5}},
		}},
		'.': {Width: 0.2, Strokes: [][]point{
			{{0.1, 0}, {0.1, 0.05}},
		}},
		'/': {Width: 0.4, Strokes: [][]point{
			{{0, 0}, {0.4, 1}},
		}},
		'"': {Width: 0.3, Strokes: [][]point{
			{{0.05, 1}, {0.05, 0.8}},
			{{0.25, 1}, {0.25, 0.8}},
		}},
	}
}

// TextToStrokes converts a string to physical stroke paths.
// Returns a slice of polylines (each polyline is []point in physical coords).
// originX, originY is the text anchor point. angleDeg rotates the text
// around the anchor. If centerOn, the text is centered on the anchor.
func TextToStrokes(text string, originX, originY, height, angleDeg float64, centerOn bool) [][]point {
	spacing := 0.15 // inter-character gap (normalized)

	// Calculate total width for centering
	totalWidth := 0.0
	for _, ch := range text {
		g, ok := strokeFont[ch]
		if !ok {
			g = strokeFont[' ']
		}
		totalWidth += g.Width + spacing
	}
	totalWidth -= spacing
	totalWidth *= height

	// Build strokes in local coordinates (text along +X from origin)
	localStartX := 0.0
	if centerOn {
		localStartX = -totalWidth / 2
	}

	var result [][]point
	curX := localStartX

	angleRad := angleDeg * math.Pi / 180.0
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)

	for _, ch := range text {
		g, ok := strokeFont[ch]
		if !ok {
			g = strokeFont[' ']
		}

		for _, stroke := range g.Strokes {
			physical := make([]point, len(stroke))
			for i, p := range stroke {
				// Local position relative to anchor
				lx := curX + p.X*height
				ly := p.Y * height
				// Rotate around anchor, then translate
				physical[i] = point{
					X: originX + lx*cosA - ly*sinA,
					Y: originY + lx*sinA + ly*cosA,
				}
			}
			result = append(result, physical)
		}

		curX += (g.Width + spacing) * height
	}

	return result
}

