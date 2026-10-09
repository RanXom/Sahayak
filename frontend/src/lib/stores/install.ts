import { writable } from 'svelte/store';
import { browser } from '$app/environment';

interface BeforeInstallPromptEvent extends Event {
	prompt: () => Promise<void>;
	userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

let deferred: BeforeInstallPromptEvent | null = null;
let listenersAttached = false;

export const canInstall = writable<boolean>(false);
export const isIos = writable<boolean>(false);
export const isStandalone = writable<boolean>(false);

function detectIos(): boolean {
	if (!browser) return false;
	const ua = navigator.userAgent;
	return /iphone|ipad|ipod/i.test(ua);
}

function detectStandalone(): boolean {
	if (!browser) return false;
	return (
		window.matchMedia('(display-mode: standalone)').matches ||
		(navigator as Navigator & { standalone?: boolean }).standalone === true
	);
}

// attachInstallListener records the browser's install prompt the first time it
// fires, so any component can trigger it from a real user gesture later.
// Safe to call from multiple components; attaches only once.
export function attachInstallListener(): void {
	if (!browser || listenersAttached) return;
	listenersAttached = true;
	isIos.set(detectIos());
	isStandalone.set(detectStandalone());
	window.addEventListener('beforeinstallprompt', (e) => {
		e.preventDefault();
		deferred = e as BeforeInstallPromptEvent;
		canInstall.set(true);
	});
	window.addEventListener('appinstalled', () => {
		deferred = null;
		canInstall.set(false);
		isStandalone.set(true);
	});
}

// triggerInstall shows the browser install dialog. Must run in a user gesture.
export async function triggerInstall(): Promise<'accepted' | 'dismissed' | 'unavailable'> {
	if (!deferred) return 'unavailable';
	await deferred.prompt();
	const { outcome } = await deferred.userChoice;
	if (outcome === 'accepted') {
		deferred = null;
		canInstall.set(false);
	}
	return outcome;
}
