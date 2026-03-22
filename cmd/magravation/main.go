package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"magravation/generate"
)

func main() {
	webMode := flag.Bool("web", false, "Run as web server")
	webPort := flag.String("port", "8080", "Web server port")

	boardDiam := flag.Float64("board-diameter", 0, "Board diameter in inches (0 = auto-calculate minimum)")
	marbleDiam := flag.Float64("marble-diameter", 0.625, "Marble diameter in inches (default 5/8\")")
	numPlayers := flag.Int("players", 4, "Number of players (3-6)")

	output := flag.String("output", "aggravation", "Output file prefix")
	splitFiles := flag.Bool("split", false, "Generate separate files per tool")
	previewOnly := flag.Bool("preview", false, "Generate SVG preview only")

	flag.Parse()

	if *webMode {
		runWeb(*webPort)
		return
	}

	p := generate.DefaultParams()
	p.MarbleDiameter = *marbleDiam
	p.NumPlayers = *numPlayers

	if *boardDiam > 0 {
		p.BoardDiameter = *boardDiam
	} else {
		p.BoardDiameter = math.Ceil(p.MinBoardDiameter())
	}

	board, err := generate.GenerateBoard(p)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Aggravation Board Generator\n")
	fmt.Printf("  Board diameter:   %.1f\" (round, min %.1f\")\n", p.BoardDiameter, p.MinBoardDiameter())
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
	mux.HandleFunc("/api/mindiameter", handleMinDiameter)
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

type generateRequest struct {
	BoardDiameter  float64 `json:"boardDiameter"`
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
		if req.MarbleDiameter > 0 {
			p.MarbleDiameter = req.MarbleDiameter
		}
		if req.NumPlayers > 0 {
			p.NumPlayers = req.NumPlayers
		}
		if req.BoardDiameter > 0 {
			p.BoardDiameter = req.BoardDiameter
		} else {
			p.BoardDiameter = math.Ceil(p.MinBoardDiameter())
		}
		if req.OutputFormat != "" {
			outputFormat = req.OutputFormat
		}
	} else {
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
		if v := r.URL.Query().Get("boardDiameter"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				p.BoardDiameter = f
			}
		} else {
			p.BoardDiameter = math.Ceil(p.MinBoardDiameter())
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
		BoardDiameter:  math.Ceil(p.MinBoardDiameter()),
		MarbleDiameter: p.MarbleDiameter,
		NumPlayers:     p.NumPlayers,
		OutputFormat:   "combined",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type minDiameterRequest struct {
	MarbleDiameter float64 `json:"marbleDiameter"`
	SideAPlayers   int     `json:"sideAPlayers"`
	SideBPlayers   int     `json:"sideBPlayers"`
}

type minDiameterResponse struct {
	MinDiameterA float64 `json:"minDiameterA"`
	MinDiameterB float64 `json:"minDiameterB"`
	MinDiameter  float64 `json:"minDiameter"`
}

func handleMinDiameter(w http.ResponseWriter, r *http.Request) {
	var req minDiameterRequest
	if r.Method == http.MethodPost {
		json.NewDecoder(r.Body).Decode(&req)
	} else {
		if v := r.URL.Query().Get("marbleDiameter"); v != "" {
			req.MarbleDiameter, _ = strconv.ParseFloat(v, 64)
		}
		if v := r.URL.Query().Get("sideAPlayers"); v != "" {
			req.SideAPlayers, _ = strconv.Atoi(v)
		}
		if v := r.URL.Query().Get("sideBPlayers"); v != "" {
			req.SideBPlayers, _ = strconv.Atoi(v)
		}
	}

	if req.MarbleDiameter <= 0 {
		req.MarbleDiameter = 0.625
	}
	if req.SideAPlayers < 3 {
		req.SideAPlayers = 4
	}
	// SideBPlayers = 0 means single-sided
	minA := generate.MinBoardDiameterForPlayers(req.SideAPlayers, req.MarbleDiameter)
	minB := 0.0
	if req.SideBPlayers >= 3 {
		minB = generate.MinBoardDiameterForPlayers(req.SideBPlayers, req.MarbleDiameter)
	}
	minBoth := math.Max(minA, minB)

	resp := minDiameterResponse{
		MinDiameterA: math.Ceil(minA*10) / 10,
		MinDiameterB: math.Ceil(minB*10) / 10,
		MinDiameter:  math.Ceil(minBoth*10) / 10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
