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
	boardDiam := flag.Float64("board-diameter", 26.0, "Board diameter in inches (round board)")
	boardThick := flag.Float64("board-thickness", 0.75, "Board thickness in inches")
	marbleDiam := flag.Float64("marble-diameter", 0.625, "Marble diameter in inches (default 5/8\")")
	numPlayers := flag.Int("players", 4, "Number of players (4 or 6)")

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
	p.BoardDiameter = *boardDiam
	p.BoardThickness = *boardThick
	p.MarbleDiameter = *marbleDiam
	p.NumPlayers = *numPlayers

	board, err := generate.GenerateBoard(p)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Aggravation Board Generator\n")
	fmt.Printf("  Board diameter:   %.1f\" (round)\n", p.BoardDiameter)
	fmt.Printf("  Board thickness:  %.3f\"\n", p.BoardThickness)
	fmt.Printf("  Marble diameter:  %.3f\"\n", p.MarbleDiameter)
	fmt.Printf("  Players:          %d\n", p.NumPlayers)
	fmt.Printf("  Total holes:      %d\n", len(board.Holes))
	fmt.Printf("  Text items:       %d\n", len(board.TextItems))
	fmt.Printf("  Hole diameter:    %.4f\"\n", p.HoleDiameter())
	fmt.Printf("  Hole depth:       %.4f\"\n", p.HoleDepth())
	fmt.Printf("  Grid spacing:     %.4f\"\n", p.GridSpacing())
	fmt.Printf("  Origin:           center of board\n")
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
			"_ballend.nc": gcode.BallEnd,
			"_vbit.nc":    gcode.VBit,
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

// --- Web handlers ---

type generateRequest struct {
	BoardDiameter  float64 `json:"boardDiameter"`
	BoardThickness float64 `json:"boardThickness"`
	MarbleDiameter float64 `json:"marbleDiameter"`
	NumPlayers     int     `json:"numPlayers"`
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
		if req.BoardDiameter > 0 {
			p.BoardDiameter = req.BoardDiameter
		}
		if req.BoardThickness > 0 {
			p.BoardThickness = req.BoardThickness
		}
		if req.MarbleDiameter > 0 {
			p.MarbleDiameter = req.MarbleDiameter
		}
		if req.NumPlayers > 0 {
			p.NumPlayers = req.NumPlayers
		}
		if req.OutputFormat != "" {
			outputFormat = req.OutputFormat
		}
	} else {
		if v := r.URL.Query().Get("boardDiameter"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.BoardDiameter = f
			}
		}
		if v := r.URL.Query().Get("boardThickness"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.BoardThickness = f
			}
		}
		if v := r.URL.Query().Get("marbleDiameter"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.MarbleDiameter = f
			}
		}
		if v := r.URL.Query().Get("numPlayers"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				p.NumPlayers = n
			}
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
		BoardDiameter:  p.BoardDiameter,
		BoardThickness: p.BoardThickness,
		MarbleDiameter: p.MarbleDiameter,
		NumPlayers:     p.NumPlayers,
		OutputFormat:   "combined",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
