import { browser } from '$app/environment';

const ENV_BASE = import.meta.env.VITE_API_URL as string | undefined;

const API_BASE = ENV_BASE && ENV_BASE.length > 0 ? ENV_BASE : '/api/v1';

export interface ApiErrorBody {
	error: { code: string; message: string; details?: Record<string, string> };
}

export class ApiError extends Error {
	code: string;
	status: number;
	details?: Record<string, string>;
	constructor(status: number, body: ApiErrorBody) {
		super(body.error.message);
		this.status = status;
		this.code = body.error.code;
		this.details = body.error.details;
	}
}

interface TokenState {
	accessToken: string | null;
	refreshToken: string | null;
}

const tokens: TokenState = {
	accessToken: browser ? localStorage.getItem('sahayak_access') : null,
	refreshToken: browser ? localStorage.getItem('sahayak_refresh') : null
};

export function setTokens(access: string, refresh: string) {
	tokens.accessToken = access;
	tokens.refreshToken = refresh;
	if (browser) {
		localStorage.setItem('sahayak_access', access);
		localStorage.setItem('sahayak_refresh', refresh);
	}
}

export function clearTokens() {
	tokens.accessToken = null;
	tokens.refreshToken = null;
	if (browser) {
		localStorage.removeItem('sahayak_access');
		localStorage.removeItem('sahayak_refresh');
	}
}

export function isAuthenticated(): boolean {
	return !!tokens.accessToken;
}

async function tryRefresh(): Promise<boolean> {
	if (!tokens.refreshToken) return false;
	const res = await fetch(`${API_BASE}/auth/refresh`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ refresh_token: tokens.refreshToken })
	});
	if (!res.ok) return false;
	const data = await res.json();
	setTokens(data.access_token, data.refresh_token);
	return true;
}

async function request<T>(
	path: string,
	options: RequestInit = {},
	lang = 'hi',
	retry = true
): Promise<T> {
	const headers = new Headers(options.headers);
	headers.set('Accept-Language', lang);
	if (!(options.body instanceof FormData)) {
		headers.set('Content-Type', 'application/json');
	}
	if (tokens.accessToken) {
		headers.set('Authorization', `Bearer ${tokens.accessToken}`);
	}

	const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

	if (res.status === 401 && retry && tokens.refreshToken) {
		if (await tryRefresh()) {
			return request<T>(path, options, lang, false);
		}
		clearTokens();
	}

	if (!res.ok) {
		let body: ApiErrorBody;
		try {
			body = await res.json();
		} catch {
			body = { error: { code: 'UNKNOWN', message: `Request failed (${res.status})` } };
		}
		throw new ApiError(res.status, body);
	}

	if (res.status === 204) return undefined as T;
	const text = await res.text();
	return text ? (JSON.parse(text) as T) : (undefined as T);
}
// ---- Domain types (mirrors the backend API) ----

export interface User {
	id: string;
	role: string;
	preferred_lang: string;
	phone_last4: string;
}

export interface AppConfig {
	app_name: string;
	version: string;
	voice_enabled: boolean;
	llm_enabled: boolean;
	demo_mode: boolean;
	languages: string[];
}
export const api = {
	// Config / capabilities
	getConfig: () => request<AppConfig>('/config'),

	// Auth
	requestOtp: (phone: string) =>
		request<{ challenge_id: string; expires_in: number; demo_otp?: string }>(
			'/auth/otp/request',
			{
				method: 'POST',
				body: JSON.stringify({ phone })
			}
		),
	verifyOtp: (challenge_id: string, otp: string) =>
		request<{ access_token: string; refresh_token: string; expires_in: number; user: User }>(
			'/auth/otp/verify',
			{ method: 'POST', body: JSON.stringify({ challenge_id, otp }) }
		),
	me: () => request<User>('/auth/me'),
	logout: () =>
		request<{ status: string }>('/auth/logout', { method: 'POST', body: JSON.stringify({}) }),
};
export function apiBase(): string {
	return API_BASE;
}