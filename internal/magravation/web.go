package magravation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// NewApp returns an http.Handler for the web version.
// cookiePath is the mount prefix (e.g., "/magravation/").
func NewApp(cookiePath string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/generate", handleGenerate)
	mux.HandleFunc("/api/preview", handlePreview)
	mux.HandleFunc("/api/defaults", handleDefaults)
	mux.Handle("/", http.FileServer(http.Dir("./web/magravation")))

	return mux
}

type generateRequest struct {
	BoardSize      float64 `json:"boardSize"`
	MarbleDiameter float64 `json:"marbleDiameter"`
	DiceSize       float64 `json:"diceSize"`
	NumPlayers     int     `json:"numPlayers"`
	CenterOrigin   bool    `json:"centerOrigin"`
	OutputFormat   string  `json:"outputFormat"` // "combined", "ballend", "straight", "vbit"
}

func parseRequest(r *http.Request) (Params, string, error) {
	p := DefaultParams()
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
		// GET with query params
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

	board, err := GenerateBoard(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	gcode := GenerateGCode(board)

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

	board, err := GenerateBoard(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	svg := GenerateSVG(board)

	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, svg)
}

func handleDefaults(w http.ResponseWriter, r *http.Request) {
	p := DefaultParams()
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
