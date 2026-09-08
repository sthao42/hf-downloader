<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { ShieldCheck, AlertTriangle, AlertCircle, X, Copy, Check, HardDrive, FileCheck, FolderOpen, Hash } from 'lucide-svelte'
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
    class="fixed inset-0 z-50 flex items-center justify-center p-4 verify-modal-backdrop animate-in fade-in duration-150"
    on:click|self={() => dispatch('close')}
    on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
  >
    <!-- Centered Pop Up Card with Theme-Defined Border, Surface, & Glow -->
    <div
      class="verify-modal-card rounded-2xl w-full max-w-lg shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150 transition-all
      {isSuccess ? 'is-success' : isNotice ? 'is-notice' : 'is-error'}"
    >
      <!-- Modal Header -->
      <div
        class="verify-modal-header flex items-center justify-between p-4 px-6
        {isSuccess ? 'is-success' : isNotice ? 'is-notice' : 'is-error'}"
      >
        <div class="flex items-center gap-3">
          {#if isSuccess}
            <div class="verify-icon-badge is-success p-2 rounded-xl flex items-center justify-center flex-shrink-0">
              <ShieldCheck class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold verify-modal-title tracking-tight">File Verification Succeeded</h3>
              <p class="text-[11px] verify-modal-subtitle is-success font-medium">Integrity & Checksum Validated</p>
            </div>
          {:else if isNotice}
            <div class="verify-icon-badge is-notice p-2 rounded-xl flex items-center justify-center flex-shrink-0">
              <AlertTriangle class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold verify-modal-title tracking-tight">Verification Notice</h3>
              <p class="text-[11px] verify-modal-subtitle is-notice font-medium">Attention Required</p>
            </div>
          {:else}
            <div class="verify-icon-badge is-error p-2 rounded-xl flex items-center justify-center flex-shrink-0">
              <AlertCircle class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-bold verify-modal-title tracking-tight">Verification Error</h3>
              <p class="text-[11px] verify-modal-subtitle is-error font-medium">Inspection Failed</p>
            </div>
          {/if}
        </div>

        <button
          type="button"
          on:click={() => dispatch('close')}
          class="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/60 dark:hover:bg-dark-800 transition-colors"
          title="Close dialog"
          aria-label="Close"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Content Body -->
      <div class="p-6 flex flex-col gap-4">
        <!-- Target File Information Card -->
        {#if item}
          <div class="p-3 verify-info-box flex items-start gap-2.5">
            <FileCheck class="w-4 h-4 text-indigo-600 dark:text-accent-indigo flex-shrink-0 mt-0.5" />
            <div class="min-w-0 flex-1">
              <div class="text-xs font-bold text-slate-900 dark:text-slate-100 truncate" title={item.finalFilename}>
                {item.finalFilename}
              </div>
              <div class="text-[11px] text-slate-500 dark:text-slate-400 font-mono truncate mt-0.5" title={item.destinationDir}>
                {item.destinationDir}
              </div>
            </div>
          </div>
        {/if}

        {#if isSuccess && result}
          <!-- Success Status Banner -->
          <div class="p-3.5 verify-status-banner is-success rounded-xl text-xs flex items-center gap-2.5">
            <ShieldCheck class="w-4 h-4 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
            <span>The local file is complete, uncorrupted, and exactly matches the expected remote size and checksum digest.</span>
          </div>

          <!-- File Size Card -->
          <div class="p-3.5 verify-info-box flex items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <HardDrive class="w-4 h-4 text-cyan-700 dark:text-accent-cyan flex-shrink-0" />
              <span class="text-xs text-slate-700 dark:text-slate-300 font-semibold">Verified File Size:</span>
            </div>
            <div class="text-right">
              <span class="text-xs font-mono font-bold text-slate-900 dark:text-white">
                {formatBytes(result.actualSize)}
              </span>
              <span class="text-[10px] text-slate-500 dark:text-slate-400 font-mono block">
                ({result.actualSize.toLocaleString()} bytes)
              </span>
            </div>
          </div>

          <!-- SHA-256 Checksum Card with High-Contrast Console Box -->
          {@const hashToDisplay = result.actualSha256 || item?.expectedSha256 || ''}
          <div class="p-3.5 verify-info-box flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <span class="text-xs text-slate-800 dark:text-slate-200 font-bold flex items-center gap-1.5">
                <Hash class="w-3.5 h-3.5 text-indigo-600 dark:text-accent-indigo flex-shrink-0" />
                <span>SHA-256 Checksum:</span>
              </span>
              {#if hashToDisplay}
                <button
                  type="button"
                  on:click={() => handleCopyHash(hashToDisplay)}
                  class="verify-btn-secondary px-2.5 py-1 rounded-lg text-[11px] font-medium flex items-center gap-1.5 transition-colors"
                  title="Copy SHA-256 hash"
                >
                  {#if copied}
                    <Check class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
                    <span class="text-emerald-700 dark:text-emerald-400 font-bold">Copied!</span>
                  {:else}
                    <Copy class="w-3.5 h-3.5" />
                    <span>Copy Hash</span>
                  {/if}
                </button>
              {/if}
            </div>

            <!-- Sleek High-Contrast Monospace Code Box -->
            <div class="p-3 verify-checksum-box rounded-lg text-xs font-mono break-all select-all leading-relaxed tracking-wide">
              {hashToDisplay || 'Matches expected remote digest'}
            </div>

            <div class="flex items-center gap-1.5 text-[11px] font-semibold text-emerald-700 dark:text-emerald-400 mt-0.5">
              <Check class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
              <span>Integrity Verified &bull; Exact Bit-for-Bit Digest Match</span>
            </div>
          </div>

        {:else if isNotice && result}
          <!-- Notice / Incomplete / Mismatch Banner -->
          <div class="p-3.5 verify-status-banner is-notice rounded-xl text-xs flex items-start gap-2.5">
            <AlertTriangle class="w-4 h-4 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <div class="font-bold text-amber-950 dark:text-amber-200 mb-1">Verification Notice:</div>
              <div class="text-[11px] leading-relaxed text-amber-900 dark:text-amber-300/90">{result.message}</div>
            </div>
          </div>

          {#if result.exists}
            <div class="p-3.5 verify-info-box flex items-center justify-between text-xs">
              <span class="text-slate-700 dark:text-slate-300 font-medium">Current Local Size:</span>
              <span class="font-mono text-slate-900 dark:text-slate-100 font-bold">{formatBytes(result.actualSize)}</span>
            </div>
          {/if}

        {:else if isError}
          <!-- Error Banner -->
          <div class="p-3.5 verify-status-banner is-error rounded-xl text-xs flex items-start gap-2.5">
            <AlertCircle class="w-4 h-4 text-rose-600 dark:text-rose-400 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <div class="font-bold text-rose-950 dark:text-rose-200 mb-1">Failed to Verify File:</div>
              <div class="text-[11px] font-mono leading-relaxed text-rose-900 dark:text-rose-300/90">{error}</div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Modal Footer Actions -->
      <div class="flex items-center justify-between p-4 px-6 verify-modal-footer">
        <div>
          {#if item?.destinationDir}
            <button
              type="button"
              on:click={() => item && openTaskFolder(item.destinationDir)}
              class="verify-btn-secondary px-3.5 py-1.5 rounded-xl text-xs font-medium flex items-center gap-1.5 transition-colors"
            >
              <FolderOpen class="w-3.5 h-3.5" />
              <span>Open Folder</span>
            </button>
          {/if}
        </div>

        <button
          type="button"
          on:click={() => dispatch('close')}
          class="verify-btn-primary px-6 py-1.5 rounded-xl text-xs font-bold transition-all
          {isSuccess ? 'is-success' : 'is-action'}"
        >
          Done
        </button>
      </div>
    </div>
  </div>
{/if}
