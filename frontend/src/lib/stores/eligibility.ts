import { browser } from '$app/environment';
import { writable, get } from 'svelte/store';
import { api, type EligibilityResponse } from '$lib/api/client';

export const eligibility = writable<EligibilityResponse | null>(null);
export const eligibilityLoading = writable<boolean>(false);
export const eligibilityError = writable<string>('');

export async function loadEligibility(force = false, lang = 'hi'): Promise<EligibilityResponse | null> {
	if (!force && get(eligibility)) return get(eligibility);
	eligibilityLoading.set(true);
	eligibilityError.set('');
	try {
		const res = await api.evaluate(lang);
		eligibility.set(res);
		return res;
	} catch (e) {
		eligibilityError.set(e instanceof Error ? e.message : 'Could not evaluate eligibility');
		return null;
	} finally {
		eligibilityLoading.set(false);
	}
}
