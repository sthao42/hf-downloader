<script lang="ts">
  import { settingsStore, persistSettings, openTokenPage } from '../stores/settings'
  import { browseDirectory } from '../stores/bookmarks'
  import type { Settings, RoutingRule } from '../types'
  import { Settings as SettingsIcon, Save, KeyRound, ExternalLink, HardDrive, Plus, Trash2, Cpu, FolderOpen } from 'lucide-svelte'

  let localSettings: Settings | null = null
  let saving: boolean = false
  let saveSuccess: boolean = false

  // New routing rule inputs
  let newPattern: string = ''
  let newTargetDir: string = ''
  let newRuleLabel: string = ''

  $: if ($settingsStore && !localSettings) {
    localSettings = JSON.parse(JSON.stringify($settingsStore))
  }

  async function handleBrowseDefaultDir() {
    if (!localSettings) return
    const chosen = await browseDirectory(localSettings.defaultDownloadDir || '')
    if (chosen) {
      localSettings.defaultDownloadDir = chosen
      await handleSave()
    }
  }

  async function handleBrowseRuleDir() {
    const chosen = await browseDirectory(newTargetDir || '')
    if (chosen) {
      newTargetDir = chosen
    }
  }

  function addRoutingRule() {
    if (!localSettings || !newPattern.trim() || !newTargetDir.trim()) return
    const rule: RoutingRule = {
      pattern: newPattern.trim(),
      targetDir: newTargetDir.trim(),
      label: newRuleLabel.trim() || newPattern.trim(),
    }
    localSettings.routingRules = [...(localSettings.routingRules || []), rule]
    newPattern = ''
    newTargetDir = ''
    newRuleLabel = ''
  }

  function removeRoutingRule(index: number) {
    if (!localSettings) return
    localSettings.routingRules = localSettings.routingRules.filter((_, i) => i !== index)
  }

  async function handleSave() {
    if (!localSettings) return
    saving = true
    saveSuccess = false
    try {
      await persistSettings(localSettings)
      saveSuccess = true
      setTimeout(() => (saveSuccess = false), 3000)
    } catch (e) {
      console.error('Failed to save settings:', e)
    } finally {
      saving = false
    }
  }
</script>

{#if localSettings}
  <div class="bg-dark-850 border border-dark-700/60 rounded-2xl p-6 shadow-xl flex flex-col gap-6 max-w-4xl mx-auto">
    <div class="flex items-center justify-between pb-4 border-b border-dark-700/60">
      <div class="flex items-center gap-2">
        <SettingsIcon class="w-5 h-5 text-accent-indigo" />
        <h3 class="text-base font-bold text-white tracking-tight">Application Settings</h3>
      </div>

      <button
        type="button"
        on:click={handleSave}
        disabled={saving}
        class="px-5 py-2 rounded-xl bg-accent-indigo hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/25 flex items-center gap-1.5 transition-colors disabled:opacity-50"
      >
        <Save class="w-4 h-4" />
        <span>{saving ? 'Saving...' : saveSuccess ? 'Saved!' : 'Save Preferences'}</span>
      </button>
    </div>

    <!-- Default Download Location Section -->
    <div class="bg-dark-900 border border-dark-700/70 rounded-xl p-5 flex flex-col gap-3.5 shadow-sm">
      <div>
        <label for="default-download-location-input" class="text-sm font-bold text-slate-100 dark:text-white flex items-center gap-2">
          <FolderOpen class="w-4 h-4 text-accent-indigo" />
          <span>Default Download Location</span>
        </label>
        <p class="text-xs text-slate-400 dark:text-slate-400 mt-1">
          The default download directory path used whenever you paste in a Hugging Face link. Files will download here unless matched by a path routing rule.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <input
          id="default-download-location-input"
          type="text"
          bind:value={localSettings.defaultDownloadDir}
          placeholder="C:\Users\*username*\Downloads"
          class="flex-1 px-3.5 py-2.5 bg-dark-950 border border-dark-700 rounded-xl text-xs text-slate-100 font-mono placeholder:text-slate-500 focus:outline-none focus:border-accent-indigo focus:ring-1 focus:ring-accent-indigo transition-all shadow-inner"
        />
        <button
          type="button"
          on:click={handleBrowseDefaultDir}
          class="px-3.5 py-2.5 bg-dark-800 hover:bg-dark-700 active:bg-dark-750 text-accent-indigo hover:text-accent-cyan rounded-xl border border-dark-700/80 hover:border-slate-500 transition-all flex items-center justify-center flex-shrink-0 shadow-sm group"
          title="Open Windows Explorer to locate and select a new default download path"
          aria-label="Open Windows Explorer to select default download path"
        >
          <FolderOpen class="w-4 h-4 group-hover:scale-110 transition-transform" />
        </button>
      </div>

      {#if localSettings.defaultDownloadDir}
        <div class="text-[11px] text-slate-500 dark:text-slate-400 flex items-center gap-1.5 font-mono truncate">
          <span class="text-slate-600 dark:text-slate-500">•</span>
          <span>Current Default Path:</span>
          <span class="text-slate-300 dark:text-slate-200 truncate">{localSettings.defaultDownloadDir}</span>
        </div>
      {/if}
    </div>

    <!-- Hugging Face Access Token Section -->
    <div class="bg-dark-900 border border-dark-700/70 rounded-xl p-5 flex flex-col gap-3 shadow-sm">
      <div class="flex items-center justify-between">
        <label for="hf-token-input" class="text-sm font-bold text-slate-100 dark:text-white flex items-center gap-2">
          <KeyRound class="w-4 h-4 text-accent-cyan" />
          <span>Hugging Face Access Token</span>
        </label>
        <button
          type="button"
          on:click={openTokenPage}
          class="text-xs text-accent-indigo hover:text-indigo-400 hover:underline flex items-center gap-1 font-medium transition-colors"
        >
          <span>Get Token</span>
          <ExternalLink class="w-3 h-3" />
        </button>
      </div>
      <input
        id="hf-token-input"
        type="password"
        bind:value={localSettings.hfToken}
        placeholder="hf_..."
        class="w-full px-3.5 py-2.5 bg-dark-950 border border-dark-700 rounded-xl text-xs text-slate-100 font-mono focus:outline-none focus:border-accent-indigo focus:ring-1 focus:ring-accent-indigo transition-all shadow-inner"
      />
      <p class="text-xs text-slate-400 dark:text-slate-400">
        Enables downloading gated models like FLUX.1-dev, SD3, and Llama 3 without authentication prompts.
      </p>
    </div>

    <!-- Concurrency Limits -->
    <div class="bg-dark-900 border border-dark-700/70 rounded-xl p-4 flex flex-col gap-4">
      <div class="flex items-center gap-2 text-xs font-bold text-slate-200">
        <Cpu class="w-4 h-4 text-accent-purple" />
        <span>Performance & Network Concurrency</span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div>
          <label for="max-concurrent-files" class="text-[11px] text-slate-400 block mb-1">Max Concurrent Files ({localSettings.maxConcurrentFiles})</label>
          <input
            id="max-concurrent-files"
            type="range"
            min="1"
            max="6"
            step="1"
            bind:value={localSettings.maxConcurrentFiles}
            class="w-full accent-indigo-500"
          />
        </div>

        <div>
          <label for="max-connections-per-file" class="text-[11px] text-slate-400 block mb-1">Max Sockets / Connections Per File ({localSettings.maxConnectionsPerFile})</label>
          <input
            id="max-connections-per-file"
            type="range"
            min="2"
            max="16"
            step="2"
            bind:value={localSettings.maxConnectionsPerFile}
            class="w-full accent-indigo-500"
          />
        </div>
      </div>
    </div>

    <!-- Automatic Pattern Routing Rules -->
    <div class="bg-dark-900 border border-dark-700/70 rounded-xl p-4 flex flex-col gap-4">
      <div class="flex items-center justify-between">
        <div>
          <h4 class="text-xs font-bold text-slate-200">Decoupled Path Routing Rules</h4>
          <p class="text-[11px] text-slate-500 mt-0.5">
            Files matching these patterns are automatically assigned to specific destination folders.
          </p>
        </div>
      </div>

      <!-- Add New Rule Form -->
      <div class="p-3 bg-dark-950/80 border border-dark-700/80 rounded-xl grid grid-cols-1 sm:grid-cols-3 gap-2">
        <input
          type="text"
          bind:value={newPattern}
          placeholder="Pattern (e.g. *vae*.safetensors)"
          class="px-2.5 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 font-mono focus:outline-none focus:border-accent-indigo"
        />
        <div class="flex items-center gap-1.5">
          <input
            type="text"
            bind:value={newTargetDir}
            placeholder="Target Directory"
            class="w-full px-2.5 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 font-mono focus:outline-none focus:border-accent-indigo"
          />
          <button
            type="button"
            on:click={handleBrowseRuleDir}
            class="px-2 py-1.5 bg-dark-800 hover:bg-dark-700 text-slate-300 rounded-lg border border-dark-700 text-xs flex-shrink-0"
          >
            ...
          </button>
        </div>
        <div class="flex items-center gap-2">
          <input
            type="text"
            bind:value={newRuleLabel}
            placeholder="Label (optional)"
            class="w-full px-2.5 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-accent-indigo"
          />
          <button
            type="button"
            on:click={addRoutingRule}
            disabled={!newPattern.trim() || !newTargetDir.trim()}
            class="px-3 py-1.5 bg-accent-indigo hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold flex items-center gap-1 flex-shrink-0 disabled:opacity-50"
          >
            <Plus class="w-3.5 h-3.5" />
            <span>Add</span>
          </button>
        </div>
      </div>

      <!-- Existing Rules Table -->
      <div class="divide-y divide-dark-800 border border-dark-700/60 rounded-xl overflow-hidden">
        {#each localSettings.routingRules || [] as rule, index}
          <div class="p-2.5 px-3 flex items-center justify-between gap-3 text-xs">
            <div class="flex items-center gap-2 min-w-0 flex-1">
              <span class="px-2 py-0.5 rounded bg-indigo-950/40 text-accent-indigo border border-indigo-500/30 font-mono font-semibold">
                {rule.pattern}
              </span>
              <span class="text-slate-400 font-medium">➔</span>
              <span class="text-slate-300 font-mono truncate" title={rule.targetDir}>{rule.targetDir}</span>
              {#if rule.label}
                <span class="text-[10px] text-slate-500">({rule.label})</span>
              {/if}
            </div>

            <button
              type="button"
              on:click={() => removeRoutingRule(index)}
              class="p-1 text-slate-500 hover:text-rose-400 rounded transition-colors"
              title="Remove rule"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}
