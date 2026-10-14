// Demo backend for vdatacloud.com/sandbox/astro-clipboard. Cloudflare can't run
// the Go server, so this Worker stands in for it: it serves the built frontend
// from dist/ and answers the two API routes. Uploads are validated the way
// backend/main.go validates them, then discarded -- this is a public demo, so
// visitors' screenshots are never stored.
//
// It has no route of its own (workers_dev is off in wrangler.jsonc). The
// website Worker (vdatacloud/website) forwards /sandbox/astro-clipboard/* here
// through a service binding and adds its security headers to the response.

const BASE = '/sandbox/astro-clipboard';
const MAX_UPLOAD_BYTES = 10 << 20;
const IMAGE_TYPES = new Set(['image/png', 'image/gif']);

// Keep in sync with getRequirementsHandler in backend/main.go.
const REQUIREMENTS = [
	{ id: 1, text: 'The submit button must remain disabled until a valid screenshot is pasted.' },
	{ id: 2, text: 'The preview image container must display a dashed border when empty.' },
	{ id: 3, text: 'All uploaded images must automatically generate a timestamped filename.' },
	{ id: 4, text: 'Backend router must properly support CORS preflight OPTIONS requests.' },
];

function text(body, status = 200) {
	return new Response(body, { status, headers: { 'Content-Type': 'text/plain; charset=utf-8' } });
}

async function upload(request) {
	const length = Number(request.headers.get('Content-Length') ?? 0);
	if (length > MAX_UPLOAD_BYTES) return text('File too large', 400);

	let form;
	try {
		form = await request.formData();
	} catch {
		return text('Invalid form data', 400);
	}

	const image = form.get('image');
	if (!(image instanceof File)) return text('Error retrieving the file', 400);
	if (image.size > MAX_UPLOAD_BYTES) return text('File too large', 400);
	if (!IMAGE_TYPES.has(image.type)) return text('Only PNG or GIF images are accepted', 400);

	const status = form.get('status');
	if (status !== null) {
		if (status !== 'draft' && status !== 'final') return text('Invalid status', 400);
		let pins;
		try {
			pins = JSON.parse(String(form.get('pins')));
		} catch {
			return text('Invalid pins', 400);
		}
		if (!Array.isArray(pins)) return text('Invalid pins', 400);
	}

	return text('Upload successful (demo: nothing was stored)');
}

export default {
	async fetch(request, env) {
		const url = new URL(request.url);
		if (url.pathname !== BASE && !url.pathname.startsWith(`${BASE}/`)) {
			return text('Not found', 404);
		}
		const path = url.pathname.slice(BASE.length) || '/';

		if (path === '/api/requirements') {
			if (request.method !== 'GET') return text('Method not allowed', 405);
			return Response.json(REQUIREMENTS);
		}
		if (path === '/api/upload') {
			if (request.method !== 'POST') return text('Method not allowed', 405);
			return upload(request);
		}

		// Astro builds with base=BASE but writes files at the root of dist/.
		url.pathname = path;
		return env.ASSETS.fetch(new Request(url, request));
	},
};
