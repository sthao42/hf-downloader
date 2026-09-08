<script lang="ts">
  import { bookmarksStore, createBookmark, removeBookmark, browseDirectory, reorderBookmarks } from '../stores/bookmarks'
  import { settingsStore, reorderRecentPaths } from '../stores/settings'
  import { openTaskFolder } from '../stores/queue'
  import { Bookmark, Plus, Trash2, FolderOpen, ExternalLink, HardDrive, GripVertical } from 'lucide-svelte'

  let newLabel: string = ''
  let newPath: string = ''
  let showAddForm: boolean = false
  let adding: boolean = false

  // Drag-and-drop state for saved folder bookmarks
  let draggedBMIndex: number | null = null
  let dragOverBMIndex: number | null = null

  // Drag-and-drop state for quick picks / recent paths
  let draggedRecentIndex: number | null = null
  let dragOverRecentIndex: number | null = null

  function handleDragStartBM(e: DragEvent, index: number) {
    draggedBMIndex = index
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', String(index))
    }
  }

  function handleDragOverBM(e: DragEvent, index: number) {
    e.preventDefault()
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move'
    }
    dragOverBMIndex = index
  }

  function handleDropBM(e: DragEvent, targetIndex: number) {
    e.preventDefault()
    if (draggedBMIndex === null || draggedBMIndex === targetIndex) {
      draggedBMIndex = null
      dragOverBMIndex = null
      return
    }
    const items = [...$bookmarksStore]
    const [moved] = items.splice(draggedBMIndex, 1)
    items.splice(targetIndex, 0, moved)
    reorderBookmarks(items)
    draggedBMIndex = null
    dragOverBMIndex = null
  }

  function handleDragEndBM() {
    draggedBMIndex = null
    dragOverBMIndex = null
  }

  function handleDragStartRecent(e: DragEvent, index: number) {
    draggedRecentIndex = index
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move'
      e.dataTransfer.setData('text/plain', String(index))
    }
  }

  function handleDragOverRecent(e: DragEvent, index: number) {
    e.preventDefault()
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move'
    }
    dragOverRecentIndex = index
  }

  function handleDropRecent(e: DragEvent, targetIndex: number) {
    e.preventDefault()
    const recentList = $settingsStore?.recentPaths || []
    if (draggedRecentIndex === null || draggedRecentIndex === targetIndex || !recentList.length) {
      draggedRecentIndex = null
      dragOverRecentIndex = null
      return
    }
    const items = [...recentList]
    const [moved] = items.splice(draggedRecentIndex, 1)
    items.splice(targetIndex, 0, moved)
    reorderRecentPaths(items)
    draggedRecentIndex = null
    dragOverRecentIndex = null
  }

  function handleDragEndRecent() {
    draggedRecentIndex = null
    dragOverRecentIndex = null
  }

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
  <div class="flex items-center justify-between flex-wrap gap-2">
    <div class="flex items-center gap-2">
      <Bookmark class="w-4 h-4 text-accent-cyan" />
      <h3 class="text-sm font-bold text-white tracking-tight">Folder Bookmarks & Quick Picks</h3>
      <span class="text-[10px] text-slate-500 font-mono hidden sm:inline-block ml-1">
        • Click and drag to reorder
      </span>
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

  <!-- Bookmarks Cards Grid with Click & Drag Reordering -->
  {#if $bookmarksStore.length === 0}
    <div class="py-8 text-center text-slate-500 text-xs border border-dashed border-dark-700/60 rounded-xl">
      No folder bookmarks saved yet. Click "Add Bookmark" above to save and organize your destination folders.
    </div>
  {:else}
    <div role="list" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
      {#each $bookmarksStore as bm, index (bm.id)}
        {@const isDragging = draggedBMIndex === index}
        {@const isDragOver = dragOverBMIndex === index && draggedBMIndex !== index}
        <div
          role="listitem"
          draggable="true"
          on:dragstart={(e) => handleDragStartBM(e, index)}
          on:dragover={(e) => handleDragOverBM(e, index)}
          on:dragenter={() => (dragOverBMIndex = index)}
          on:dragleave={() => { if (dragOverBMIndex === index) dragOverBMIndex = null }}
          on:drop={(e) => handleDropBM(e, index)}
          on:dragend={handleDragEndBM}
          class="p-3 bg-dark-900/90 border rounded-xl flex flex-col justify-between gap-2 group transition-all duration-150 {isDragging ? 'opacity-35 border-dashed border-accent-indigo scale-[0.98]' : isDragOver ? 'ring-2 ring-accent-indigo border-accent-indigo bg-indigo-950/25 scale-[1.02] shadow-lg shadow-indigo-500/15' : 'border-dark-700/70 hover:border-slate-600'}"
        >
          <div class="flex items-start justify-between gap-2">
            <!-- Drag Handle & Label -->
            <div class="flex items-start gap-1.5 min-w-0 flex-1">
              <div
                class="cursor-grab active:cursor-grabbing p-0.5 -ml-1 text-slate-600 group-hover:text-slate-400 hover:text-slate-200 transition-colors flex-shrink-0 mt-0.5"
                title="Click and drag to organize bookmarks"
              >
                <GripVertical class="w-4 h-4" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="text-xs font-bold text-slate-200 truncate flex items-center gap-1.5">
                  <HardDrive class="w-3.5 h-3.5 text-accent-indigo flex-shrink-0" />
                  <span class="truncate">{bm.label}</span>
                </div>
                <div class="text-[11px] text-slate-500 font-mono truncate mt-0.5" title={bm.path}>
                  {bm.path}
                </div>
              </div>
            </div>

            <button
              type="button"
              draggable="false"
              on:click={() => removeBookmark(bm.id)}
              class="opacity-0 group-hover:opacity-100 p-1 hover:text-rose-400 text-slate-500 transition-opacity flex-shrink-0"
              title="Delete bookmark"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>

          <button
            type="button"
            draggable="false"
            on:click={() => openTaskFolder(bm.path)}
            class="w-full py-1 text-[11px] bg-dark-800/70 hover:bg-dark-700 text-slate-400 hover:text-slate-200 rounded border border-dark-700/50 flex items-center justify-center gap-1 transition-colors"
          >
            <ExternalLink class="w-3 h-3" />
            <span>Open in Explorer</span>
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- Recent Destination Paths / Quick Picks with Click & Drag Reordering -->
  {#if $settingsStore?.recentPaths && $settingsStore.recentPaths.length > 0}
    <div class="pt-3 border-t border-dark-700/50">
      <div class="flex items-center justify-between mb-2">
        <span class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block">
          Quick Picks & Recent Paths:
        </span>
        <span class="text-[10px] text-slate-500 font-mono">
          Drag pills to reorganize
        </span>
      </div>
      <div role="list" class="flex items-center gap-2 flex-wrap">
        {#each $settingsStore.recentPaths as recent, rIdx (recent)}
          {@const isRecentDragging = draggedRecentIndex === rIdx}
          {@const isRecentDragOver = dragOverRecentIndex === rIdx && draggedRecentIndex !== rIdx}
          <div
            role="listitem"
            draggable="true"
            on:dragstart={(e) => handleDragStartRecent(e, rIdx)}
            on:dragover={(e) => handleDragOverRecent(e, rIdx)}
            on:dragenter={() => (dragOverRecentIndex = rIdx)}
            on:dragleave={() => { if (dragOverRecentIndex === rIdx) dragOverRecentIndex = null }}
            on:drop={(e) => handleDropRecent(e, rIdx)}
            on:dragend={handleDragEndRecent}
            class="inline-flex items-center rounded-lg border transition-all duration-150 {isRecentDragging ? 'opacity-35 border-dashed border-accent-indigo scale-95' : isRecentDragOver ? 'ring-2 ring-accent-indigo border-accent-indigo bg-indigo-950/30 scale-105 shadow-md shadow-indigo-500/10' : 'bg-dark-900 hover:bg-dark-800 border-dark-700'}"
          >
            <div
              class="cursor-grab active:cursor-grabbing pl-2 pr-0.5 py-1 text-slate-600 hover:text-slate-300 transition-colors"
              title="Drag to organize quick pick"
            >
              <GripVertical class="w-3 h-3" />
            </div>
            <button
              type="button"
              draggable="false"
              on:click={() => openTaskFolder(recent)}
              class="pr-2.5 py-1 text-xs font-mono text-slate-400 hover:text-slate-200 flex items-center gap-1.5 transition-colors"
              title="Open {recent}"
            >
              <FolderOpen class="w-3 h-3 text-slate-500" />
              <span class="max-w-[200px] truncate">{recent}</span>
            </button>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

