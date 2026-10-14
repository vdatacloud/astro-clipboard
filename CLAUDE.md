# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

"Clipboard Mapping Tool": a user pastes a screenshot, then drags requirement cards from a sidebar onto the image to pin annotated markers. The image can then be uploaded to the backend. It has two independent parts with no shared build:

- `backend/`: a single-file Go HTTP server (`main.go`, stdlib only, module `astro-clipboard`, Go 1.27.1) on `:8080`.
- `frontend/`: an Astro 7 site (Node >= 22.12.0). Almost all logic lives in one page, `src/pages/index.astro`, as an inline `<script>`.

## Commands

Backend (run from `backend/`, because uploads are written to the relative `./uploads` path):
```sh
go run .            # start server on http://localhost:8080
go build ./...
go vet ./...
```

Frontend (run from `frontend/`):
```sh
npm install
npm run dev         # http://localhost:4321 (another local Astro project, daml-escrow, often already holds 4321; use `npx astro dev --port 4322`)
npm run build       # output to dist/
npm run preview
npm run astro -- check   # type-check (tsconfig extends astro/tsconfigs/strict)
```
`frontend/AGENTS.md` says to start the dev server in the background with `astro dev --background` and manage it with `astro dev stop|status|logs`.

The repo has no tests, linter config or git repository yet.

## Architecture

**API (backend/main.go)**
- `GET /api/requirements` returns a hardcoded `[]Requirement{ID, Text}` as JSON. There is no database.
- `POST /api/upload` reads a multipart field `image` (10 MB limit) and saves it to `./uploads/<client filename>`. It handles CORS preflight (`OPTIONS`) itself. Every handler sets `Access-Control-Allow-Origin: *` by hand, because the frontend runs on a different origin (port 4321).
- The filename is reduced with `filepath.Base`. When a `status` field is present, the pins are also written to `uploads/<image-name>.json` as `{image, status, pins}`.

**Frontend (frontend/src/pages/index.astro)**
- The backend URL `http://localhost:8080` is hardcoded in the page's `fetch` calls. There is no proxy or env config.
- The page `<style>` is `is:global` on purpose: `.req-item` and `.dropped-badge` elements are created by script, and Astro's default scoped styles never reach them.
- Flow: `loadRequirements()` builds draggable sidebar items. A paste event wraps the clipboard image as `screenshot-<Date.now()>.png` (this name becomes the filename on the server). Dropping a requirement onto `#drop-zone-wrapper` stores a pin in `pinnedRequirements`.
- Pins are stored as **percentages** (`pctX`, `pctY`) of the displayed image (`#preview`), plus `x`/`y` in the image's natural pixels. `renderAllPins()` clears everything and redraws it on every change and on window resize: SVG pin circles and dashed leader lines go in `#overlay-canvas`; the requirement cards (`.dropped-badge`) stack without overlapping in a right margin column created by the wrapper's `padding-right`, which must match `MARGIN_GAP + BADGE_WIDTH` in the script. Drops outside the image are ignored. Keep pins as percentages so they stay aligned with the image when the window is resized.
- Pin colors come from the requirement's index in the `colors` array, and are passed through the drag `dataTransfer` payload as JSON.
- Buttons are SVG icon buttons (each has an `aria-label` and a `title` tooltip): Upload Image, Refresh, Save Draft, Save. Upload Image is always enabled and opens a hidden `#file-input` that accepts PNG or GIF. A chosen file, like a pasted screenshot, goes through `loadImage()` and is renamed with a timestamp (`image-<ts>.<ext>`). Save Draft and Save go through `submit()`, which POSTs to `/api/upload` with the image, a `pins` JSON field and `status=draft|final`. `updateButtons()` runs on every `renderAllPins()`: Save Draft is shown until every requirement has a pin, and then it is hidden and Save is enabled. Refresh clears the screenshot and all pins, then re-fetches the requirements. Results go in `#status-msg` rather than `alert()`.

The sample requirements in `main.go` describe the intended UI behavior (submit button disabled until a paste, dashed empty-state border, timestamped filenames, CORS preflight support).
