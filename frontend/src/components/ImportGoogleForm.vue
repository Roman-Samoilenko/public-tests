<template>
  <div>
    <!-- Панель ввода ссылки -->
    <div v-if="!importedTest" class="import-panel">
      <div class="import-inner">
        <p class="import-label">Импорт из Google Forms</p>
        <div class="import-row">
          <input v-model="url" type="url" placeholder="https://docs.google.com/forms/d/..." />
          <button class="btn" :disabled="loading" @click="doImport">
            <span v-if="loading" class="spinner"></span>
            <span v-else>Загрузить</span>
          </button>
        </div>
        <p v-if="error" class="error-msg">{{ error }}</p>
      </div>
    </div>

    <!-- Превью импортированного теста -->
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

const url   = ref('')
const error = ref('')
const loading = ref(false)
const importedTest = ref(null)

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

function editImported() {
  store.setImportedTest(importedTest.value)
  router.push('/create?import=true')
}

function reset() {
  importedTest.value = null
  url.value = ''
}
</script>