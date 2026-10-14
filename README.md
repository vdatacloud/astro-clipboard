# Clipped Image Upload and Markup

Paste a screenshot, drag requirement cards onto it to pin numbered markers, and save the marked-up image together with its pin coordinates.

**Live demo:** https://vdatacloud.com/sandbox/astro-clipboard/

Built by [Data Cloud LLC](https://vdatacloud.com).

## What it does

1. The sidebar lists requirements loaded from the backend.
2. Paste an image (Ctrl+V or Cmd+V), or choose a PNG or GIF with the upload button.
3. Drag a requirement onto the image. A colored pin marks the spot, and the requirement's card is placed in a margin beside the image, joined to the pin by a leader line. Pins are stored as percentages of the image, so they stay in place when the window is resized.
4. **Save Draft** uploads the image and the pins mapped so far. Once every requirement has a pin, **Save** uploads the final version. **Refresh** clears the image and starts over.

Each save sends the image, a `pins` JSON array (each pin has the requirement `id` and `text`, its natural-pixel `x`/`y`, and its `pctX`/`pctY` fractions of the image), and `status=draft|final`.

## Repository layout

| Path | What it is |
|---|---|
| `backend/` | Go HTTP server (standard library only). Serves the requirements and saves uploads to `backend/uploads/`. |
| `frontend/` | Astro site. Almost everything is in `src/pages/index.astro`. |
| `frontend/worker/` | Cloudflare Worker that stands in for the Go server on the public demo. |
| `.github/workflows/deploy.yml` | Builds every PR, and deploys the demo on merge to `main` or on a release. |

## Running locally

### Requirements

- **Go 1.27.1 or later** for the backend
- **Node.js 22.12.0 or later** and npm for the frontend
- A browser with clipboard paste and HTML drag-and-drop support (any current desktop browser)

No database or other services are needed.

### 1. Start the backend

```sh
cd backend
go run .
```

This serves http://localhost:8080. Run it from `backend/`, because uploads are written to the relative path `./uploads` (created on first save).

### 2. Start the frontend

In a second terminal:

```sh
cd frontend
npm install
npm run dev
```

Open http://localhost:4321. If that port is taken, run `npx astro dev --port 4322`.

The page calls the backend at `http://localhost:8080` by default, and the backend allows any origin (CORS), so no proxy is needed.

### 3. Try it

Paste a screenshot, drag the requirements onto it, then click Save Draft or Save. Each save writes two files to `backend/uploads/`:

- the image, e.g. `screenshot-1791905982315.png`
- its pins, e.g. `screenshot-1791905982315.json`:

```json
{
  "image": "screenshot-1791905982315.png",
  "status": "final",
  "pins": [{ "id": 1, "text": "…", "x": 412, "y": 96, "pctX": 0.42, "pctY": 0.18 }]
}
```

### Other commands

```sh
# backend/
go build ./...
go vet ./...

# frontend/
npm run build        # static build to frontend/dist/
npm run preview      # serve that build
```

## Configuration

The frontend reads two build-time environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `PUBLIC_API_BASE` | `http://localhost:8080` | Where the page sends `/api/requirements` and `/api/upload` |
| `BASE_PATH` | `/` | The path the site is served under |

The requirements are hardcoded in `backend/main.go` (and copied into `frontend/worker/index.js` for the demo; keep the two in sync).

## API

| Method and path | Description |
|---|---|
| `GET /api/requirements` | JSON array of `{ "id": number, "text": string }` |
| `POST /api/upload` | Multipart form: `image` (file, 10 MB max), and optionally `status` (`draft` or `final`) with `pins` (JSON array). Returns `200` with a text message, or `400` with the reason. |
| `OPTIONS /api/upload` | CORS preflight |

## Demo deployment

Cloudflare can't run the Go server, so the public demo uses a Cloudflare Worker (`frontend/worker/index.js`, configured in `frontend/wrangler.jsonc`) in its place. It serves the built frontend and answers both API routes. **Uploads to the demo are validated and then discarded; nothing is stored.**

The Worker has no public address of its own. The [vdatacloud.com website Worker](https://github.com/vdatacloud/website) forwards `/sandbox/astro-clipboard/*` to it through a service binding and adds the site's security headers.

### Automatic deploys

`.github/workflows/deploy.yml` runs on every pull request, every push to `main` (including merged PRs), every published release, and on manual runs from the Actions tab:

- **check** (always): `go vet` and `go build` for the backend, then `npm ci` and `npm run build:demo` for the frontend.
- **deploy** (not on pull requests): `wrangler deploy` of the build from the check job, recorded in the `demo` GitHub environment.

The workflow needs the following in the repository's **Settings → Secrets and variables → Actions**:

| Name | Kind | Value |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | Secret | A Cloudflare API token made from the **Edit Cloudflare Workers** template, limited to the account below |
| `CLOUDFLARE_ACCOUNT_ID` | Variable | The Cloudflare account that also hosts the `website` Worker |

### Deploying by hand

```sh
cd frontend
npx wrangler login
npm run deploy       # build:demo + wrangler deploy
```

### Testing the demo build locally

```sh
cd frontend
npm run build:demo
npx wrangler dev
```

Then open http://localhost:8787/sandbox/astro-clipboard/.

## Credits

Built by [Data Cloud LLC](https://vdatacloud.com). Colors, the Plus Jakarta Sans font, and the logo follow the Data Cloud design system ([@vdatacloud/cx-commons](https://github.com/vdatacloud/cx-commons)).

## License

[MIT](LICENSE), © 2026 Data Cloud LLC.
