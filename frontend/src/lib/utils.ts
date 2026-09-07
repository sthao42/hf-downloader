export function formatBytes(bytes: number, decimals: number = 2): string {
  if (bytes === 0 || !bytes) return '0 B'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

export function detectQuantBadge(filename: string): { label: string; color: string } | null {
  const upper = filename.toUpperCase()

  // GGUF / ExLlama quants
  if (upper.includes('Q4_K_M') || upper.includes('Q4_0') || upper.includes('Q4_1')) {
    return { label: 'Q4 Quant', color: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/40' }
  }
  if (upper.includes('Q5_K_M') || upper.includes('Q5_0')) {
    return { label: 'Q5 Quant', color: 'bg-teal-500/20 text-teal-400 border-teal-500/40' }
  }
  if (upper.includes('Q8_0') || upper.includes('Q8_K')) {
    return { label: 'Q8 Quant', color: 'bg-cyan-500/20 text-cyan-400 border-cyan-500/40' }
  }

  // Precision formats
  if (upper.includes('FP8') || upper.includes('E4M3FN') || upper.includes('E5M2')) {
    return { label: 'FP8 Fast', color: 'bg-amber-500/20 text-amber-400 border-amber-500/40' }
  }
  if (upper.includes('FP16')) {
    return { label: 'FP16 Half', color: 'bg-indigo-500/20 text-indigo-400 border-indigo-500/40' }
  }
  if (upper.includes('BF16')) {
    return { label: 'BF16', color: 'bg-blue-500/20 text-blue-400 border-blue-500/40' }
  }
  if (upper.includes('FP32')) {
    return { label: 'FP32 Full', color: 'bg-purple-500/20 text-purple-400 border-purple-500/40' }
  }

  // Type tags
  if (upper.includes('VAE')) {
    return { label: 'VAE', color: 'bg-rose-500/20 text-rose-400 border-rose-500/40' }
  }
  if (upper.includes('CLIP') || upper.includes('T5XXL')) {
    return { label: 'Text Encoder', color: 'bg-sky-500/20 text-sky-400 border-sky-500/40' }
  }
  if (upper.includes('LORA')) {
    return { label: 'LoRA', color: 'bg-pink-500/20 text-pink-400 border-pink-500/40' }
  }

  return null
}
