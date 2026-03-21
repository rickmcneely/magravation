package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"magravation/internal/magravation"
)

func main() {
	// Mode flags
	webMode := flag.Bool("web", false, "Run as web server")
	webPort := flag.String("port", "8080", "Web server port")

	// Board parameters
	boardSize := flag.Float64("board-size", 24.0, "Board side length in inches")
	marbleDiam := flag.Float64("marble-diameter", 0.625, "Marble diameter in inches (default 5/8\")")
	diceSize := flag.Float64("dice-size", 0.625, "Dice side length in inches")
	numPlayers := flag.Int("players", 4, "Number of players (4 or 6)")
	centerOrigin := flag.Bool("center-origin", true, "Use center of board as origin (false = bottom-left)")

	// Output options
	output := flag.String("output", "aggravation", "Output file prefix")
	splitFiles := flag.Bool("split", false, "Generate separate files per tool")
	previewOnly := flag.Bool("preview", false, "Generate SVG preview only")

	flag.Parse()

	if *webMode {
		runWeb(*webPort)
		return
	}

	// CLI mode
	p := magravation.DefaultParams()
	p.BoardSize = *boardSize
	p.MarbleDiameter = *marbleDiam
	p.DiceSize = *diceSize
	p.NumPlayers = *numPlayers
	p.CenterOrigin = *centerOrigin

	board, err := magravation.GenerateBoard(p)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Aggravation Board Generator\n")
	fmt.Printf("  Board size:       %.1f\" x %.1f\"\n", p.BoardSize, p.BoardSize)
	fmt.Printf("  Marble diameter:  %.3f\"\n", p.MarbleDiameter)
	fmt.Printf("  Dice size:        %.3f\"\n", p.DiceSize)
	fmt.Printf("  Players:          %d\n", p.NumPlayers)
	fmt.Printf("  Total holes:      %d\n", len(board.Holes))
	fmt.Printf("  Dice pockets:     %d\n", len(board.DicePockets))
	fmt.Printf("  Text items:       %d\n", len(board.TextItems))
	fmt.Printf("  Hole diameter:    %.4f\"\n", p.HoleDiameter())
	fmt.Printf("  Hole depth:       %.4f\"\n", p.HoleDepth())
	fmt.Printf("  Grid spacing:     %.4f\"\n", p.GridSpacing())
	fmt.Printf("  Origin:           %s\n", originStr(p.CenterOrigin))
	fmt.Println()

	if *previewOnly {
		svg := magravation.GenerateSVG(board)
		svgFile := *output + ".svg"
		if err := os.WriteFile(svgFile, []byte(svg), 0644); err != nil {
			log.Fatalf("Error writing SVG: %v", err)
		}
		fmt.Printf("Preview written to %s\n", svgFile)
		return
	}

	gcode := magravation.GenerateGCode(board)

	if *splitFiles {
		files := map[string]string{
			"_ballend.nc":  gcode.BallEnd,
			"_straight.nc": gcode.Straight,
			"_vbit.nc":     gcode.VBit,
		}
		for suffix, content := range files {
			fname := *output + suffix
			if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
				log.Fatalf("Error writing %s: %v", fname, err)
			}
			lines := strings.Count(content, "\n")
			fmt.Printf("Written: %s (%d lines)\n", fname, lines)
		}
	} else {
		fname := *output + ".nc"
		if err := os.WriteFile(fname, []byte(gcode.Combined), 0644); err != nil {
			log.Fatalf("Error writing %s: %v", fname, err)
		}
		lines := strings.Count(gcode.Combined, "\n")
		fmt.Printf("Written: %s (%d lines)\n", fname, lines)
	}

	// Also generate SVG preview
	svg := magravation.GenerateSVG(board)
	svgFile := *output + ".svg"
	if err := os.WriteFile(svgFile, []byte(svg), 0644); err != nil {
		log.Printf("Warning: could not write SVG preview: %v", err)
	} else {
		fmt.Printf("Preview: %s\n", svgFile)
	}
}

func originStr(center bool) string {
	if center {
		return "center of board"
	}
	return "bottom-left corner"
}

func runWeb(port string) {
	// Find web directory
	webDir := "./web/magravation"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// Try relative to executable
		exe, _ := os.Executable()
		webDir = filepath.Join(filepath.Dir(exe), "web", "magravation")
	}

	fmt.Printf("Magravation Web Server\n")
	fmt.Printf("  Static files: %s\n", webDir)
	fmt.Printf("  Listening on: http://localhost:%s\n", port)

	handler := magravation.NewApp("/")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
