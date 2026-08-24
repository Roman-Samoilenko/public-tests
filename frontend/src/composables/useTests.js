import { ref } from 'vue'
import { api } from '../api/index.js'

export function useTests(defaultParams = {}) {
  const tests = ref([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref(null)

  const params = ref({
    limit: 12,
    offset: 0,
    sort: 'rating',
    filter: 'all',
    search: '',
    tags: [],
    status: undefined,
    ...defaultParams
  })

  async function loadTests(additionalParams = {}) {
    loading.value = true
    error.value = null
    try {
      const merged = { ...params.value, ...additionalParams }
      const response = await api.getTests(merged)
      tests.value = response.items || []
      total.value = response.total || 0
    } catch (e) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  function changePage(dir) {
    const newOffset = params.value.offset + dir * params.value.limit
    if (newOffset < 0) return
    params.value.offset = newOffset
    loadTests()
  }

  function resetPage() {
    params.value.offset = 0
  }

  return {
    tests,
    total,
    loading,
    error,
    params,
    loadTests,
    changePage,
    resetPage
  }
}