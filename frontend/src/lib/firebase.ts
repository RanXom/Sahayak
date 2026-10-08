import { browser } from '$app/environment';
import { initializeApp, getApps, type FirebaseApp } from 'firebase/app';
import {
	getAuth,
	RecaptchaVerifier,
	signInWithPhoneNumber,
	type Auth,
	type ConfirmationResult
} from 'firebase/auth';

function env(name: string): string {
	const v = (import.meta.env[name] as string | undefined) ?? '';
	return v.trim();
}

// firebaseEnabled is true when the web config was baked in at build time.
// The backend independently reports firebase_auth in /config; the login page
// requires both before offering this method.
export const firebaseEnabled: boolean =
	browser && env('VITE_FIREBASE_API_KEY') !== '' && env('VITE_FIREBASE_PROJECT_ID') !== '';

let app: FirebaseApp | null = null;
let auth: Auth | null = null;
let verifier: RecaptchaVerifier | null = null;
let confirmation: ConfirmationResult | null = null;

function ensureAuth(): Auth {
	if (auth) return auth;
	if (!app) {
		const existing = getApps();
		app =
			existing.length > 0
				? existing[0]
				: initializeApp({
						apiKey: env('VITE_FIREBASE_API_KEY'),
						authDomain: env('VITE_FIREBASE_AUTH_DOMAIN'),
						projectId: env('VITE_FIREBASE_PROJECT_ID'),
						appId: env('VITE_FIREBASE_APP_ID')
					});
	}
	auth = getAuth(app);
	return auth;
}

// normalizePhone converts user input to E.164 (+91 default for 10 digits).
export function normalizePhone(raw: string): string {
	const digits = raw.replace(/\D/g, '');
	if (raw.trim().startsWith('+')) return '+' + digits;
	if (digits.length === 10) return '+91' + digits;
	return '+' + digits;
}

// sendFirebaseOtp sends the SMS via Firebase (invisible reCAPTCHA) and keeps
// the confirmation handle for the code step. containerId must exist in the DOM.
export async function sendFirebaseOtp(phone: string, containerId: string): Promise<void> {
	const a = ensureAuth();
	if (verifier) {
		try {
			verifier.clear();
		} catch {
			// ignore — replaced below
		}
		verifier = null;
	}
	verifier = new RecaptchaVerifier(a, containerId, { size: 'invisible' });
	confirmation = await signInWithPhoneNumber(a, normalizePhone(phone), verifier);
}

// confirmFirebaseOtp verifies the SMS code and returns the Firebase ID token
// to exchange with our backend.
export async function confirmFirebaseOtp(code: string): Promise<string> {
	if (!confirmation) throw new Error('No verification in progress. Request a code first.');
	const cred = await confirmation.confirm(code.trim());
	confirmation = null;
	return await cred.user.getIdToken();
}

export async function firebaseSignOut(): Promise<void> {
	if (auth) {
		try {
			await auth.signOut();
		} catch {
			// ignore — local session is what matters
		}
	}
}
