<script lang="ts">
  import { bookmarksStore, createBookmark, removeBookmark, browseDirectory } from '../stores/bookmarks'
  import { settingsStore } from '../stores/settings'
  import { openTaskFolder } from '../stores/queue'
  import { Bookmark, Plus, Trash2, FolderOpen, ExternalLink, HardDrive } from 'lucide-svelte'

  let newLabel: string = ''
  let newPath: string = ''
  let showAddForm: boolean = false
  let adding: boolean = false

  async function handleBrowse() {
    const chosen = await browseDirectory(newPath || '')
    if (chosen) {
      newPath = chosen
      if (!newLabel) {
        const parts = chosen.split(/[\\/]/)
        newLabel = parts[parts.length - 1] || 'My Models'
      }
    }
  }

  async function handleAddBookmark() {
    if (!newLabel.trim() || !newPath.trim()) return
    adding = true
    try {
      await createBookmark(newLabel.trim(), newPath.trim())
      newLabel = ''
      newPath = ''
      showAddForm = false
    } catch (e) {
      console.error(e)
    } finally {
      adding = false
    }
  }
</script>

<div class="bg-dark-850 border border-dark-700/60 rounded-2xl p-5 shadow-xl flex flex-col gap-4">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <Bookmark class="w-4 h-4 text-accent-cyan" />
      <h3 class="text-sm font-bold text-white tracking-tight">Folder Bookmarks & Quick Picks</h3>
    </div>
    <button
      type="button"
      on:click={() => (showAddForm = !showAddForm)}
      class="px-3 py-1.5 rounded-lg bg-dark-800 hover:bg-dark-700 text-slate-200 border border-dark-700 text-xs font-medium flex items-center gap-1.5 transition-colors"
    >
      <Plus class="w-3.5 h-3.5 text-accent-indigo" />
      <span>{showAddForm ? 'Cancel' : 'Add Bookmark'}</span>
    </button>
  </div>

  <!-- Add Bookmark Form -->
  {#if showAddForm}
    <form on:submit|preventDefault={handleAddBookmark} class="p-4 bg-dark-950/70 border border-dark-700 rounded-xl flex flex-col gap-3 animate-in fade-in duration-150">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div>
          <label for="bm-label" class="text-[11px] text-slate-400 font-medium block mb-1">Bookmark Label</label>
          <input
            id="bm-label"
            type="text"
            bind:value={newLabel}
            placeholder="e.g. ComfyUI Checkpoints"
            class="w-full px-3 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-accent-indigo"
          />
        </div>
        <div>
          <label for="bm-path" class="text-[11px] text-slate-400 font-medium block mb-1">Directory Path</label>
          <div class="flex items-center gap-2">
            <input
              id="bm-path"
              type="text"
              bind:value={newPath}
              placeholder="e.g. D:\ComfyUI\models\checkpoints"
              class="w-full px-3 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-xs text-slate-200 font-mono focus:outline-none focus:border-accent-indigo"
            />
            <button
              type="button"
              on:click={handleBrowse}
              class="px-2.5 py-1.5 bg-dark-800 hover:bg-dark-700 text-slate-300 rounded-lg border border-dark-700 text-xs flex-shrink-0"
            >
              Browse
            </button>
          </div>
        </div>
      </div>
      <div class="flex justify-end pt-1">
        <button
          type="submit"
          disabled={adding || !newLabel.trim() || !newPath.trim()}
          class="px-4 py-1.5 bg-accent-indigo hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold disabled:opacity-50 transition-colors"
        >
          Save Bookmark
        </button>
      </div>
    </form>
  {/if}

  <!-- Bookmarks Cards Grid -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
    {#each $bookmarksStore as bm}
      <div class="p-3 bg-dark-900/90 border border-dark-700/70 rounded-xl flex flex-col justify-between gap-2 group hover:border-slate-600 transition-colors">
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0 flex-1">
            <div class="text-xs font-bold text-slate-200 truncate flex items-center gap-1.5">
              <HardDrive class="w-3.5 h-3.5 text-accent-indigo flex-shrink-0" />
              <span class="truncate">{bm.label}</span>
            </div>
            <div class="text-[11px] text-slate-500 font-mono truncate mt-0.5" title={bm.path}>
              {bm.path}
            </div>
          </div>
          <button
            type="button"
            on:click={() => removeBookmark(bm.id)}
            class="opacity-0 group-hover:opacity-100 p-1 hover:text-rose-400 text-slate-500 transition-opacity"
            title="Delete bookmark"
          >
            <Trash2 class="w-3.5 h-3.5" />
          </button>
        </div>

        <button
          type="button"
          on:click={() => openTaskFolder(bm.path)}
          class="w-full py-1 text-[11px] bg-dark-800/70 hover:bg-dark-700 text-slate-400 hover:text-slate-200 rounded border border-dark-700/50 flex items-center justify-center gap-1 transition-colors"
        >
          <ExternalLink class="w-3 h-3" />
          <span>Open in Explorer</span>
        </button>
      </div>
    {/each}
  </div>

  <!-- Recent Paths -->
  {#if $settingsStore?.recentPaths && $settingsStore.recentPaths.length > 0}
    <div class="pt-3 border-t border-dark-700/50">
      <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">Recent Destination Paths:</span>
      <div class="flex items-center gap-2 flex-wrap">
        {#each $settingsStore.recentPaths as recent}
          <button
            type="button"
            on:click={() => openTaskFolder(recent)}
            class="px-2.5 py-1 bg-dark-900 hover:bg-dark-800 border border-dark-700 rounded-lg text-xs font-mono text-slate-400 hover:text-slate-200 flex items-center gap-1.5 transition-colors"
            title="Open {recent}"
          >
            <FolderOpen class="w-3 h-3 text-slate-500" />
            <span class="max-w-[200px] truncate">{recent}</span>
          </button>
        {/each}
      </div>
    </div>
  {/if}
</div>
