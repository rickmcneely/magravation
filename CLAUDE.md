# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Does

Magravation generates G-code files for cutting an Aggravation marble board game on a MASSO G3 CNC router. The board is round and double-sided: 4-player cross layout on one side, 6-player star layout on the other. Two CNC tools are used: 1/4" ball end mill for marble holes (helical bore) and 60° V-bit for text engraving.

## Build & Run

```bash
go build ./cmd/magravation/          # build CLI
./magravation -players 4             # generate 4-player G-code + SVG
./magravation -players 6             # generate 6-player G-code + SVG
./magravation -preview -output test  # SVG preview only
./magravation -split                 # separate .nc files per tool
./magravation -web -port 8080        # standalone web server
```

No external dependencies — standard library only.

## Architecture

**`generate/`** — Public library package (importable by CharmToolWeb):
- `params.go` — `Params` struct, `DefaultParams()`, validation, computed values (hole diameter/depth/spacing)
- `board.go` — `GenerateBoard()` dispatches to `generate4Player()` or `generate6Player()`. Both use `rotatePoint()` to create symmetric layouts from a single arm template. The 6-player version starts flanking track columns at r=3 (not r=1) to avoid overlap at 60° arm spacing, with connecting holes bridging adjacent arms.
- `gcode.go` — `GenerateGCode()` produces `GCodeOutput` (Combined, BallEnd, VBit). `GenerateSVG()` renders a round board preview. Marble holes use helical boring (`G3` arcs with Z descent).
- `font.go` — Stroke font (single-line glyphs as polylines) for V-bit text engraving. `TextToStrokes()` converts strings to physical tool paths.

**`cmd/magravation/`** — CLI entry point with two modes: command-line G-code generation and standalone web server (`-web`). The web handlers duplicate the CharmToolWeb handler pattern inline.

**`web/magravation/`** — Single-page web UI. Uses relative URLs for API calls (`api/generate`, `api/preview`) so it works under any mount prefix.

## CharmToolWeb Integration

This project follows the same integration pattern as `flaggen`:
- Core logic in public `generate/` package (not `internal/`)
- `install.sh` automates integration into `/home/zditech/CharmToolWeb`:
  1. Creates `CharmToolWeb/internal/magravation/handler.go` (HTTP wrapper importing `magravation/generate`)
  2. Copies static files to `CharmToolWeb/web/magravation/`
  3. Adds `replace magravation => ../magravation` to `go.mod`
  4. Patches `cmd/webserver/main.go` with import + route mount at `/magravation/`
  5. Runs `go mod tidy` + build verification
- `./install.sh --deploy` also triggers CharmToolWeb's `deploy.sh`

## Key Design Constraints

- Origin is always board center (for flip alignment on double-sided board)
- Board thickness validation prevents holes from opposite sides breaking through (2 × depth < thickness - 1/16")
- Grid spacing = marble_diameter + clearance (1/32") + wall_min (3/16")
- Arm tips reach ~7.07 grid units from center — board radius must accommodate this
- The `generate/` package must stay public (not under `internal/`) so CharmToolWeb can import it via replace directive
