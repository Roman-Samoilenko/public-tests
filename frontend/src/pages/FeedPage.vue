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

    <!-- Панель импорта (используем компонент ImportForm) -->
    <div v-if="showImport" class="import-panel">
      <ImportForm @imported="onImported" />
    </div>

    <!-- Превью импортированного теста -->
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

    <!-- Активные теги -->
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
import { onMounted, ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/index.js'
import TestCard from '../components/TestCard.vue'
import { useImportedTestStore } from '../store/importedTest.js'
import { useTests } from '../composables/useTests.js'
import ImportForm from '../components/ImportForm.vue'

const router = useRouter()
const store = useImportedTestStore()

// Состояние фильтров
const searchQuery = ref('')
const activeTags = ref([])
const filter = ref('all')
const sort = ref('rating')

// --- Добавляем недостающие переменные ---
const importedTest = ref(null)               // для превью после импорта
const showImport = ref(false)

// Фильтры (пример)
const filterOptions = [
  { value: 'all', label: 'Все' },
  { value: 'popular', label: 'Популярные' },
  { value: 'new', label: 'Новые' },
  { value: 'my', label: 'Мои тесты' },
]

// Используем композабл — он должен возвращать limit и offset
const { tests, total, loading, params, loadTests, changePage, resetPage, limit, offset } = useTests({
  limit: 12,
  sort: 'rating',
  filter: 'all'
})

// Если useTests не возвращает limit/offset, объявите их вручную:
// const limit = ref(12)
// const offset = ref(0)
// и синхронизируйте с пагинацией.

// При изменении фильтров перезагружаем
watch([searchQuery, activeTags, filter, sort], () => {
  params.value.search = searchQuery.value
  params.value.tags = activeTags.value
  params.value.filter = filter.value
  params.value.sort = sort.value
  resetPage()
  loadTests()
})

onMounted(() => loadTests())

// Обработчик события от ImportForm
function onImported(testData) {
  importedTest.value = testData
  // можно также сохранить в store, если нужно
  store.setImportedTest(testData)
}

// Функции для тегов
function filterByTag(tag) {
  if (!activeTags.value.includes(tag)) {
    activeTags.value.push(tag)
  }
}
function removeTag(tag) {
  activeTags.value = activeTags.value.filter(t => t !== tag)
}

function editImported() {
  // например, перейти на страницу создания с импортированными данными
  router.push('/create?import=true')
}

// функция setFilter (если нужна)
function setFilter(value) {
  filter.value = value
}

// debouncedSearch (если нужна)
function debouncedSearch() {
  // можно реализовать debounce через lodash или простой setTimeout
}
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
/* Фильтры */

</style>
