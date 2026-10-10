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
	import InstallButton from '$lib/components/InstallButton.svelte';

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
				<svg
					class="mark"
					width="24"
					height="26"
					viewBox="0 0 40 43"
					aria-hidden="true"
					focusable="false"
				>
					<g id="logogram" style="opacity: 1;"><path fill="#0E6E55" d="M20 28.4356C16.5306 28.4357 13.3035 30.2166 11.4521 33.1512L9.1543 36.7934L10.0371 37.3511C10.3132 37.5101 10.5937 37.6622 10.8779 37.8082L13.1436 34.2188C14.6285 31.8648 17.2171 30.437 20 30.4369C22.783 30.4369 25.3714 31.8647 26.8564 34.2188L29.1211 37.8082C29.4053 37.6623 29.6858 37.5101 29.9619 37.3511L30.8457 36.7934L28.5479 33.1512C26.6964 30.2165 23.4695 28.4356 20 28.4356ZM20.123 36.4428L20 36.4379C19.3855 36.4381 18.8239 36.774 18.5312 37.3072L18.4766 37.4166L17.3691 39.834C18.0662 39.9256 18.7744 39.9821 19.4922 40L20 38.8914L20.5068 40C21.2249 39.9821 21.9335 39.9256 22.6309 39.834L21.5244 37.4166L21.4697 37.3072C21.1966 36.8096 20.689 36.4842 20.123 36.4428ZM21 12.4063V0H19V12.4063L15.7891 0.422921L13.8574 0.940585L17.0674 12.923L10.8662 2.18005L9.13379 3.18021L15.335 13.9222L6.56543 5.15124L5.15039 6.56651L13.9199 15.3375L3.17969 9.13529L2.17969 10.868L12.9209 17.0702L0.94043 13.8597L0.422852 15.7917L12.4053 19.0031H0V21.0035H12.4053L0.422852 24.2149L0.94043 26.1469L12.9209 22.9354L2.17969 29.1386L3.17969 30.8713L13.9199 24.6681L5.15039 33.4401L6.56543 34.8554L13.1216 28.2964C13.684 27.7337 14 26.9706 14 26.175V20.435C14 17.1209 16.6865 14.4342 20 14.434C23.3137 14.434 26 17.1208 26 20.435V26.1759C26 26.9716 26.316 27.7347 26.8785 28.2974L33.4346 34.8554L34.8496 33.4401L26.0801 24.6691L36.8203 30.8713L37.8203 29.1386L27.0791 22.9354L39.0596 26.1469L39.5771 24.2149L27.5957 21.0035H40V19.0031H27.5947L39.5771 15.7917L39.0596 13.8597L27.0791 17.0692L37.8203 10.868L36.8203 9.13529L26.0781 15.3375L34.8496 6.56554L33.4355 5.15124L24.6641 13.9232L30.8662 3.18021L29.1338 2.18005L22.9316 12.923L26.1426 0.940585L24.2109 0.422921L21 12.4063ZM20 32.4372C18.0122 32.4374 16.1792 33.5112 15.207 35.2453L13.2021 38.8201C13.8407 39.051 14.4942 39.2501 15.1611 39.4159L16.9512 36.224C17.5694 35.1209 18.7357 34.4377 20 34.4376C21.2645 34.4376 22.4305 35.1208 23.0488 36.224L24.8379 39.4159C25.5048 39.2502 26.1583 39.0509 26.7969 38.8201L24.793 35.2453C23.8208 33.5111 21.9879 32.4372 20 32.4372Z"/></g>
				</svg>
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
				<InstallButton compact />
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
