<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { ShieldCheck, AlertTriangle, AlertCircle, X, Copy, Check, HardDrive, FileCheck, FolderOpen, Globe } from 'lucide-svelte'
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

  let copiedExpected: boolean = false
  let copiedActual: boolean = false
  let copyExpectedTimeout: any = null
  let copyActualTimeout: any = null

  $: expectedHash = (item?.expectedSha256 || result?.expectedSha256 || '').trim()
  $: actualHash = (result?.actualSha256 || '').trim()

  $: hasExpected = Boolean(expectedHash)
  $: hasActual = Boolean(actualHash)
  $: hashesMatch = Boolean(hasExpected && hasActual && expectedHash.toLowerCase() === actualHash.toLowerCase())
  $: hashesDiffer = Boolean(hasExpected && hasActual && expectedHash.toLowerCase() !== actualHash.toLowerCase())

  $: isSuccess = !error && result && result.exists && result.valid && !hashesDiffer
  $: isNotice = !error && result && (!result.exists || !result.valid || hashesDiffer || (!hasExpected && hasActual))
  $: isError = Boolean(error)

  async function handleCopyHash(hashText: string, target: 'expected' | 'actual') {
    if (!hashText) return
    try {
      await navigator.clipboard.writeText(hashText)
      if (target === 'expected') {
        copiedExpected = true
        if (copyExpectedTimeout) clearTimeout(copyExpectedTimeout)
        copyExpectedTimeout = setTimeout(() => {
          copiedExpected = false
        }, 2000)
      } else {
        copiedActual = true
        if (copyActualTimeout) clearTimeout(copyActualTimeout)
        copyActualTimeout = setTimeout(() => {
          copiedActual = false
        }, 2000)
      }
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
      <div class="p-6 flex flex-col gap-4 max-h-[80vh] overflow-y-auto">
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

        <!-- Status Banner -->
        {#if isSuccess && result}
          <div class="p-3.5 verify-status-banner is-success rounded-xl text-xs flex items-center gap-2.5">
            <ShieldCheck class="w-4 h-4 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
            <span>The local file is complete, uncorrupted, and verified against remote repository metadata.</span>
          </div>
        {:else if isNotice && result}
          <div class="p-3.5 verify-status-banner is-notice rounded-xl text-xs flex items-start gap-2.5">
            <AlertTriangle class="w-4 h-4 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <div class="font-bold text-amber-950 dark:text-amber-200 mb-1">Verification Notice:</div>
              <div class="text-[11px] leading-relaxed text-amber-900 dark:text-amber-300/90">{result.message}</div>
            </div>
          </div>
        {:else if isError}
          <div class="p-3.5 verify-status-banner is-error rounded-xl text-xs flex items-start gap-2.5">
            <AlertCircle class="w-4 h-4 text-rose-600 dark:text-rose-400 flex-shrink-0 mt-0.5" />
            <div class="flex-1">
              <div class="font-bold text-rose-950 dark:text-rose-200 mb-1">Failed to Verify File:</div>
              <div class="text-[11px] font-mono leading-relaxed text-rose-900 dark:text-rose-300/90">{error}</div>
            </div>
          </div>
        {/if}

        <!-- File Size Card -->
        {#if result && result.exists}
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
        {/if}

        <!-- Checksum Cards (Shown whenever result exists) -->
        {#if result}
          <!-- 1. Expected Remote SHA-256 Card -->
          <div class="p-3.5 verify-info-box flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2 min-w-0">
                <Globe class="w-3.5 h-3.5 text-indigo-600 dark:text-accent-indigo flex-shrink-0" />
                <span class="text-xs text-slate-800 dark:text-slate-200 font-bold">Expected Remote SHA-256:</span>
                <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 font-medium">Remote Metadata</span>
              </div>
              {#if hasExpected}
                <button
                  type="button"
                  on:click={() => handleCopyHash(expectedHash, 'expected')}
                  class="verify-btn-secondary px-2.5 py-1 rounded-lg text-[11px] font-medium flex items-center gap-1.5 transition-colors flex-shrink-0"
                  title="Copy Expected Remote SHA-256"
                >
                  {#if copiedExpected}
                    <Check class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
                    <span class="text-emerald-700 dark:text-emerald-400 font-bold">Copied!</span>
                  {:else}
                    <Copy class="w-3.5 h-3.5" />
                    <span>Copy</span>
                  {/if}
                </button>
              {/if}
            </div>

            <!-- Monospace Code Box -->
            <div class="p-3 verify-checksum-box rounded-lg text-xs font-mono break-all select-all leading-relaxed tracking-wide">
              {#if hasExpected}
                {expectedHash}
              {:else}
                <span class="italic text-slate-400 dark:text-slate-500 font-sans">Not provided in repository metadata</span>
              {/if}
            </div>
          </div>

          <!-- 2. Scanned Local SHA-256 Card -->
          <div class="p-3.5 verify-info-box flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2 min-w-0">
                <HardDrive class="w-3.5 h-3.5 text-cyan-600 dark:text-accent-cyan flex-shrink-0" />
                <span class="text-xs text-slate-800 dark:text-slate-200 font-bold">Scanned Local SHA-256:</span>
                <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 font-medium">Computed on Disk</span>
              </div>
              {#if hasActual}
                <button
                  type="button"
                  on:click={() => handleCopyHash(actualHash, 'actual')}
                  class="verify-btn-secondary px-2.5 py-1 rounded-lg text-[11px] font-medium flex items-center gap-1.5 transition-colors flex-shrink-0"
                  title="Copy Scanned Local SHA-256"
                >
                  {#if copiedActual}
                    <Check class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
                    <span class="text-emerald-700 dark:text-emerald-400 font-bold">Copied!</span>
                  {:else}
                    <Copy class="w-3.5 h-3.5" />
                    <span>Copy</span>
                  {/if}
                </button>
              {/if}
            </div>

            <!-- Monospace Code Box -->
            <div class="p-3 verify-checksum-box rounded-lg text-xs font-mono break-all select-all leading-relaxed tracking-wide">
              {#if hasActual}
                {actualHash}
              {:else}
                <span class="italic text-slate-400 dark:text-slate-500 font-sans">Local file hash not available</span>
              {/if}
            </div>
          </div>

          <!-- 3. Comparison Result Match Indicator -->
          {#if hasExpected && hasActual}
            {#if hashesMatch}
              <div class="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-800 dark:text-emerald-300 text-xs flex items-center gap-2.5">
                <Check class="w-4 h-4 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
                <div>
                  <span class="font-bold">Exact Bit-for-Bit Digest Match:</span>
                  <span class="block text-[11px] opacity-90 mt-0.5">The scanned local SHA-256 matches the expected remote repository hash.</span>
                </div>
              </div>
            {:else}
              <div class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-800 dark:text-rose-300 text-xs flex items-center gap-2.5">
                <AlertTriangle class="w-4 h-4 text-rose-600 dark:text-rose-400 flex-shrink-0" />
                <div>
                  <span class="font-bold">Checksum Mismatch:</span>
                  <span class="block text-[11px] opacity-90 mt-0.5">The scanned local file SHA-256 does not match the expected remote repository checksum.</span>
                </div>
              </div>
            {/if}
          {:else if hasActual && !hasExpected}
            <div class="p-3 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-900 dark:text-amber-300 text-xs flex items-center gap-2.5">
              <AlertCircle class="w-4 h-4 text-amber-600 dark:text-amber-400 flex-shrink-0" />
              <div>
                <span class="font-bold">Scanned Local SHA-256 Computed:</span>
                <span class="block text-[11px] opacity-90 mt-0.5">No expected hash was provided in the repository metadata to verify against.</span>
              </div>
            </div>
          {/if}
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
