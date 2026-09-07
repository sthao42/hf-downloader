<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import type { FileNode } from '../types'
  import { bookmarksStore, browseDirectory } from '../stores/bookmarks'
  import { formatBytes } from '../utils'
  import { X, Folder, HardDrive, Check, FolderOpen, SlidersHorizontal } from 'lucide-svelte'

  export let show: boolean = false
  export let targetFiles: { file: FileNode; dest: string }[] = []

  const dispatch = createEventDispatcher<{
    save: { files: { file: FileNode; dest: string }[] }
    close: void
  }>()

  let editableFiles: { file: FileNode; dest: string }[] = []

  $: if (show && targetFiles) {
    editableFiles = targetFiles.map(t => ({ ...t }))
  }

  async function handleBrowseAll() {
    const chosen = await browseDirectory('')
    if (chosen) {
      editableFiles = editableFiles.map(f => ({ ...f, dest: chosen }))
    }
  }

  function applyBookmarkToAll(path: string) {
    editableFiles = editableFiles.map(f => ({ ...f, dest: path }))
  }

  async function handleBrowseSingle(index: number) {
    const current = editableFiles[index]?.dest || ''
    const chosen = await browseDirectory(current)
    if (chosen) {
      editableFiles[index].dest = chosen
    }
  }

  function handleSave() {
    dispatch('save', { files: editableFiles })
    dispatch('close')
  }
</script>

{#if show}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
    <div class="bg-dark-900 border border-dark-700 rounded-2xl w-full max-w-3xl max-h-[85vh] shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150">
      <!-- Modal Header -->
      <div class="flex items-center justify-between p-4 px-6 border-b border-dark-700/70 bg-dark-850">
        <div class="flex items-center gap-2">
          <SlidersHorizontal class="w-5 h-5 text-accent-indigo" />
          <h3 class="text-base font-bold text-white">Route File Destinations</h3>
        </div>
        <button
          type="button"
          on:click={() => dispatch('close')}
          class="p-1 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-dark-800 transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Quick Batch Assignment Bar -->
      <div class="p-4 px-6 bg-dark-950/60 border-b border-dark-800 flex flex-col gap-2">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-300 uppercase tracking-wider">
            Quick Route All Files:
          </span>
          <button
            type="button"
            on:click={handleBrowseAll}
            class="px-2.5 py-1 text-xs bg-dark-800 hover:bg-dark-700 text-slate-200 rounded-lg border border-dark-700 flex items-center gap-1.5 transition-colors"
          >
            <FolderOpen class="w-3.5 h-3.5 text-accent-cyan" />
            <span>Browse Directory for All</span>
          </button>
        </div>

        <!-- Bookmarks Quick Picks -->
        <div class="flex items-center gap-2 flex-wrap">
          {#each $bookmarksStore as bm}
            <button
              type="button"
              on:click={() => applyBookmarkToAll(bm.path)}
              class="px-2.5 py-1 rounded-lg text-xs bg-dark-800 hover:bg-indigo-950/40 text-slate-300 hover:text-accent-indigo border border-dark-700 hover:border-accent-indigo/40 transition-colors flex items-center gap-1.5"
            >
              <Folder class="w-3 h-3 text-slate-500" />
              <span>{bm.label}</span>
            </button>
          {/each}
        </div>
      </div>

      <!-- Scrollable File Routing Table -->
      <div class="flex-1 overflow-y-auto p-6 divide-y divide-dark-800/80">
        {#each editableFiles as item, idx}
          <div class="py-3 flex items-center justify-between gap-4 first:pt-0 last:pb-0">
            <div class="min-w-0 flex-1">
              <div class="text-xs font-mono font-medium text-slate-200 truncate">
                {item.file.path}
              </div>
              <div class="text-[11px] text-slate-500 font-mono mt-0.5">
                {formatBytes(item.file.size)}
              </div>
            </div>

            <!-- Destination path with browse button -->
            <div class="flex items-center gap-2 flex-1 max-w-sm">
              <div class="flex items-center gap-1.5 px-2.5 py-1.5 bg-dark-950 border border-dark-700 rounded-lg text-xs text-slate-300 font-mono flex-1 truncate">
                <HardDrive class="w-3.5 h-3.5 text-slate-500 flex-shrink-0" />
                <span class="truncate">{item.dest || 'Not assigned'}</span>
              </div>
              <button
                type="button"
                on:click={() => handleBrowseSingle(idx)}
                class="px-2 py-1.5 bg-dark-800 hover:bg-dark-700 text-slate-300 rounded-lg border border-dark-700 text-xs transition-colors flex-shrink-0"
                title="Browse folder for this file"
              >
                Browse
              </button>
            </div>
          </div>
        {/each}
      </div>

      <!-- Modal Footer -->
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
          class="px-5 py-2 rounded-xl bg-accent-indigo hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/25 flex items-center gap-1.5 transition-colors"
        >
          <Check class="w-4 h-4" />
          <span>Apply Routing</span>
        </button>
      </div>
    </div>
  </div>
{/if}
