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

export interface StateInfo {
	id: number;
	code: string;
	name_en: string;
	name_hi: string;
	is_ut: boolean;
	coverage: { level: string; scheme_count: number; note?: string };
}

export interface District {
	id: number;
	state_id: number;
	name_en: string;
	name_hi: string;
}

export interface ProfileDocument {
	document_type: string;
	has_document: boolean;
}

export interface ProfileView {
	id: string;
	state?: StateInfo;
	district?: District;
	facts: Record<string, unknown>;
	documents: ProfileDocument[];
	completed: boolean;
	updated_at: string;
}

export interface RuleResult {
	rule: string;
	status: 'pass' | 'fail' | 'unknown' | 'error';
	description_en?: string;
	description_hi?: string;
	missing_fields?: string[];
}

export interface DocumentStatus {
	type: string;
	name: string;
	mandatory: boolean;
	user_has: boolean;
	status: string;
	acquisition?: {
		issuing_authority?: string;
		steps?: string;
		typical_time?: string;
		typical_fee?: string;
		source_url?: string;
		nearest_offices?: Array<{ name: string; address?: string }>;
	};
}

export interface Scheme {
	id: string;
	slug: string;
	name_en: string;
	name_hi: string;
	description_en?: string;
	description_hi?: string;
	benefit_type: string;
	benefit_amount?: string;
	scope: string;
	ministry?: string;
	official_url?: string;
	application_url?: string;
	deadline?: string | null;
	verified_at?: string | null;
}

export interface EligibilityResult {
	scheme: Scheme;
	status: 'eligible' | 'potentially_eligible' | 'not_eligible';
	match_score: number;
	rule_results: RuleResult[];
	missing_fields?: string[];
	explanation: string;
	documents?: {
		documents: DocumentStatus[];
		missing_count: number;
		all_present: boolean;
	};
}

export interface EligibilityResponse {
	evaluated_at: string;
	profile_completeness: number;
	coverage_note: string;
	results: EligibilityResult[];
	summary: Record<string, number>;
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

	// Locations
	states: () => request<{ states: StateInfo[] }>('/locations/states'),
	districts: (stateId: number) =>
		request<{ districts: District[] }>(`/locations/states/${stateId}/districts`),

	// Profile
	getProfile: () => request<ProfileView>('/profile'),
	updateProfile: (payload: { state_id?: number; district_id?: number; facts: Record<string, unknown> }) =>
		request<ProfileView>('/profile', { method: 'PUT', body: JSON.stringify(payload) }),
	mergeFacts: (facts: Record<string, unknown>, state_id?: number, district_id?: number) =>
		request<ProfileView>('/profile/facts', {
			method: 'PATCH',
			body: JSON.stringify({ facts, state_id, district_id })
		}),
	getDocuments: () =>
		request<{ documents: ProfileDocument[] }>('/profile/documents'),
	updateDocuments: (documents: ProfileDocument[]) =>
		request<{ documents: ProfileDocument[] }>('/profile/documents', {
			method: 'PUT',
			body: JSON.stringify({ documents })
		}),

	// Eligibility
	evaluate: (lang = 'hi') =>
		request<EligibilityResponse>('/eligibility/evaluate', { method: 'POST', body: JSON.stringify({}) }, lang),
	scheme: (id: string) => request<Scheme>(`/schemes/${id}`),
	schemeDocuments: (id: string) =>
		request<{ scheme_id: string; documents: DocumentStatus[]; missing_count: number; all_present: boolean }>(
			`/schemes/${id}/documents`
		),
	documentTypes: () => request<{ document_types: unknown[] }>('/documents/types'),
};

export function apiBase(): string {
	return API_BASE;
}