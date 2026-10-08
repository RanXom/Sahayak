import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { api, clearTokens, isAuthenticated, setTokens, type User } from '$lib/api/client';

export const user = writable<User | null>(null);
export const authenticated = writable<boolean>(browser ? isAuthenticated() : false);

export async function login(challengeId: string, otp: string): Promise<void> {
	const res = await api.verifyOtp(challengeId, otp);
	setTokens(res.access_token, res.refresh_token);
	user.set(res.user);
	authenticated.set(true);
}

export async function loadUser(): Promise<void> {
	if (!isAuthenticated()) {
		authenticated.set(false);
		return;
	}
	try {
		const me = await api.me();
		user.set(me);
		authenticated.set(true);
	} catch {
		logout();
	}
}

export function logout(): void {
	api.logout().catch(() => {});
	clearTokens();
	user.set(null);
	authenticated.set(false);
}
