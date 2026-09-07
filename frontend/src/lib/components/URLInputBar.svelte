<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { settingsStore } from '../stores/settings'
  import { Search, KeyRound, Sparkles, Clipboard, Loader2 } from 'lucide-svelte'

  export let loading: boolean = false
  export let currentURL: string = ''

  const dispatch = createEventDispatcher<{
    inspect: string
    openTokenModal: void
  }>()

  let inputValue: string = currentURL

  function handleSubmit() {
    if (!inputValue.trim() || loading) return
    dispatch('inspect', inputValue.trim())
  }

  async function handlePaste() {
    try {
      const text = await navigator.clipboard.readText()
      if (text) {
        inputValue = text.trim()
        handleSubmit()
      }
    } catch (e) {
      console.error('Failed to read clipboard', e)
    }
  }

  function pickPreset(preset: string) {
    inputValue = preset
    handleSubmit()
  }

  const quickPresets = [
    { label: 'FLUX.1-schnell', repo: 'black-forest-labs/FLUX.1-schnell' },
    { label: 'FLUX.1-dev', repo: 'black-forest-labs/FLUX.1-dev' },
    { label: 'SDXL Base', repo: 'stabilityai/stable-diffusion-xl-base-1.0' },
    { label: 'Comfy-Org FLUX Split', repo: 'Comfy-Org/flux1-dev' },
  ]
</script>

<div class="bg-dark-850 border border-dark-700/60 rounded-2xl p-4 shadow-xl backdrop-blur-md">
  <div class="flex items-center justify-between mb-2 px-1">
    <label for="hf-url-input" class="text-xs font-semibold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
      <Sparkles class="w-3.5 h-3.5 text-accent-indigo" />
      Hugging Face URL or Repo Identifier
    </label>

    <!-- Token Status Badge -->
    <button
      type="button"
      on:click={() => dispatch('openTokenModal')}
      class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors border {$settingsStore?.hfToken ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 hover:bg-emerald-500/20' : 'bg-slate-800 text-slate-400 border-slate-700 hover:text-slate-200 hover:border-slate-600'}"
    >
      <KeyRound class="w-3 h-3" />
      {#if $settingsStore?.hfToken}
        <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
        <span>Token Active</span>
      {:else}
        <span class="w-1.5 h-1.5 rounded-full bg-amber-400"></span>
        <span>Public Only (Set Token)</span>
      {/if}
    </button>
  </div>

  <form on:submit|preventDefault={handleSubmit} class="relative flex items-center gap-2">
    <div class="relative flex-1">
      <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
        <Search class="w-4 h-4" />
      </div>
      <input
        id="hf-url-input"
        type="text"
        bind:value={inputValue}
        placeholder="e.g. black-forest-labs/FLUX.1-dev or https://huggingface.co/.../tree/main/vae"
        class="w-full pl-10 pr-24 py-2.5 bg-dark-950/80 border border-dark-700 rounded-xl text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-accent-indigo focus:ring-1 focus:ring-accent-indigo transition-all font-mono"
        disabled={loading}
      />
      <button
        type="button"
        on:click={handlePaste}
        title="Paste from clipboard"
        class="absolute inset-y-1 right-1 px-3 flex items-center gap-1 text-xs text-slate-400 hover:text-slate-200 hover:bg-dark-800 rounded-lg transition-colors"
      >
        <Clipboard class="w-3.5 h-3.5" />
        <span>Paste</span>
      </button>
    </div>

    <button
      type="submit"
      disabled={loading || !inputValue.trim()}
      class="px-5 py-2.5 bg-gradient-to-r from-accent-indigo to-accent-purple hover:from-indigo-500 hover:to-purple-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium text-sm rounded-xl shadow-lg shadow-indigo-500/20 transition-all flex items-center gap-2 active:scale-95"
    >
      {#if loading}
        <Loader2 class="w-4 h-4 animate-spin" />
        <span>Sniffing...</span>
      {:else}
        <span>Inspect Target</span>
      {/if}
    </button>
  </form>

  <!-- Quick Presets -->
  <div class="mt-3 flex items-center gap-2 flex-wrap text-xs text-slate-400 px-1">
    <span class="text-slate-500">Quick Picks:</span>
    {#each quickPresets as preset}
      <button
        type="button"
        on:click={() => pickPreset(preset.repo)}
        class="px-2 py-0.5 rounded-md bg-dark-800/80 hover:bg-dark-700 text-slate-300 border border-dark-700/50 hover:border-slate-600 transition-colors"
      >
        {preset.label}
      </button>
    {/each}
  </div>
</div>
