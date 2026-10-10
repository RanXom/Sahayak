import { sveltekit } from '@sveltejs/kit/vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';
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
	plugins: [
		sveltekit(),
		SvelteKitPWA({
			registerType: 'autoUpdate',
			strategies: 'generateSW',
			manifest: {
				name: 'Sahayak',
				short_name: 'Sahayak',
				description: 'Discover government welfare schemes you are eligible for, by voice.',
				start_url: '/',
				display: 'standalone',
				orientation: 'portrait',
				background_color: '#F4F4F1',
				theme_color: '#0E6E55',
				lang: 'hi',
				id: '/',
				icons: [
					{ src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
					{ src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
					{
						src: '/icons/maskable-512.png',
						sizes: '512x512',
						type: 'image/png',
						purpose: 'maskable'
					},
					{ src: '/icons/icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'any' }
				]
			},
			workbox: {
				globPatterns: ['**/*.{js,css,html,svg,png,woff2}'],
				navigateFallback: '/',
				runtimeCaching: [
					{
						urlPattern: ({ url }) => url.pathname.startsWith('/api/v1/locations'),
						handler: 'StaleWhileRevalidate',
						options: { cacheName: 'sahayak-locations' }
					}
				]
			},
			devOptions: { enabled: false }
		})
	]
});
