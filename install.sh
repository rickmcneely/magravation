#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Magravation – CharmToolWeb Integration Installer
#
# Integrates the Magravation web app into the CharmToolWeb
# webserver, then optionally deploys to the Hostinger server.
#
# What it does:
#   1. Creates internal/magravation/handler.go in CharmToolWeb
#   2. Copies web/magravation/ static files to CharmToolWeb
#   3. Adds magravation to CharmToolWeb go.mod (replace directive)
#   4. Registers the /magravation/ route in cmd/webserver/main.go
#   5. Runs go mod tidy to verify the build
#   6. Optionally deploys via deploy.sh
#
# Usage:
#   ./install.sh                # integrate only
#   ./install.sh --deploy       # integrate + deploy to production
# ============================================================

MAGRAVATION_DIR="$(cd "$(dirname "$0")" && pwd)"
CHARMTOOL_DIR="/home/zditech/CharmToolWeb"

DEPLOY=false
if [[ "${1:-}" == "--deploy" ]]; then
    DEPLOY=true
fi

echo "==> Magravation → CharmToolWeb Installer"
echo "    Magravation: ${MAGRAVATION_DIR}"
echo "    CharmToolWeb: ${CHARMTOOL_DIR}"
echo ""

# Verify CharmToolWeb exists
if [[ ! -f "${CHARMTOOL_DIR}/go.mod" ]]; then
    echo "ERROR: CharmToolWeb not found at ${CHARMTOOL_DIR}"
    exit 1
fi

# ──────────────────────────────────────────────
# Step 1: Create the HTTP handler wrapper
# ──────────────────────────────────────────────
echo "==> Step 1: Creating handler in CharmToolWeb/internal/magravation/"
mkdir -p "${CHARMTOOL_DIR}/internal/magravation"

cat > "${CHARMTOOL_DIR}/internal/magravation/handler.go" << 'GOEOF'
// Package magravation provides an HTTP handler for the Magravation
// Aggravation board game G-code generator web app.
package magravation

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"

	"magravation/generate"
)

type GenerateRequest struct {
	BoardDiameter  float64 `json:"boardDiameter"`
	MarbleDiameter float64 `json:"marbleDiameter"`
	NumPlayers     int     `json:"numPlayers"`
	OutputFormat   string  `json:"outputFormat"`
	DrawBorder     bool    `json:"drawBorder"`
	CornerOrigin   bool    `json:"cornerOrigin"`
	FontSize       float64 `json:"fontSize"`
}

func NewApp(staticDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/generate", handleGenerate)
	mux.HandleFunc("/api/preview", handlePreview)
	mux.HandleFunc("/api/defaults", handleDefaults)
	mux.HandleFunc("/api/mindiameter", handleMinDiameter)
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	return mux
}

func NewAppDefault() http.Handler {
	return NewApp(filepath.Join(".", "web", "magravation"))
}

func parseRequest(r *http.Request) (generate.Params, string, error) {
	p := generate.DefaultParams()
	outputFormat := "combined"

	if r.Method == http.MethodPost {
		var req GenerateRequest
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
		p.DrawBorder = req.DrawBorder
		p.CornerOrigin = req.CornerOrigin
		if req.FontSize > 0 {
			p.TextHeight = req.FontSize
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
	var content, filename string
	switch outputFormat {
	case "ballend":
		content, filename = gcode.BallEnd, "wahoo_ballend.nc"
	case "vbit":
		content, filename = gcode.VBit, "wahoo_vbit.nc"
	default:
		content, filename = gcode.Combined, "aggravation.nc"
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
	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, generate.GenerateSVG(board))
}

func handleDefaults(w http.ResponseWriter, r *http.Request) {
	p := generate.DefaultParams()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateRequest{
		BoardDiameter:  math.Ceil(p.MinBoardDiameter()),
		MarbleDiameter: p.MarbleDiameter,
		NumPlayers:     p.NumPlayers,
		OutputFormat:   "combined",
	})
}

type minDiamReq struct {
	MarbleDiameter float64 `json:"marbleDiameter"`
	SideAPlayers   int     `json:"sideAPlayers"`
	SideBPlayers   int     `json:"sideBPlayers"`
}

func handleMinDiameter(w http.ResponseWriter, r *http.Request) {
	var req minDiamReq
	if r.Method == http.MethodPost {
		json.NewDecoder(r.Body).Decode(&req)
	}
	if req.MarbleDiameter <= 0 { req.MarbleDiameter = 0.625 }
	if req.SideAPlayers < 3 { req.SideAPlayers = 4 }
	if req.SideBPlayers < 3 { req.SideBPlayers = 6 }
	minA := generate.MinBoardDiameterForPlayers(req.SideAPlayers, req.MarbleDiameter)
	minB := generate.MinBoardDiameterForPlayers(req.SideBPlayers, req.MarbleDiameter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"minDiameterA": math.Ceil(minA*10) / 10,
		"minDiameterB": math.Ceil(minB*10) / 10,
		"minDiameter":  math.Ceil(math.Max(minA, minB)*10) / 10,
	})
}
GOEOF
echo "    Created handler.go"

# ──────────────────────────────────────────────
# Step 2: Copy static web files
# ──────────────────────────────────────────────
echo "==> Step 2: Copying web/magravation/ static files"
mkdir -p "${CHARMTOOL_DIR}/web/magravation"
cp -r "${MAGRAVATION_DIR}/web/magravation/"* "${CHARMTOOL_DIR}/web/magravation/"
echo "    Copied to ${CHARMTOOL_DIR}/web/magravation/"

# ──────────────────────────────────────────────
# Step 3: Add magravation to go.mod
# ──────────────────────────────────────────────
echo "==> Step 3: Updating go.mod"
cd "${CHARMTOOL_DIR}"

# Add require if not present
if ! grep -q 'magravation' go.mod; then
    # Add require
    sed -i '/^require (/a\\tmagravation v0.0.0' go.mod
    # Add replace directive
    echo "" >> go.mod
    echo "replace magravation => ../magravation" >> go.mod
    echo "    Added magravation dependency to go.mod"
else
    echo "    magravation already in go.mod (skipped)"
fi

# ──────────────────────────────────────────────
# Step 4: Register route in cmd/webserver/main.go
# ──────────────────────────────────────────────
echo "==> Step 4: Registering /magravation/ route in webserver"
MAIN_GO="${CHARMTOOL_DIR}/cmd/webserver/main.go"

# Add import if not present
if ! grep -q '"charmtool/internal/magravation"' "${MAIN_GO}"; then
    # Add import line after the last existing charmtool import
    sed -i '/"charmtool\/internal\/mlint"/a\\t"charmtool/internal/magravation"' "${MAIN_GO}"
    echo "    Added import"
else
    echo "    Import already present (skipped)"
fi

# Add route mount if not present
if ! grep -q '/magravation/' "${MAIN_GO}"; then
    # Insert the magravation mount block after the mlint block
    sed -i '/apps = append(apps, App{/{N;N;N;N;}' "${MAIN_GO}"  # normalize

    # Find the line with "Landing page" comment and insert before it
    MOUNT_BLOCK='	// --- Magravation mounted at /magravation/ ---\
	magravationApp := magravation.NewAppDefault()\
	mux.Handle("/magravation/", http.StripPrefix("/magravation", magravationApp))\
	apps = append(apps, App{\
		Name:        "Magravation",\
		Path:        "/magravation/",\
		Description: "Generate Aggravation board game G-code for Masso CNC router",\
	})\
'
    sed -i "/\/\/ Landing page/i\\${MOUNT_BLOCK}" "${MAIN_GO}"
    echo "    Added route mount"
else
    echo "    Route already registered (skipped)"
fi

# ──────────────────────────────────────────────
# Step 5: Build verification
# ──────────────────────────────────────────────
echo "==> Step 5: Running go mod tidy and build verification"
cd "${CHARMTOOL_DIR}"
go mod tidy
go build ./cmd/webserver/
echo "    Build successful!"

# Clean up build artifact
rm -f webserver

# ──────────────────────────────────────────────
# Step 6: Deploy (optional)
# ──────────────────────────────────────────────
if [[ "$DEPLOY" == true ]]; then
    echo ""
    echo "==> Step 6: Deploying to production..."
    cd "${CHARMTOOL_DIR}"
    ./deploy.sh
else
    echo ""
    echo "==> Installation complete!"
    echo ""
    echo "    To test locally:"
    echo "      cd ${CHARMTOOL_DIR}"
    echo "      go run ./cmd/webserver"
    echo "      # Visit http://localhost:8080/magravation/"
    echo ""
    echo "    To deploy to production:"
    echo "      cd ${CHARMTOOL_DIR}"
    echo "      ./deploy.sh"
    echo "    Or re-run this script:"
    echo "      ./install.sh --deploy"
fi
