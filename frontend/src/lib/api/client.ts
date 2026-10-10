import { browser } from '$app/environment';

const ENV_BASE = import.meta.env.VITE_API_URL as string | undefined;

const API_BASE = ENV_BASE && ENV_BASE.length > 0 ? ENV_BASE : '/api/v1';

// resolveMediaUrl turns a backend-provided path (e.g. /api/v1/audio/temp/x.wav)
// into a fetchable URL. The backend returns origin-relative paths, but the
// browser may be on a different origin than the API (local dev, split
// subdomains), so prefix the API origin when the base is absolute.
export function resolveMediaUrl(path: string): string {
	if (!path.startsWith('/')) return path;
	if (!browser) return path;
	if (API_BASE.startsWith('http')) {
		try {
			return new URL(API_BASE).origin + path;
		} catch {
			return path;
		}
	}
	return path;
}

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

	// Bound every request: a hung connection otherwise leaves pages on their
	// loading state forever (fetch has no built-in timeout).
	const controller = new AbortController();
	const timer = setTimeout(() => controller.abort(), 90000);
	let res: Response;
	try {
		res = await fetch(`${API_BASE}${path}`, { ...options, headers, signal: controller.signal });
	} catch (e) {
		if (e instanceof DOMException && e.name === 'AbortError') {
			throw new ApiError(0, {
				error: { code: 'TIMEOUT', message: 'Request timed out. Check your connection and retry.' }
			});
		}
		throw e;
	} finally {
		clearTimeout(timer);
	}

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

export interface HelperProfile {
	id: string;
	user_id: string;
	helper_type: string;
	organization?: string;
	display_name?: string;
	languages: string[];
	categories: string[];
	experience?: string;
	is_available: boolean;
	terms_accepted_at?: string | null;
	verified: boolean;
	active: boolean;
}

export interface HelperRegisterInput {
	helper_type: string;
	organization?: string;
	id_number?: string;
	display_name?: string;
	languages: string[];
	categories: string[];
	experience?: string;
	terms_accepted: boolean;
}

export interface AssistRequest {
	id: string;
	citizen_user_id?: string;
	application_id?: string | null;
	scheme_id?: string | null;
	language: string;
	category: string;
	details?: string;
	share_profile_facts: boolean;
	share_documents: boolean;
	status: string;
	matched_helper_user_id?: string | null;
	created_at: string;
	waiting_seconds?: number;
	helper_display_name?: string | null;
}

export interface AssistRequestInput {
	application_id?: string;
	scheme_id?: string;
	language: string;
	category: string;
	details?: string;
	share_profile_facts: boolean;
	share_documents: boolean;
}

export interface AssistSession {
	id: string;
	request_id?: string | null;
	helper_user_id: string;
	citizen_user_id?: string | null;
	status: string;
	created_at: string;
	completed_at?: string | null;
}

export interface AssistMessage {
	id: string;
	session_id: string;
	sender_role: string;
	body: string;
	created_at: string;
	mine?: boolean;
}

export interface AssistSignal {
	id: string;
	session_id: string;
	kind: string;
	payload: Record<string, unknown>;
	created_at: string;
	mine?: boolean;
}

export interface DraftValue {
	field_name: string;
	value: unknown;
	proposed_by: string;
	status: string;
	updated_at: string;
}

export interface AssistContext {
	request: AssistRequest;
	drafts: DraftValue[];
	profile_facts?: Record<string, unknown>;
	documents?: Array<{ document_type: string; has_document: boolean }>;
	application?: { id: string; scheme_id: string; status: string; form_data: Record<string, unknown> };
}

export interface AppConfig {
	app_name: string;
	version: string;
	voice_enabled: boolean;
	llm_enabled: boolean;
	demo_mode: boolean;
	firebase_auth?: boolean;
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

export interface TurnResult {
	conversation_id: string;
	transcript: string;
	detected_language: string;
	response_text: string;
	response_audio_url?: string;
	facts_extracted: Record<string, unknown>;
	profile_snapshot: Record<string, unknown>;
	follow_up?: { question_about: string; is_last_question: boolean };
	ready_for_review: boolean;
	turn_count: number;
}

export interface FormFieldView {
	name: string;
	label: string;
	type: string;
	required: boolean;
	value?: unknown;
	auto_filled: boolean;
	options?: unknown;
}

export interface ApplicationView {
	application: {
		id: string;
		status: string;
		has_pdf: boolean;
		form_data: Record<string, unknown>;
	};
	scheme: Scheme;
	form_fields: FormFieldView[];
	auto_filled: Record<string, unknown>;
	missing_fields: string[];
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
	verifyOtp: (challenge_id: string, otp: string, role?: string) =>
		request<{ access_token: string; refresh_token: string; expires_in: number; user: User }>(
			'/auth/otp/verify',
			{ method: 'POST', body: JSON.stringify({ challenge_id, otp, role }) }
		),
	firebaseLogin: (id_token: string, role?: string) =>
		request<{ access_token: string; refresh_token: string; expires_in: number; user: User }>(
			'/auth/firebase',
			{ method: 'POST', body: JSON.stringify({ id_token, role }) }
		),
	refreshSession: () => {
		const rt =
			typeof localStorage !== 'undefined' ? localStorage.getItem('sahayak_refresh') : null;
		return request<{ access_token: string; refresh_token: string; expires_in: number; user: User }>(
			'/auth/refresh',
			{ method: 'POST', body: JSON.stringify({ refresh_token: rt }) }
		);
	},
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

	// Conversation
	startConversation: (lang = 'hi') =>
		request<{ id: string; messages: Array<{ role: string; content_text: string }> }>(
			'/conversations',
			{ method: 'POST', body: JSON.stringify({}) },
			lang
		),
	sendMessage: (id: string, text: string, language: string) =>
		request<TurnResult>(`/conversations/${id}/message`, {
			method: 'POST',
			body: JSON.stringify({ text, language })
		}),
	sendVoice: (id: string, audio: Blob, language: string) => {
		const form = new FormData();
		form.append('audio', audio, 'audio.webm');
		form.append('language', language);
		return request<TurnResult>(`/conversations/${id}/voice`, { method: 'POST', body: form });
	},

	// Eligibility
	evaluate: (lang = 'hi') =>
		request<EligibilityResponse>('/eligibility/evaluate', { method: 'POST', body: JSON.stringify({}) }, lang),
	scheme: (id: string) => request<Scheme>(`/schemes/${id}`),
	schemeDocuments: (id: string) =>
		request<{ scheme_id: string; documents: DocumentStatus[]; missing_count: number; all_present: boolean }>(
			`/schemes/${id}/documents`
		),
	documentTypes: () => request<{ document_types: unknown[] }>('/documents/types'),

	// Applications
	createApplication: (scheme_id: string, lang = 'hi') =>
		request<ApplicationView>(
			'/applications',
			{
				method: 'POST',
				body: JSON.stringify({ scheme_id })
			},
			lang
		),
	getApplication: (id: string, lang = 'hi') =>
		request<ApplicationView>(`/applications/${id}`, {}, lang),
	listApplications: () =>
		request<{ applications: Array<{ id: string; scheme_id: string; status: string; has_pdf: boolean }> }>(
			'/applications'
		),
	updateApplicationForm: (id: string, form_data: Record<string, unknown>) =>
		request<{ id: string; status: string }>(`/applications/${id}/form`, {
			method: 'PUT',
			body: JSON.stringify({ form_data })
		}),
	generatePdf: (id: string, lang = 'hi') =>
		request<{ id: string; status: string; has_pdf: boolean }>(
			`/applications/${id}/pdf`,
			{
				method: 'POST'
			},
			lang
		),
	recordSubmission: (id: string, submission_ref: string) =>
		request<unknown>(`/applications/${id}/submit`, {
			method: 'POST',
			body: JSON.stringify({ submission_ref })
		}),

	// Notifications
	getPreferences: () =>
		request<{ preferences: Array<{ channel: string; enabled: boolean }> }>(
			'/notifications/preferences'
		),
	updatePreferences: (channel: string, enabled: boolean) =>
		request<{ preferences: Array<{ channel: string; enabled: boolean }> }>(
			'/notifications/preferences',
			{ method: 'PUT', body: JSON.stringify({ channel, enabled }) }
		),
	listNotifications: () => request<{ notifications: unknown[] }>('/notifications'),

	// Helper profile
	getHelperProfile: () => request<HelperProfile>('/helper/profile'),
	registerHelper: (input: HelperRegisterInput) =>
		request<HelperProfile>('/helper/register', {
			method: 'POST',
			body: JSON.stringify(input)
		}),
	updateHelperProfile: (input: HelperRegisterInput) =>
		request<HelperProfile>('/helper/profile', {
			method: 'PUT',
			body: JSON.stringify(input)
		}),
	setAvailability: (is_available: boolean) =>
		request<{ is_available: boolean }>('/helper/availability', {
			method: 'PUT',
			body: JSON.stringify({ is_available })
		}),

	// Assistance marketplace
	createAssistRequest: (input: AssistRequestInput) =>
		request<AssistRequest>('/assistance/requests', {
			method: 'POST',
			body: JSON.stringify(input)
		}),
	myAssistRequests: () => request<{ requests: AssistRequest[] }>('/assistance/requests'),
	getAssistRequest: (id: string) => request<AssistRequest>(`/assistance/requests/${id}`),
	joinAssistRequest: (id: string, scope: { share_profile_facts: boolean; share_documents: boolean }) =>
		request<AssistSession>(`/assistance/requests/${id}/join`, {
			method: 'POST',
			body: JSON.stringify(scope)
		}),
	cancelAssistRequest: (id: string) =>
		request<AssistRequest>(`/assistance/requests/${id}/cancel`, { method: 'POST', body: JSON.stringify({}) }),
	assistQueue: () =>
		request<{ requests: AssistRequest[]; has_active_session: boolean }>('/assistance/queue'),
	acceptAssistRequest: (id: string) =>
		request<AssistSession>(`/assistance/requests/${id}/accept`, {
			method: 'POST',
			body: JSON.stringify({})
		}),
	myAssistSessions: () => request<{ sessions: AssistSession[] }>('/assistance/sessions'),
	getAssistSession: (id: string) => request<AssistSession>(`/assistance/sessions/${id}`),
	assistContext: (id: string) => request<AssistContext>(`/assistance/sessions/${id}/context`),
	endAssistSession: (id: string) =>
		request<AssistSession>(`/assistance/sessions/${id}/end`, {
			method: 'POST',
			body: JSON.stringify({})
		}),
	sendAssistMessage: (id: string, body: string, client_msg_id: string) =>
		request<AssistMessage>(`/assistance/sessions/${id}/messages`, {
			method: 'POST',
			body: JSON.stringify({ body, client_msg_id })
		}),
	listAssistMessages: (id: string, since?: string) =>
		request<{ messages: AssistMessage[] }>(
			`/assistance/sessions/${id}/messages${since ? `?since=${encodeURIComponent(since)}` : ''}`
		),
	sendAssistSignal: (id: string, kind: string, payload: Record<string, unknown>) =>
		request<AssistSignal>(`/assistance/sessions/${id}/signals`, {
			method: 'POST',
			body: JSON.stringify({ kind, payload })
		}),
	listAssistSignals: (id: string, since?: string) =>
		request<{ signals: AssistSignal[] }>(
			`/assistance/sessions/${id}/signals${since ? `?since=${encodeURIComponent(since)}` : ''}`
		),
	proposeDraftValue: (id: string, field_name: string, value: unknown) =>
		request<DraftValue>(`/assistance/sessions/${id}/drafts`, {
			method: 'PUT',
			body: JSON.stringify({ field_name, value })
		}),
	confirmDraftValue: (id: string, field_name: string) =>
		request<DraftValue>(`/assistance/sessions/${id}/drafts/confirm`, {
			method: 'POST',
			body: JSON.stringify({ field_name })
		}),
	applyDraftValues: (id: string) =>
		request<{ applied: Record<string, unknown> }>(`/assistance/sessions/${id}/drafts/apply`, {
			method: 'POST',
			body: JSON.stringify({})
		})
};

export function apiBase(): string {
	return API_BASE;
}
