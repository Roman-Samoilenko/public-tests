<template>
  <div>
    <!-- Переключатель источника -->
    <div class="import-source-selector">
      <label v-for="source in sources" :key="source.key" class="source-option">
        <input type="radio" v-model="selectedSource" :value="source.key" />
        <span>{{ source.label }}</span>
      </label>
    </div>

    <!-- Панель ввода (в зависимости от источника) -->
    <div v-if="!importedTest" class="import-panel">
      <div class="import-inner">
        <p class="import-label">{{ currentSource.label }}</p>

        <!-- Поле для ввода ссылки (для Google Forms) -->
        <div v-if="selectedSource === 'google'" class="import-row">
          <input v-model="url" type="url" placeholder="https://docs.google.com/forms/d/..." />
          <button class="btn" :disabled="loading" @click="doImport">
            <span v-if="loading" class="spinner"></span>
            <span v-else>Загрузить</span>
          </button>
        </div>

        <!-- Заготовка для других источников (JSON, CSV, Яндекс) -->
        <div v-else-if="selectedSource === 'json'" class="import-row">
          <input type="file" accept=".json" @change="onFileUpload" />
          <button class="btn" :disabled="loading" @click="importFromJSON">Загрузить</button>
        </div>
        <div v-else-if="selectedSource === 'csv'" class="import-row">
          <input type="file" accept=".csv" @change="onFileUpload" />
          <button class="btn" :disabled="loading" @click="importFromCSV">Загрузить</button>
        </div>

        <p v-if="error" class="error-msg">{{ error }}</p>
      </div>
    </div>

    <!-- Превью импортированного теста (общее для всех источников) -->
    <div v-else class="import-preview">
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
        <button class="btn-ghost" @click="reset">Отмена</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/index.js'
import { useImportedTestStore } from '../store/importedTest.js'

const router = useRouter()
const store = useImportedTestStore()

// Список доступных источников
const sources = [
  { key: 'google', label: 'Google Forms' },
  { key: 'json', label: 'JSON' },
  { key: 'csv', label: 'CSV' },
  // можно добавить 'yandex', 'typeform' и т.д.
]
const selectedSource = ref('google')

// Общие состояния
const url = ref('')
const file = ref(null)
const error = ref('')
const loading = ref(false)
const importedTest = ref(null)

// Логика импорта
async function doImport() {
  error.value = ''
  if (!url.value.trim()) { error.value = 'Введите ссылку'; return }
  loading.value = true
  try {
    importedTest.value = await api.importGoogleForm(url.value)
  } catch (e) {
    error.value = e.data?.error || e.message
  } finally {
    loading.value = false
  }
}

// Заготовка для JSON
function onFileUpload(event) {
  file.value = event.target.files[0]
}
async function importFromJSON() {
  // Здесь будет вызов api.importJSON(file)
  // Пока просто заглушка
  error.value = 'Импорт из JSON ещё не реализован'
}

// Заготовка для CSV
async function importFromCSV() {
  error.value = 'Импорт из CSV ещё не реализован'
}

// Общие действия
function editImported() {
  store.setImportedTest(importedTest.value)
  router.push('/create?import=true')
}
function reset() {
  importedTest.value = null
  url.value = ''
  file.value = null
  error.value = ''
}
</script>

<style scoped>
/* Стили для переключателя, если нужны */
.import-source-selector {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}
.source-option {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  cursor: pointer;
  font-family: var(--font-mono);
  font-size: 0.8rem;
  color: var(--text-muted);
}
.source-option input { width: auto; }
.source-option span { transition: color var(--transition); }
.source-option:has(input:checked) span { color: var(--accent); }
</style>