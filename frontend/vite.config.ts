import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	server: {
		proxy: {
			'/api': {
				target:
					(globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env
						?.PUBLIC_API_URL ?? 'http://localhost:8080',
				changeOrigin: true
			}
		}
	},
	plugins: [sveltekit()]
});
