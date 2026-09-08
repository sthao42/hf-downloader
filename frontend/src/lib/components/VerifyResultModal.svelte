<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { ShieldCheck, AlertTriangle, AlertCircle, X, Copy, Check, HardDrive, FileCheck, FolderOpen } from 'lucide-svelte'
  import { formatBytes } from '../utils'
  import { openTaskFolder } from '../stores/queue'
  import type { DownloadItem, VerificationResult } from '../types'

  export let show: boolean = false
  export let item: DownloadItem | null = null
  export let result: VerificationResult | null = null
  export let error: string | null = null

  const dispatch = createEventDispatcher<{
    close: void
  }>()

  let copied: boolean = false
  let copyTimeout: any = null

  $: isSuccess = !error && result && result.exists && result.valid
  $: isNotice = !error && result && (!result.exists || !result.valid)
  $: isError = Boolean(error)

  async function handleCopyHash(hashText: string) {
    if (!hashText) return
    try {
      await navigator.clipboard.writeText(hashText)
      copied = true
      if (copyTimeout) clearTimeout(copyTimeout)
      copyTimeout = setTimeout(() => {
        copied = false
      }, 2000)
    } catch (e) {
      console.error('Failed to copy hash:', e)
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      dispatch('close')
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
  <!-- Centered Modal Backdrop -->
  <div
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150"
    on:click|self={() => dispatch('close')}
    on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
  >
    <!-- Centered Pop Up Card with Theme Matching Border & Rounded Corners -->
    <div
      class="bg-dark-900 rounded-2xl w-full max-w-lg shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150 transition-all border-2
      {isSuccess ? 'border-emerald-500/50 shadow-emerald-950/40' : isNotice ? 'border-amber-500/50 shadow-amber-950/40' : 'border-rose-500/50 shadow-rose-950/40'}"
    >
      <!-- Modal Header -->
      <div
        class="flex items-center justify-between p-4 px-6 border-b
        {isSuccess ? 'bg-gradient-to-r from-emerald-950/40 via-dark-850 to-dark-850 border-emerald-500/20' : isNotice ? 'bg-gradient-to-r from-amber-950/40 via-dark-850 to-dark-850 border-amber-500/20' : 'bg-gradient-to-r from-rose-950/40 via-dark-850 to-dark-850 border-rose-500/20'}"
      >
        <div class="flex items-center gap-3">
          {#if isSuccess}
            <div class="p-2 rounded-xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
              <ShieldCheck class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-white tracking-tight">File Verification Succeeded</h3>
              <p class="text-[11px] text-emerald-400 font-medium">Integrity & Checksum Validated</p>
            </div>
          {:else if isNotice}
            <div class="p-2 rounded-xl bg-amber-500/15 text-amber-400 border border-amber-500/30">
              <AlertTriangle class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-white tracking-tight">Verification Notice</h3>
              <p class="text-[11px] text-amber-400 font-medium">Attention Required</p>
            </div>
          {:else}
            <div class="p-2 rounded-xl bg-rose-500/15 text-rose-400 border border-rose-500/30">
              <AlertCircle class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-white tracking-tight">Verification Error</h3>
              <p class="text-[11px] text-rose-400 font-medium">Inspection Failed</p>
            </div>
          {/if}
        </div>

        <button
          type="button"
          on:click={() => dispatch('close')}
          class="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-800 transition-colors"
          title="Close dialog"
          aria-label="Close"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Content -->
      <div class="p-6 flex flex-col gap-4">
        <!-- Target File Information -->
        {#if item}
          <div class="p-3 bg-dark-950/70 border border-dark-700/70 rounded-xl flex items-start gap-2.5">
            <FileCheck class="w-4 h-4 text-accent-indigo flex-shrink-0 mt-0.5" />
            <div class="min-w-0 flex-1">
              <div class="text-xs font-bold text-slate-200 truncate" title={item.finalFilename}>
                {item.finalFilename}
              </div>
              <div class="text-[11px] text-slate-500 font-mono truncate mt-0.5" title={item.destinationDir}>
                {item.destinationDir}
              </div>
            </div>
          </div>
        {/if}

        {#if isSuccess && result}
          <!-- Success Status Banner -->
          <div class="p-3 bg-emerald-950/25 border border-emerald-500/30 rounded-xl text-xs text-emerald-300 flex items-center gap-2">
            <ShieldCheck class="w-4 h-4 text-emerald-400 flex-shrink-0" />
            <span>The local file is complete and exactly matches the expected remote size and checksum digest.</span>
          </div>

          <!-- File Size Card -->
          <div class="p-3.5 bg-dark-950/60 border border-dark-700/70 rounded-xl flex items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <HardDrive class="w-4 h-4 text-accent-cyan" />
              <span class="text-xs text-slate-400 font-medium">Verified File Size:</span>
            </div>
            <div class="text-right">
              <span class="text-xs font-mono font-bold text-white">
                {formatBytes(result.actualSize)}
              </span>
              <span class="text-[10px] text-slate-500 font-mono block">
                ({result.actualSize.toLocaleString()} bytes)
              </span>
            </div>
          </div>

          <!-- SHA-256 Checksum Card -->
          {@const hashToDisplay = result.actualSha256 || item?.expectedSha256 || ''}
          <div class="p-3.5 bg-dark-950/60 border border-dark-700/70 rounded-xl flex flex-col gap-2">
            <div class="flex items-center justify-between">
              <span class="text-xs text-slate-400 font-medium">SHA-256 Checksum:</span>
              {#if hashToDisplay}
                <button
                  type="button"
                  on:click={() => handleCopyHash(hashToDisplay)}
                  class="px-2 py-0.5 rounded bg-dark-800 hover:bg-dark-700 text-[11px] text-slate-300 hover:text-white border border-dark-700 flex items-center gap-1 transition-colors"
                  title="Copy SHA-256 hash"
                >
                  {#if copied}
                    <Check class="w-3 h-3 text-emerald-400" />
                    <span class="text-emerald-400 font-semibold">Copied!</span>
                  {:else}
                    <Copy class="w-3 h-3" />
                    <span>Copy Hash</span>
                  {/if}
                </button>
              {/if}
            </div>

            <div class="p-2.5 bg-dark-900 border border-dark-700 rounded-lg text-xs font-mono text-emerald-400 break-all select-all leading-relaxed">
              {hashToDisplay || 'Matches expected remote digest'}
            </div>
          </div>

        {:else if isNotice && result}
          <!-- Notice / Incomplete / Mismatch Banner -->
          <div class="p-3.5 bg-amber-950/25 border border-amber-500/30 rounded-xl text-xs text-amber-300 flex items-start gap-2.5">
            <AlertTriangle class="w-4 h-4 text-amber-400 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <div class="font-semibold text-amber-200 mb-1">Verification Notice:</div>
              <div class="text-[11px] leading-relaxed text-amber-300/90">{result.message}</div>
            </div>
          </div>

          {#if result.exists}
            <div class="p-3 bg-dark-950/60 border border-dark-700/70 rounded-xl flex items-center justify-between text-xs">
              <span class="text-slate-400">Current Local Size:</span>
              <span class="font-mono text-slate-200 font-bold">{formatBytes(result.actualSize)}</span>
            </div>
          {/if}

        {:else if isError}
          <!-- Error Banner -->
          <div class="p-3.5 bg-rose-950/25 border border-rose-500/30 rounded-xl text-xs text-rose-300 flex items-start gap-2.5">
            <AlertCircle class="w-4 h-4 text-rose-400 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <div class="font-semibold text-rose-200 mb-1">Failed to Verify File:</div>
              <div class="text-[11px] font-mono leading-relaxed text-rose-300/90">{error}</div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Modal Footer Actions -->
      <div class="flex items-center justify-between p-4 px-6 border-t border-dark-700/60 bg-dark-850/80">
        <div>
          {#if item?.destinationDir}
            <button
              type="button"
              on:click={() => item && openTaskFolder(item.destinationDir)}
              class="px-3 py-1.5 rounded-xl bg-dark-800 hover:bg-dark-700 text-slate-300 border border-dark-700 text-xs font-medium flex items-center gap-1.5 transition-colors"
            >
              <FolderOpen class="w-3.5 h-3.5 text-slate-400" />
              <span>Open Folder</span>
            </button>
          {/if}
        </div>

        <button
          type="button"
          on:click={() => dispatch('close')}
          class="px-5 py-1.5 rounded-xl text-xs font-semibold transition-colors shadow-sm
          {isSuccess ? 'bg-emerald-600 hover:bg-emerald-500 text-white' : 'bg-accent-indigo hover:bg-indigo-500 text-white'}"
        >
          Done
        </button>
      </div>
    </div>
  </div>
{/if}
