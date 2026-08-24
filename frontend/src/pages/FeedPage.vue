<template>
  <div class="page-wide">

    <!-- Заголовок -->
    <div class="feed-header">
      <div>
        <h1 class="feed-title">Лента тестов</h1>
        <p class="feed-sub">Исследования сообщества</p>
      </div>
      <div class="feed-actions">
        <router-link to="/create" class="btn">+ Создать тест</router-link>
        <button class="btn btn-outline" @click="showImport = !showImport">⊕ Импорт</button>
      </div>
    </div>

    <!-- Панель импорта из Google Forms -->
    <div v-if="showImport" class="import-panel">
      <div class="import-inner">
        <p class="import-label">Импорт из Google Forms</p>
        <div class="import-row">
          <input v-model="importUrl" type="url" placeholder="https://docs.google.com/forms/d/..." />
          <button class="btn" :disabled="importLoading" @click="doImport">
            <span v-if="importLoading" class="spinner"></span>
            <span v-else>Загрузить</span>
          </button>
        </div>
        <p v-if="importError" class="error-msg">{{ importError }}</p>
      </div>
    </div>

    <!-- Если импорт успешен — превью и форма публикации -->
    <div v-if="importedTest" class="import-preview">
      <div class="preview-header">
        <div>
          <p class="import-label">Предварительный просмотр</p>
          <h2 class="preview-title">{{ importedTest.title }}</h2>
          <p class="preview-desc">{{ importedTest.description }}</p>
        </div>
        <div class="preview-meta">
          <span class="tag">{{ importedTest.questions.length }} вопросов</span>
        </div>
      </div>
      <div class="preview-actions">
  <button class="btn" @click="editImported">Редактировать</button>
  <button class="btn-ghost" @click="importedTest = null">Отмена</button>
</div>
    </div>

    <!-- Поиск -->
    <div class="search-row">
      <input
        v-model="searchQuery"
        type="text"
        class="search-input"
        placeholder="Поиск по названию, описанию, тегам..."
        @input="debouncedSearch"
      />
    </div>

    <!-- Активные теги-фильтры -->
    <div v-if="activeTags.length" class="active-tags">
      <span class="active-tags-label font-mono">Теги:</span>
      <span
        v-for="tag in activeTags"
        :key="tag"
        class="tag tag-removable"
        @click="removeTag(tag)"
      >{{ tag }} ×</span>
    </div>

    <!-- Фильтры и сортировка -->
    <div class="filters">
      <div class="filter-tabs">
        <button
          v-for="f in filterOptions" :key="f.value"
          :class="['filter-tab', filter === f.value && 'active']"
          @click="setFilter(f.value)"
        >{{ f.label }}</button>
      </div>
      <div class="sort-select">
        <select v-model="sort" @change="loadTests(true)">
          <option value="rating">По рейтингу</option>
          <option value="newest">По дате</option>
          <option value="pass_count">По популярности</option>
          <option value="comments">По комментариям</option>
        </select>
      </div>
    </div>

    <!-- Список тестов -->
    <div v-if="loading && !tests.length" class="loading-state">
      <div class="spinner" style="width:32px;height:32px"></div>
    </div>

    <div v-else-if="!tests.length" class="empty-state">
      <p>Тестов пока нет</p>
    </div>

    <div v-else class="tests-grid">
      <TestCard
        v-for="t in tests"
        :key="t.id"
        :test="t"
        @filter-tag="filterByTag"
      />
    </div>

    <!-- Пагинация -->
    <div v-if="total > limit" class="pagination">
      <button
        class="page-btn"
        :disabled="offset === 0"
        @click="changePage(-1)"
      >← Назад</button>

      <span class="page-info font-mono">
        {{ Math.floor(offset / limit) + 1 }} / {{ Math.ceil(total / limit) }}
      </span>

      <button
        class="page-btn"
        :disabled="offset + limit >= total"
        @click="changePage(1)"
      >Вперёд →</button>
    </div>

  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/index.js'
import TestCard from '../components/TestCard.vue'
import { useImportedTestStore } from '../store/importedTest.js' // создайте такой стор
const importedStore = useImportedTestStore()

function editImported() {
  // Сохраняем импортированные данные в стор
  importedStore.setImportedTest(importedTest.value)
  // Переходим на страницу создания
  router.push('/create?import=true')
}
const router = useRouter()

const tests   = ref([])
const total   = ref(0)
const loading = ref(false)
const sort    = ref('rating')
const filter  = ref('all')
const limit   = 12
const offset  = ref(0)

const searchQuery = ref('')
const activeTags  = ref([])

const filterOptions = [
  { value: 'all',      label: 'Все' },
  { value: 'official', label: 'Официальные' },
  { value: 'my',       label: 'Мои' },
]

const showImport    = ref(false)
const importUrl     = ref('')
const importError   = ref('')
const importLoading = ref(false)
const importedTest  = ref(null)
const publishLoading = ref(false)

// Debounce без lodash
let searchTimer = null
function debouncedSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadTests(true), 400)
}

async function loadTests(reset = false) {
  if (reset) offset.value = 0
  loading.value = true
  try {
    const params = {
      sort:   sort.value,
      limit,
      offset: offset.value,
    }
    if (searchQuery.value)   params.search   = searchQuery.value
    if (activeTags.value.length) params.tags = activeTags.value
    if (filter.value === 'official') params.official = 1
    if (filter.value === 'my')       params.my = 1

    const data = await api.getTests(params)
    tests.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function setFilter(val) {
  filter.value = val
  loadTests(true)
}

function filterByTag(tag) {
  if (!activeTags.value.includes(tag)) {
    activeTags.value.push(tag)
    loadTests(true)
  }
}

function removeTag(tag) {
  activeTags.value = activeTags.value.filter(t => t !== tag)
  loadTests(true)
}

function changePage(dir) {
  offset.value = Math.max(0, offset.value + dir * limit)
  loadTests()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function doImport() {
  importError.value = ''
  importedTest.value = null
  if (!importUrl.value.trim()) { importError.value = 'Введите ссылку'; return }
  importLoading.value = true
  try {
    importedTest.value = await api.importGoogleForm(importUrl.value)
    showImport.value = false
  } catch (e) {
    importError.value = e.data?.error || e.message
  } finally {
    importLoading.value = false
  }
}

async function publishImported() {
  publishLoading.value = true
  try {
    const t = await api.createTest({
      title:       importedTest.value.title,
      description: importedTest.value.description,
      questions:   importedTest.value.questions,
    })
    importedTest.value = null
    importUrl.value = ''
    router.push(`/tests/${t.id}`)
  } catch (e) {
    console.error(e)
  } finally {
    publishLoading.value = false
  }
}

onMounted(() => loadTests())
</script>

<style scoped>
.feed-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 2rem;
  gap: 1rem;
  flex-wrap: wrap;
}
.feed-title {
  font-family: var(--font-serif);
  font-size: 2.2rem;
  font-weight: 700;
  line-height: 1.1;
}
.feed-sub {
  color: var(--text-muted);
  font-size: 0.9rem;
  margin-top: 0.3rem;
}
.feed-actions { display: flex; gap: 0.8rem; }

/* Поиск */
.search-row {
  margin-bottom: 1rem;
}
.search-input {
  width: 100%;
  max-width: 520px;
}

/* Активные теги */
.active-tags {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.8rem;
}
.active-tags-label {
  font-family: var(--font-mono);
  font-size: 0.7rem;
  letter-spacing: 0.06em;
  color: var(--text-muted);
  text-transform: uppercase;
}
.tag-removable { cursor: pointer; }
.tag-removable:hover { opacity: 0.75; }

/* Импорт */
.import-panel {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 1.2rem 1.5rem;
  margin-bottom: 1.5rem;
}
.import-inner { max-width: 600px; }
.import-label {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.7rem;
}
.import-row {
  display: flex;
  gap: 0.8rem;
  align-items: stretch;
}
.import-row input { flex: 1; }

/* Превью импорта */
.import-preview {
  background: var(--bg-card);
  border: 1px solid rgba(212,168,67,0.25);
  border-radius: 6px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}
.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.2rem;
}
.preview-title {
  font-family: var(--font-serif);
  font-size: 1.3rem;
  margin: 0.3rem 0;
}
.preview-desc { font-size: 0.88rem; color: var(--text-muted); }
.preview-actions { display: flex; gap: 1rem; align-items: center; }

/* Фильтры */
.filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.8rem;
  flex-wrap: wrap;
}
.filter-tabs { display: flex; gap: 0.3rem; }
.filter-tab {
  padding: 0.4rem 1rem;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-muted);
  border: 1px solid transparent;
  border-radius: var(--radius);
  transition: all var(--transition);
}
.filter-tab:hover { color: var(--text); }
.filter-tab.active {
  color: var(--accent);
  border-color: var(--accent);
  background: var(--accent-dim);
}

.sort-select select { width: auto; }

/* Сетка тестов */
.tests-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

/* Пагинация */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
}
.page-btn {
  font-family: var(--font-mono);
  font-size: 0.8rem;
  color: var(--text-muted);
  transition: color var(--transition);
}
.page-btn:hover:not(:disabled) { color: var(--accent); }
.page-btn:disabled { opacity: 0.3; }
.page-info {
  font-family: var(--font-mono);
  font-size: 0.8rem;
  color: var(--text-muted);
}

.loading-state, .empty-state {
  display: flex;
  justify-content: center;
  padding: 4rem 0;
  color: var(--text-muted);
}
</style>
