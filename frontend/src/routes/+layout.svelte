<script lang="ts">
	import '$lib/app.css';
	import '@fontsource-variable/manrope';
	import '@fontsource/noto-sans-devanagari/400.css';
	import '@fontsource/noto-sans-devanagari/500.css';
	import '@fontsource/noto-sans-devanagari/600.css';
	import '@fontsource/noto-sans-devanagari/700.css';
	import '@fontsource/noto-sans-tamil/400.css';
	import '@fontsource/noto-sans-tamil/600.css';

	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { language, translate, type Lang } from '$lib/i18n';
	import { authenticated, loadUser, logout } from '$lib/stores/auth';
	import Icon from '$lib/components/Icon.svelte';

	let { children } = $props();

	let online = $state(true);
	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));

	let pathname = $derived(page.url.pathname);
	let isTalk = $derived(pathname.startsWith('/talk'));

	const tabs = [
		{ href: '/', key: 'home', icon: 'home' },
		{ href: '/schemes', key: 'schemes', icon: 'doc' },
		{ href: '/talk', key: 'talk', icon: 'mic', primary: true },
		{ href: '/status', key: 'applications', icon: 'bell' },
		{ href: '/profile', key: 'profile', icon: 'user' }
	];

	function t(key: string): string {
		return translate(lang, key);
	}

	function isActive(href: string): boolean {
		if (href === '/') return pathname === '/';
		return pathname.startsWith(href);
	}

	function switchLang(next: Lang) {
		language.set(next);
	}

	function doLogout() {
		logout();
		goto('/');
	}

	onMount(() => {
		loadUser();
		const update = () => (online = navigator.onLine);
		update();
		window.addEventListener('online', update);
		window.addEventListener('offline', update);
		return () => {
			window.removeEventListener('online', update);
			window.removeEventListener('offline', update);
		};
	});
</script>

<svelte:head>
	<title>{t('app_name')} · {t('tagline')}</title>
</svelte:head>

<a class="skip-link" href="#main">{t('skip_to_content')}</a>
<div class="grain" aria-hidden="true"></div>

{#if !isTalk}
	<header class="appbar">
		<div class="inner">
			<a class="brand" href="/" aria-label={t('app_name')}>
				<span class="mark" aria-hidden="true"></span>
				सहायक
			</a>

			<nav class="navlinks" aria-label="Primary">
				{#each tabs.filter((x) => !x.primary) as tab (tab.href)}
					<a href={tab.href} class:active={isActive(tab.href)}>{t(tab.key)}</a>
				{/each}
			</nav>

			<div class="row" style="gap:6px">
				<select
					class="select lang"
					value={lang}
					onchange={(e) => switchLang((e.currentTarget as HTMLSelectElement).value as Lang)}
					aria-label={t('language')}
				>
					<option value="hi">हिन्दी</option>
					<option value="en">English</option>
					<option value="ta">தமிழ்</option>
				</select>
				{#if $authenticated}
					<button class="btn quiet sm" onclick={doLogout}>{t('logout')}</button>
				{/if}
			</div>
		</div>
	</header>

	{#if !online}
		<div class="shell" style="padding-top:12px">
			<div class="banner">{t('offline')}</div>
		</div>
	{/if}
{/if}

{#key pathname}
	<div class="page-enter">
		{@render children()}
	</div>
{/key}

{#if !isTalk && $authenticated}
	<nav class="tabbar" aria-label="Primary">
		{#each tabs as tab (tab.href)}
			<a href={tab.href} class:active={isActive(tab.href)} class:talk={tab.primary}>
				{#if tab.primary}
					<span class="orbdot"><Icon name="mic" size={18} /></span>
					<span class="sr-only">{t('talk_to_sahayak')}</span>
				{:else}
					<Icon name={tab.icon} size={22} />
					<span>{t(tab.key)}</span>
				{/if}
			</a>
		{/each}
	</nav>
{/if}

<style>
	.row {
		display: flex;
		align-items: center;
	}
	.lang {
		width: auto;
		min-height: 36px;
		padding: 4px 32px 4px 12px;
		font-size: var(--step--1);
		font-weight: 600;
		border-radius: var(--r-pill);
		background-position: right 10px center;
	}
	.tabbar a.talk {
		gap: 2px;
	}
</style>
