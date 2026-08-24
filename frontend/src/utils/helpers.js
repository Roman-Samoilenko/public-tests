export function formatDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString('ru-RU', {
    day: 'numeric', month: 'short', year: 'numeric'
  })
}

export function formatDateTime(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString('ru-RU', {
    day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit'
  })
}

export function truncate(str, len = 120) {
  return str.length > len ? str.slice(0, len) + '…' : str
}

export function reverseMapType(t) {
  const map = {
    single_choice: 'single',
    multiple_choice: 'multiple',
    scale: 'scale',
    text: 'text'
  }
  return map[t] || 'single'
}

export function axisCenter(axis) {
  return String(Math.round(((axis.min ?? -50) + (axis.max ?? 50)) / 2))
}
