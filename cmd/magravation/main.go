package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"magravation/generate"
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
	p := generate.DefaultParams()
	p.BoardSize = *boardSize
	p.MarbleDiameter = *marbleDiam
	p.DiceSize = *diceSize
	p.NumPlayers = *numPlayers
	p.CenterOrigin = *centerOrigin

	board, err := generate.GenerateBoard(p)
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
		svg := generate.GenerateSVG(board)
		svgFile := *output + ".svg"
		if err := os.WriteFile(svgFile, []byte(svg), 0644); err != nil {
			log.Fatalf("Error writing SVG: %v", err)
		}
		fmt.Printf("Preview written to %s\n", svgFile)
		return
	}

	gcode := generate.GenerateGCode(board)

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
	svg := generate.GenerateSVG(board)
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
	webDir := "./web/magravation"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		exe, _ := os.Executable()
		webDir = filepath.Join(filepath.Dir(exe), "web", "magravation")
	}

	fmt.Printf("Magravation Web Server\n")
	fmt.Printf("  Static files: %s\n", webDir)
	fmt.Printf("  Listening on: http://localhost:%s\n", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/generate", handleGenerate)
	mux.HandleFunc("/api/preview", handlePreview)
	mux.HandleFunc("/api/defaults", handleDefaults)
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// --- Web handlers (standalone mode) ---

type generateRequest struct {
	BoardSize      float64 `json:"boardSize"`
	MarbleDiameter float64 `json:"marbleDiameter"`
	DiceSize       float64 `json:"diceSize"`
	NumPlayers     int     `json:"numPlayers"`
	CenterOrigin   bool    `json:"centerOrigin"`
	OutputFormat   string  `json:"outputFormat"`
}

func parseRequest(r *http.Request) (generate.Params, string, error) {
	p := generate.DefaultParams()
	outputFormat := "combined"

	if r.Method == http.MethodPost {
		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return p, "", fmt.Errorf("invalid JSON: %w", err)
		}
		if req.BoardSize > 0 {
			p.BoardSize = req.BoardSize
		}
		if req.MarbleDiameter > 0 {
			p.MarbleDiameter = req.MarbleDiameter
		}
		if req.DiceSize > 0 {
			p.DiceSize = req.DiceSize
		}
		if req.NumPlayers > 0 {
			p.NumPlayers = req.NumPlayers
		}
		p.CenterOrigin = req.CenterOrigin
		if req.OutputFormat != "" {
			outputFormat = req.OutputFormat
		}
	} else {
		if v := r.URL.Query().Get("boardSize"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.BoardSize = f
			}
		}
		if v := r.URL.Query().Get("marbleDiameter"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.MarbleDiameter = f
			}
		}
		if v := r.URL.Query().Get("diceSize"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.DiceSize = f
			}
		}
		if v := r.URL.Query().Get("numPlayers"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				p.NumPlayers = n
			}
		}
		if r.URL.Query().Get("centerOrigin") == "true" {
			p.CenterOrigin = true
		}
		if v := r.URL.Query().Get("outputFormat"); v != "" {
			outputFormat = v
		}
	}

	return p, outputFormat, nil
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	p, outputFormat, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	board, err := generate.GenerateBoard(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gcode := generate.GenerateGCode(board)

	var content string
	var filename string
	switch outputFormat {
	case "ballend":
		content = gcode.BallEnd
		filename = "aggravation_ballend.nc"
	case "straight":
		content = gcode.Straight
		filename = "aggravation_straight.nc"
	case "vbit":
		content = gcode.VBit
		filename = "aggravation_vbit.nc"
	default:
		content = gcode.Combined
		filename = "aggravation.nc"
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	fmt.Fprint(w, content)
}

func handlePreview(w http.ResponseWriter, r *http.Request) {
	p, _, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	board, err := generate.GenerateBoard(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	svg := generate.GenerateSVG(board)

	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, svg)
}

func handleDefaults(w http.ResponseWriter, r *http.Request) {
	p := generate.DefaultParams()
	resp := generateRequest{
		BoardSize:      p.BoardSize,
		MarbleDiameter: p.MarbleDiameter,
		DiceSize:       p.DiceSize,
		NumPlayers:     p.NumPlayers,
		CenterOrigin:   p.CenterOrigin,
		OutputFormat:   "combined",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
