<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { settingsStore, persistSettings, openTokenPage } from '../stores/settings'
  import { X, KeyRound, ExternalLink, ShieldAlert, Check, Eye, EyeOff } from 'lucide-svelte'

  export let show: boolean = false

  const dispatch = createEventDispatcher<{
    close: void
  }>()

  let tokenInput: string = ''
  let showSecret: boolean = false
  let saving: boolean = false

  $: if (show && $settingsStore) {
    tokenInput = $settingsStore.hfToken || ''
  }

  async function handleSave() {
    if (!$settingsStore) return
    saving = true
    try {
      const updated = { ...$settingsStore, hfToken: tokenInput.trim() }
      await persistSettings(updated)
      dispatch('close')
    } catch (e) {
      console.error(e)
    } finally {
      saving = false
    }
  }

  function handleClear() {
    tokenInput = ''
  }
</script>

{#if show}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
    <div class="bg-dark-900 border border-dark-700 rounded-2xl w-full max-w-lg shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 px-6 border-b border-dark-700/70 bg-dark-850">
        <div class="flex items-center gap-2">
          <KeyRound class="w-5 h-5 text-accent-indigo" />
          <h3 class="text-base font-bold text-white">Hugging Face Access Token</h3>
        </div>
        <button
          type="button"
          on:click={() => dispatch('close')}
          class="p-1 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-800 transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Content Body -->
      <div class="p-6 flex flex-col gap-4">
        <div class="p-3.5 bg-indigo-950/20 border border-indigo-500/30 rounded-xl text-xs text-indigo-300 flex items-start gap-2.5">
          <ShieldAlert class="w-4 h-4 text-accent-indigo flex-shrink-0 mt-0.5" />
          <div class="leading-relaxed">
            A Hugging Face token (read access) is required for gated or private models such as <span class="font-bold text-white">FLUX.1-dev</span>, <span class="font-bold text-white">Llama 3</span>, or <span class="font-bold text-white">Stable Diffusion 3</span>.
          </div>
        </div>

        <div>
          <label for="hf-token" class="text-xs font-semibold text-slate-300 block mb-1.5">
            Personal Access Token (starts with <code class="text-accent-cyan">hf_...</code>)
          </label>
          <div class="relative">
            <input
              id="hf-token"
              type={showSecret ? 'text' : 'password'}
              bind:value={tokenInput}
              placeholder="hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
              class="w-full pl-3.5 pr-10 py-2 bg-dark-950 border border-dark-700 rounded-xl text-xs text-slate-100 font-mono focus:outline-none focus:border-accent-indigo"
            />
            <button
              type="button"
              on:click={() => (showSecret = !showSecret)}
              class="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-200"
            >
              {#if showSecret}
                <EyeOff class="w-4 h-4" />
              {:else}
                <Eye class="w-4 h-4" />
              {/if}
            </button>
          </div>
        </div>

        <div class="flex items-center justify-between pt-1">
          <button
            type="button"
            on:click={openTokenPage}
            class="text-xs text-accent-indigo hover:text-indigo-400 flex items-center gap-1.5 transition-colors font-medium"
          >
            <span>Create Read Token on Hugging Face</span>
            <ExternalLink class="w-3.5 h-3.5" />
          </button>

          {#if tokenInput}
            <button
              type="button"
              on:click={handleClear}
              class="text-xs text-rose-400 hover:text-rose-300 transition-colors"
            >
              Clear Token
            </button>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 px-6 bg-dark-850 border-t border-dark-700/70 flex items-center justify-end gap-3">
        <button
          type="button"
          on:click={() => dispatch('close')}
          class="px-4 py-2 rounded-xl bg-dark-800 hover:bg-dark-700 text-slate-300 text-xs font-medium border border-dark-700 transition-colors"
        >
          Cancel
        </button>
        <button
          type="button"
          on:click={handleSave}
          disabled={saving}
          class="px-5 py-2 rounded-xl bg-accent-indigo hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/25 flex items-center gap-1.5 transition-colors disabled:opacity-50"
        >
          <Check class="w-4 h-4" />
          <span>Save Token</span>
        </button>
      </div>
    </div>
  </div>
{/if}
