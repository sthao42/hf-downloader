<script lang="ts">
  import { queueStore, startTask, pauseTask, resumeTask, cancelTask, removeTask, openTaskFolder, verifyTaskFile } from '../stores/queue'
  import { settingsStore, persistSettings } from '../stores/settings'
  import { formatBytes } from '../utils'
  import type { DownloadItem } from '../types'
  import {
    ListOrdered,
    Play,
    Pause,
    X,
    FolderCheck,
    AlertCircle,
    CheckCircle2,
    Clock,
    Zap,
    ExternalLink,
    ShieldCheck,
    RotateCw,
    ToggleLeft,
    ToggleRight
  } from 'lucide-svelte'

  let activeTab: 'all' | 'active' | 'staged' | 'completed' = 'all'
  let verifyingMap: Record<string, boolean> = {}

  $: queue = $queueStore

  $: filteredItems = queue.filter(item => {
    if (activeTab === 'active') return item.status === 'downloading' || item.status === 'queued' || item.status === 'verifying'
    if (activeTab === 'staged') return item.status === 'staged' || item.status === 'paused'
    if (activeTab === 'completed') return item.status === 'completed' || item.status === 'failed'
    return true
  })

  $: activeCount = queue.filter(i => i.status === 'downloading').length
  $: stagedCount = queue.filter(i => i.status === 'staged').length

  function formatETA(seconds: number): string {
    if (!seconds || seconds <= 0) return ''
    if (seconds < 60) return `${seconds}s`
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    if (mins < 60) return `${mins}m ${secs}s`
    const hrs = Math.floor(mins / 60)
    return `${hrs}h ${mins % 60}m`
  }

  async function handleToggleAutoDownload() {
    if (!$settingsStore) return
    const updated = { ...$settingsStore, autoDownload: !$settingsStore.autoDownload }
    await persistSettings(updated)
  }

  function startAllStaged() {
    for (const item of queue) {
      if (item.status === 'staged' || item.status === 'paused') {
        startTask(item.id)
      }
    }
  }

  function clearCompleted() {
    for (const item of queue) {
      if (item.status === 'completed') {
        removeTask(item.id)
      }
    }
  }

  async function handleManualVerify(item: DownloadItem) {
    verifyingMap = { ...verifyingMap, [item.id]: true }
    try {
      const res = await verifyTaskFile(item)
      if (res.exists && res.valid) {
        alert(`File verified successfully!\nSize: ${formatBytes(res.actualSize)}\nSHA-256: ${res.actualSha256 || 'Matches expected'}`)
      } else {
        alert(`Verification notice:\n${res.message}`)
      }
    } catch (e) {
      alert(`Error verifying file: ${e}`)
    } finally {
      verifyingMap = { ...verifyingMap, [item.id]: false }
    }
  }
</script>

<div class="bg-dark-850 border border-dark-700/60 rounded-2xl p-5 shadow-xl flex flex-col gap-4">
  <!-- Queue Header & Controls -->
  <div class="flex items-center justify-between flex-wrap gap-3 pb-4 border-b border-dark-700/50">
    <div class="flex items-center gap-3">
      <div class="flex items-center gap-2">
        <ListOrdered class="w-5 h-5 text-accent-indigo" />
        <h3 class="text-base font-bold text-white tracking-tight">Download Queue</h3>
      </div>

      <!-- Auto-Download Toggle -->
      <button
        type="button"
        on:click={handleToggleAutoDownload}
        class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border transition-colors {$settingsStore?.autoDownload ? 'bg-emerald-500/15 text-emerald-400 border-emerald-500/40 hover:bg-emerald-500/25' : 'bg-dark-800 text-slate-400 border-dark-700 hover:text-slate-200'}"
      >
        {#if $settingsStore?.autoDownload}
          <ToggleRight class="w-4 h-4 text-emerald-400" />
          <span>Auto-Download: ON</span>
        {:else}
          <ToggleLeft class="w-4 h-4 text-slate-500" />
          <span>Auto-Download: OFF</span>
        {/if}
      </button>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2">
      {#if stagedCount > 0}
        <button
          type="button"
          on:click={startAllStaged}
          class="px-3 py-1.5 rounded-xl bg-accent-indigo hover:bg-indigo-500 text-white text-xs font-semibold flex items-center gap-1.5 transition-colors shadow-sm"
        >
          <Play class="w-3 h-3 fill-current" />
          <span>Start All Staged ({stagedCount})</span>
        </button>
      {/if}
      <button
        type="button"
        on:click={clearCompleted}
        class="px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 text-slate-300 border border-dark-700 text-xs font-medium transition-colors"
      >
        Clear Completed
      </button>
    </div>
  </div>

  <!-- Queue Filter Tabs -->
  <div class="flex items-center gap-2">
    <button
      type="button"
      on:click={() => (activeTab = 'all')}
      class="px-3 py-1 rounded-lg text-xs font-medium transition-colors {activeTab === 'all' ? 'bg-dark-700 text-white font-semibold' : 'text-slate-400 hover:text-slate-200'}"
    >
      All ({queue.length})
    </button>
    <button
      type="button"
      on:click={() => (activeTab = 'active')}
      class="px-3 py-1 rounded-lg text-xs font-medium transition-colors {activeTab === 'active' ? 'bg-dark-700 text-white font-semibold' : 'text-slate-400 hover:text-slate-200'}"
    >
      Active ({activeCount})
    </button>
    <button
      type="button"
      on:click={() => (activeTab = 'staged')}
      class="px-3 py-1 rounded-lg text-xs font-medium transition-colors {activeTab === 'staged' ? 'bg-dark-700 text-white font-semibold' : 'text-slate-400 hover:text-slate-200'}"
    >
      Staged / Paused ({stagedCount})
    </button>
    <button
      type="button"
      on:click={() => (activeTab = 'completed')}
      class="px-3 py-1 rounded-lg text-xs font-medium transition-colors {activeTab === 'completed' ? 'bg-dark-700 text-white font-semibold' : 'text-slate-400 hover:text-slate-200'}"
    >
      Completed / Failed
    </button>
  </div>

  <!-- Items List -->
  {#if filteredItems.length === 0}
    <div class="py-12 text-center text-slate-500 text-sm flex flex-col items-center gap-2">
      <ListOrdered class="w-8 h-8 text-slate-600" />
      <span>No download items in this view. Paste a Hugging Face URL above to stage models!</span>
    </div>
  {:else}
    <div class="flex flex-col gap-3">
      {#each filteredItems as item (item.id)}
        <div class="bg-dark-900 border border-dark-700/80 rounded-xl p-4 flex flex-col gap-3 shadow-md hover:border-slate-600/70 transition-colors">
          <!-- Top Row: File Name, Status Pill & Controls -->
          <div class="flex items-start justify-between gap-3 flex-wrap">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-xs font-mono font-bold text-white truncate">{item.finalFilename}</span>
                <span class="text-xs font-mono text-slate-500 truncate">({item.repoId})</span>

                <!-- Status Badge -->
                {#if item.status === 'downloading'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-accent-indigo/20 text-accent-indigo border border-accent-indigo/40 flex items-center gap-1 animate-pulse">
                    <Zap class="w-3 h-3" />
                    <span>Downloading</span>
                  </span>
                {:else if item.status === 'verifying'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/20 text-amber-400 border border-amber-500/40 flex items-center gap-1">
                    <ShieldCheck class="w-3 h-3" />
                    <span>Verifying SHA-256</span>
                  </span>
                {:else if item.status === 'queued'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-blue-500/20 text-blue-400 border border-blue-500/40 flex items-center gap-1">
                    <Clock class="w-3 h-3" />
                    <span>Queued</span>
                  </span>
                {:else if item.status === 'staged'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-700/50 text-slate-300 border border-slate-600">
                    Staged
                  </span>
                {:else if item.status === 'paused'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/20 text-amber-400 border border-amber-500/40">
                    Paused
                  </span>
                {:else if item.status === 'completed'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/40 flex items-center gap-1">
                    <CheckCircle2 class="w-3 h-3" />
                    <span>Completed & Verified</span>
                  </span>
                {:else if item.status === 'failed'}
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-rose-500/20 text-rose-400 border border-rose-500/40 flex items-center gap-1">
                    <AlertCircle class="w-3 h-3" />
                    <span>Failed</span>
                  </span>
                {/if}
              </div>

              <!-- Destination Directory -->
              <div class="text-[11px] font-mono text-slate-500 mt-1 truncate" title={item.destinationDir}>
                Target: <span class="text-slate-400">{item.destinationDir}</span>
              </div>
            </div>

            <!-- Card Actions -->
            <div class="flex items-center gap-1.5 flex-shrink-0">
              {#if item.status === 'staged' || item.status === 'paused'}
                <button
                  type="button"
                  on:click={() => startTask(item.id)}
                  class="px-2.5 py-1 bg-accent-indigo hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors"
                >
                  <Play class="w-3 h-3 fill-current" />
                  <span>Start</span>
                </button>
              {:else if item.status === 'downloading'}
                <button
                  type="button"
                  on:click={() => pauseTask(item.id)}
                  class="px-2.5 py-1 bg-dark-800 hover:bg-dark-700 text-amber-400 rounded-lg border border-dark-700 text-xs font-medium flex items-center gap-1 transition-colors"
                >
                  <Pause class="w-3 h-3" />
                  <span>Pause</span>
                </button>
                <button
                  type="button"
                  on:click={() => cancelTask(item.id)}
                  class="p-1 hover:text-rose-400 text-slate-400 rounded transition-colors"
                  title="Cancel download"
                >
                  <X class="w-4 h-4" />
                </button>
              {:else if item.status === 'failed'}
                <button
                  type="button"
                  on:click={() => startTask(item.id)}
                  class="px-2 py-1 bg-dark-800 hover:bg-dark-700 text-accent-indigo rounded-lg border border-dark-700 text-xs font-medium flex items-center gap-1 transition-colors"
                >
                  <RotateCw class="w-3 h-3" />
                  <span>Retry</span>
                </button>
              {/if}

              <!-- Verify Button -->
              <button
                type="button"
                on:click={() => handleManualVerify(item)}
                disabled={verifyingMap[item.id]}
                class="px-2 py-1 bg-dark-800 hover:bg-dark-700 text-slate-300 rounded-lg border border-dark-700 text-xs font-medium flex items-center gap-1 transition-colors disabled:opacity-50"
                title="Verify local file against SHA-256"
              >
                <ShieldCheck class="w-3 h-3 text-accent-cyan" />
                <span>Verify</span>
              </button>

              <!-- Open Destination Folder -->
              <button
                type="button"
                on:click={() => openTaskFolder(item.destinationDir)}
                class="px-2 py-1 bg-dark-800 hover:bg-dark-700 text-slate-300 rounded-lg border border-dark-700 text-xs font-medium flex items-center gap-1 transition-colors"
                title="Open containing folder"
              >
                <ExternalLink class="w-3 h-3" />
                <span>Open</span>
              </button>

              <!-- Remove from list -->
              <button
                type="button"
                on:click={() => removeTask(item.id)}
                class="p-1 hover:text-rose-400 text-slate-500 rounded transition-colors"
                title="Remove from queue"
              >
                <X class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Progress Bar -->
          <div class="w-full bg-dark-950 rounded-full h-2 overflow-hidden border border-dark-700/50">
            <div
              class="h-full rounded-full transition-all duration-300 {item.status === 'completed' ? 'bg-emerald-500' : item.status === 'failed' ? 'bg-rose-500' : 'bg-gradient-to-r from-accent-indigo to-accent-cyan'}"
              style="width: {Math.min(100, Math.max(0, item.progress))}%"
            ></div>
          </div>

          <!-- Bottom Row: Speed, ETA, Bytes -->
          <div class="flex items-center justify-between text-xs font-mono text-slate-400 flex-wrap gap-2">
            <div class="flex items-center gap-3">
              <span>{item.progress ? item.progress.toFixed(1) : '0.0'}%</span>
              <span>
                {formatBytes(item.downloadedBytes || 0)} / {formatBytes(item.size || 0)}
              </span>
            </div>

            <div class="flex items-center gap-3">
              {#if item.status === 'downloading'}
                {#if item.speedFormatted}
                  <span class="text-accent-cyan font-semibold">{item.speedFormatted}</span>
                {/if}
                {#if item.etaSeconds}
                  <span class="text-slate-400">ETA: {formatETA(item.etaSeconds)}</span>
                {/if}
              {:else if item.status === 'completed'}
                <span class="text-emerald-400 font-medium">Download Finished</span>
              {:else if item.status === 'failed' && item.errorMessage}
                <span class="text-rose-400 font-medium truncate max-w-sm">{item.errorMessage}</span>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
