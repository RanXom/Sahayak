<script lang="ts">
	import { onMount } from 'svelte';
	import {
		attachInstallListener,
		triggerInstall,
		canInstall,
		isIos,
		isStandalone
	} from '$lib/stores/install';
	import { language, translate, type Lang } from '$lib/i18n';
	import Icon from '$lib/components/Icon.svelte';

	let { compact = false }: { compact?: boolean } = $props();

	let lang = $state<Lang>('hi');
	language.subscribe((v) => (lang = v));

	onMount(() => attachInstallListener());

	async function install() {
		await triggerInstall();
	}
</script>

{#if !$isStandalone}
	{#if $canInstall}
		<button
			class={compact ? 'btn quiet sm' : 'btn ghost block'}
			class:install-btn={!compact}
			onclick={install}
		>
			<Icon name="plus" size={16} /> {translate(lang, 'install_app')}
		</button>
	{:else if $isIos}
		<p class="muted sm install-hint">{translate(lang, 'install_ios_hint')}</p>
	{/if}
{/if}

<style>
	.install-btn {
		margin-top: 12px;
	}
	.install-hint {
		margin-top: 12px;
	}
</style>
