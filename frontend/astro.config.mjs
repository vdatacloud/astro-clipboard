// @ts-check
import { defineConfig } from 'astro/config';

// https://astro.build/config
export default defineConfig({
	// The demo is served under vdatacloud.com/sandbox/astro-clipboard (see `npm run build:demo`).
	base: process.env.BASE_PATH ?? '/',
});
