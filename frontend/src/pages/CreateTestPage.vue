<template>
  <div class="page">

    <div class="create-header">
      <router-link :to="editMode ? `/tests/${testId}` : '/'" class="back-link font-mono">
        {{ editMode ? '← К тесту' : '← Лента' }}
      </router-link>
      <h1 class="create-title">{{ editMode ? 'Редактировать тест' : 'Создать тест' }}</h1>
      <div class="create-header-actions">
  <button class="btn btn-outline" @click="showImport = !showImport">⊕ Импорт из Google Forms</button>
</div>

<!-- Панель импорта -->
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
    </div>

    <div v-if="pageLoading" class="loading-state" style="min-height:40vh">
      <div class="spinner" style="width:32px;height:32px"></div>
    </div>

    <template v-else>
      <p v-if="error" class="error-msg" style="margin-bottom:1rem">{{ error }}</p>


      <!-- ─── Секция 1: Основная информация ─── -->
      <section class="section">
        <h2 class="section-title">Основная информация</h2>

        <div class="field">
          <label class="field-label font-mono">Название *</label>
          <input v-model="form.title" type="text" placeholder="Введите название теста" maxlength="200" />
        </div>

        <div class="field">
          <label class="field-label font-mono">Описание</label>
          <textarea v-model="form.description" rows="3" placeholder="Краткое описание теста..." maxlength="2000"></textarea>
        </div>

        <div class="field">
          <label class="field-label font-mono">Теги</label>
          <div class="tags-input">
            <div class="tags-chips">
              <span v-for="tag in form.tags" :key="tag" class="tag tag-chip">
                {{ tag }}<button class="chip-remove" @click="removeTag(tag)">×</button>
              </span>
            </div>
            <input
              v-model="tagInput"
              type="text"
              placeholder="Добавить тег (Enter или запятая)"
              @keydown.enter.prevent="addTag"
              @keydown.comma.prevent="addTag"
              @blur="addTag"
            />
          </div>
          <p class="field-hint font-mono">Нажмите Enter или запятую для добавления тега</p>
        </div>
      </section>

      <div class="divider"></div>

      <!-- ─── Секция 2: Вопросы ─── -->
      <section class="section">
        <h2 class="section-title">Вопросы</h2>

        <div class="questions-list">
          <div v-for="(q, i) in form.questions" :key="q.id" class="question-card">
            <div class="question-header">
              <span class="question-num font-mono">{{ i + 1 }}</span>
              <div class="question-controls">
                <button class="ctrl-btn" :disabled="i === 0" @click="moveQuestion(i, -1)" title="Вверх">↑</button>
                <button class="ctrl-btn" :disabled="i === form.questions.length - 1" @click="moveQuestion(i, 1)" title="Вниз">↓</button>
                <button class="ctrl-btn" @click="copyQuestion(i)" title="Копировать вопрос">⧉</button>
                <button
                  class="ctrl-btn ctrl-danger"
                  :disabled="form.questions.length === 1"
                  @click="removeQuestion(i)"
                  title="Удалить"
                >✕</button>
              </div>
            </div>

            <div class="question-body">
              <textarea v-model="q.text" rows="2" placeholder="Текст вопроса *" style="resize:vertical"></textarea>

              <div class="question-meta">
                <div class="field-inline">
                  <label class="field-label font-mono">Тип</label>
                  <select v-model="q.type" @change="onTypeChange(q)">
                    <option value="single">Один вариант</option>
                    <option value="multiple">Несколько вариантов</option>
                    <option value="scale">Шкала</option>
                    <option value="text">Текстовый ответ</option>
                  </select>
                </div>
                <label class="required-check">
                  <input type="checkbox" v-model="q.required" />
                  <span class="font-mono">Обязательный</span>
                </label>
              </div>

              <!-- Варианты ответов -->
              <div v-if="q.type === 'single' || q.type === 'multiple'" class="options-list">
                <div v-for="(opt, oi) in q.options" :key="opt.id" class="option-row-wrap">
                  <div class="option-row">
                    <span class="option-bullet font-mono">{{ oi + 1 }}.</span>
                    <input v-model="opt.text" type="text" placeholder="Вариант ответа" />
                    <button
                      class="ctrl-btn ctrl-danger"
                      :disabled="q.options.length <= 2"
                      @click="removeOption(q, oi)"
                    >×</button>
                  </div>
                  <div v-if="resultType !== 'none'" class="scoring-row">
                    <template v-if="resultType === 'score'">
                      <label class="scoring-label font-mono">Вес</label>
                      <input v-model.number="opt.scoring['score']" type="number" class="scoring-input" placeholder="0" />
                    </template>
                    <template v-else>
                      <template v-for="axis in form.result_config.axes" :key="axis.id">
                        <label class="scoring-label font-mono">{{ axis.label || axis.id }}</label>
                        <input v-model.number="opt.scoring[axis.id]" type="number" class="scoring-input" placeholder="0" />
                      </template>
                      <span v-if="!form.result_config.axes.length" class="scoring-hint font-mono">Сначала добавьте оси</span>
                    </template>
                  </div>
                </div>
                <button class="btn-ghost add-option-btn" @click="addOption(q)">+ вариант</button>
              </div>

              <!-- Шкала -->
              <div v-if="q.type === 'scale'" class="scale-row">
                <div class="field-inline">
                  <label class="field-label font-mono">Мин</label>
                  <input v-model.number="q.min" type="number" style="width:80px" />
                </div>
                <div class="field-inline">
                  <label class="field-label font-mono">Макс</label>
                  <input v-model.number="q.max" type="number" style="width:80px" />
                </div>
              </div>

              <!-- Коэффициенты шкального вопроса -->
              <div v-if="q.type === 'scale' && resultType !== 'none'" class="scoring-row scoring-row-scale">
                <span class="scoring-label font-mono" style="margin-right:0.3rem">Коэффициент:</span>
                <template v-if="resultType === 'score'">
                  <label class="scoring-label font-mono">Вес</label>
                  <input v-model.number="q.scoring['score']" type="number" class="scoring-input" placeholder="1" />
                </template>
                <template v-else>
                  <template v-for="axis in form.result_config.axes" :key="axis.id">
                    <label class="scoring-label font-mono">{{ axis.label || axis.id }}</label>
                    <input v-model.number="q.scoring[axis.id]" type="number" class="scoring-input" placeholder="1" />
                  </template>
                  <span v-if="!form.result_config.axes.length" class="scoring-hint font-mono">Сначала добавьте оси</span>
                </template>
              </div>
            </div>
          </div>
        </div>

        <button class="btn btn-outline add-question-btn" @click="addQuestion">+ Добавить вопрос</button>
      </section>

      <div class="divider"></div>

      <!-- ─── Секция 3: Конфигурация результата ─── -->
      <section class="section">
        <h2 class="section-title">Конфигурация результата</h2>

        <div class="result-type-row">
          <label v-for="rt in resultTypes" :key="rt.value" class="result-type-option">
            <input type="radio" v-model="resultType" :value="rt.value" />
            <span class="font-mono">{{ rt.label }}</span>
          </label>
        </div>

        <!-- Балл -->
        <div v-if="resultType === 'score'" class="result-config-block">
          <div class="scale-row">
            <div class="field-inline">
              <label class="field-label font-mono">Мин баллов</label>
              <input v-model.number="form.result_config.min" type="number" style="width:100px" />
            </div>
            <div class="field-inline">
              <label class="field-label font-mono">Макс баллов</label>
              <input v-model.number="form.result_config.max" type="number" style="width:100px" />
            </div>
          </div>
          <div class="levels-list">
            <p class="field-label font-mono" style="margin-bottom:0.5rem">Уровни</p>
            <div v-for="(lvl, li) in form.result_config.levels" :key="li" class="level-row">
              <input v-model.number="lvl.min" type="number" placeholder="от" style="width:80px" />
              <input v-model.number="lvl.max" type="number" placeholder="до" style="width:80px" />
              <input v-model="lvl.label" type="text" placeholder="Название уровня" />
              <button class="ctrl-btn ctrl-danger" @click="removeLevel(li)">×</button>
            </div>
            <button class="btn-ghost add-option-btn" @click="addLevel">+ уровень</button>
          </div>
        </div>

        <!-- Оси -->
        <div v-if="resultType === 'scales' || resultType === 'string_map' || resultType === 'combo'" class="result-config-block">
          <p class="field-label font-mono" style="margin-bottom:0.5rem">Оси</p>
          <div v-for="(axis, ai) in form.result_config.axes" :key="ai" class="axis-editor">
            <div class="axis-editor-row">
              <input v-model="axis.id"    type="text" placeholder="id (латиница, без пробелов)" style="width:160px" />
              <input v-model="axis.label" type="text" placeholder="Название оси" style="flex:1" />
              <button class="ctrl-btn ctrl-danger" @click="removeAxis(ai)">×</button>
            </div>
            <div class="axis-editor-row axis-range-row">
              <label class="field-label font-mono" style="margin:0;white-space:nowrap">Диапазон:</label>
              <input v-model.number="axis.min" type="number" style="width:80px" />
              <span class="font-mono" style="color:var(--text-muted);font-size:0.8rem">—</span>
              <input v-model.number="axis.max" type="number" style="width:80px" />
              <input v-model="axis.left_label"  type="text" placeholder="Левая метка" style="flex:1" />
              <input v-model="axis.right_label" type="text" placeholder="Правая метка" style="flex:1" />
            </div>
          </div>
          <button class="btn-ghost add-option-btn" @click="addAxis">+ ось</button>
        </div>

        <!-- Результаты типов -->
        <div v-if="resultType === 'string_map' || resultType === 'combo'" class="result-config-block">
          <p class="field-label font-mono" style="margin-bottom:0.3rem">Результаты (типы)</p>
          <p class="how-it-works">
            <strong>Как это работает:</strong> каждый вопрос имеет веса по осям. После прохождения
            система считает итоговые значения и находит результат, чей <em>идеальный профиль</em>
            ближе всего к реальному.
          </p>
          <div v-if="!form.result_config.axes.length" class="how-it-works" style="color:var(--danger)">
            ⚠ Сначала добавьте хотя бы одну ось.
          </div>
          <div v-for="(res, ri) in form.result_config.results" :key="ri" class="result-entry">
            <div class="result-entry-header">
              <input v-model="res.key"   type="text" placeholder="key (латиница)" style="width:140px" />
              <input v-model="res.label" type="text" placeholder="Название типа" style="flex:1" />
              <button class="ctrl-btn ctrl-danger" @click="removeResult(ri)">×</button>
            </div>
            <input v-model="res.description" type="text" placeholder="Описание типа" style="width:100%" />
            <div v-if="form.result_config.axes.length" class="target-row">
              <span class="scoring-label font-mono" style="white-space:nowrap;margin-right:0.4rem">Цель по осям:</span>
              <template v-for="axis in form.result_config.axes" :key="axis.id">
                <label class="scoring-label font-mono">{{ axis.label || axis.id }}</label>
                <input
                  v-model.number="res.target[axis.id]"
                  type="number"
                  class="scoring-input"
                  :placeholder="axisCenter(axis)"
                  :title="`Диапазон: ${axis.min} — ${axis.max}`"
                />
              </template>
            </div>
          </div>
          <button class="btn-ghost add-option-btn" @click="addResult">+ результат</button>
        </div>

        <!-- Пояснение автора -->
        <div v-if="resultType !== 'none'" class="field" style="margin-top:1rem">
          <label class="field-label font-mono">Пояснение автора (опционально)</label>
          <textarea
            v-model="form.result_config.description"
            rows="3"
            placeholder="Текст, который увидит пользователь под результатом..."
            maxlength="2000"
          ></textarea>
        </div>
      </section>

      <div class="divider"></div>

      <div class="publish-row">
        <button class="btn" :disabled="publishing" @click="publish">
          <span v-if="publishing" class="spinner"></span>
          <span v-else>{{ editMode ? 'Сохранить изменения →' : 'Опубликовать тест →' }}</span>
        </button>
      </div>
    </template>

  </div>
</template>

<script setup>
import { nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/index.js'
import { authStore } from '../store/auth.js'
import { useImportedTestStore } from '../store/importedTest.js'
const importedStore = useImportedTestStore()

const route  = useRouter()
const router = useRouter()
const $route = useRoute()

const testId   = $route.params.id ? Number($route.params.id) : null
const editMode = !!testId

const error      = ref('')
const publishing = ref(false)
const pageLoading = ref(editMode) // true только если редактируем — ждём загрузки
const tagInput   = ref('')

// Импорт из Google Forms
const showImport = ref(false)
const importUrl = ref('')
const importError = ref('')
const importLoading = ref(false)

// флаг подавления watch во время загрузки формы
const suppressWatch = ref(false)

const resultTypes = [
  { value: 'none',       label: 'Без результата' },
  { value: 'score',      label: 'Балл' },
  { value: 'scales',     label: 'Шкалы' },
  { value: 'string_map', label: 'Тип' },
  { value: 'combo',      label: 'Комбо' },
]
const resultType = ref('none')

function makeQuestion() {
  return {
    id: crypto.randomUUID(),
    text: '',
    type: 'single',
    required: false,
    options: [
      { id: crypto.randomUUID(), text: '', scoring: {} },
      { id: crypto.randomUUID(), text: '', scoring: {} },
    ],
    min: 1,
    max: 10,
    scoring: {},
  }
}

const form = reactive({
  title: '',
  description: '',
  tags: [],
  questions: [makeQuestion()],
  result_config: {
    min: 0, max: 100,
    levels: [],
    axes: [],
    results: [],
    description: '',
  },
})

// Сбрасываем result_config при смене типа — только не во время загрузки
watch(resultType, (val) => {
  if (suppressWatch.value) return
  form.result_config = {
    type: val === 'none' ? undefined : val,
    min: 0, max: 100,
    levels: [],
    axes: [],
    results: [],
    description: '',
  }
})

// ── Загрузка теста для редактирования ──
onMounted(async () => {
  if (editMode) {
    // Режим редактирования – загружаем тест
    try {
      const test = await api.getTest(testId)
      const userId = authStore.state.user?.id
      if (test.author_id !== userId && !authStore.state.user?.isAdmin) {
        router.push(`/tests/${testId}`)
        return
      }
      prefillForm(test)
    } catch {
      router.push('/')
    } finally {
      pageLoading.value = false
    }
  } else {
    // Режим создания – проверяем импортированные данные
    if (importedStore.data.value) {
      const test = importedStore.data.value
      prefillForm({
        title: test.title,
        description: test.description,
        questions: test.questions,
        // возможно, теги и др.
      })
      importedStore.clear() // очищаем после использования
    }
    pageLoading.value = false
  }
})

function reverseMapType(t) {
  return { single_choice: 'single', multiple_choice: 'multiple', scale: 'scale', text: 'text' }[t] || 'single'
}

function prefillForm(test) {
  suppressWatch.value = true

  form.title       = test.title || ''
  form.description = test.description || ''
  form.tags        = [...(test.tags || [])]

  const qs = (typeof test.questions === 'string'
    ? JSON.parse(test.questions)
    : test.questions) || []

  form.questions = qs.map(q => ({
    id:       q.id || crypto.randomUUID(),
    text:     q.text || '',
    type:     reverseMapType(q.type),
    required: !!q.required,
    min:      q.min ?? 1,
    max:      q.max ?? 10,
    scoring:  q.scoring ? { ...q.scoring } : {},
    options:  (q.options || []).map(o => ({
      id:      o.id || crypto.randomUUID(),
      text:    o.text || '',
      scoring: o.scoring ? { ...o.scoring } : {},
    })),
  }))

  const cfg = (typeof test.result_config === 'string'
    ? JSON.parse(test.result_config)
    : test.result_config) || {}

  if (!cfg.type || cfg.type === 'none') {
    resultType.value = 'none'
    form.result_config = { min: 0, max: 100, levels: [], axes: [], results: [], description: '' }
  } else {
    resultType.value = cfg.type
    form.result_config = {
      type:        cfg.type,
      min:         cfg.min ?? 0,
      max:         cfg.max ?? 100,
      description: cfg.description || '',
      levels:      (cfg.levels || []).map(l => ({ ...l })),
      axes:        (cfg.axes || []).map(a => ({
        id:          a.id || '',
        label:       a.label || '',
        min:         a.min ?? -50,   // гарантируем диапазон
        max:         a.max ?? 50,
        left_label:  a.left_label || '',
        right_label: a.right_label || '',
      })),
      results: (cfg.results || []).map(r => ({
        key:         r.key || '',
        label:       r.label || '',
        description: r.description || '',
        target:      r.target ? { ...r.target } : {},
      })),
    }
  }

  nextTick(() => { suppressWatch.value = false })
}

// ── Теги ──
function addTag() {
  const raw = tagInput.value.replace(/,/g, '').trim()
  if (raw && !form.tags.includes(raw)) form.tags.push(raw)
  tagInput.value = ''
}
function removeTag(tag) { form.tags = form.tags.filter(t => t !== tag) }

// ── Вопросы ──
function addQuestion()      { form.questions.push(makeQuestion()) }
function removeQuestion(i)  { if (form.questions.length > 1) form.questions.splice(i, 1) }
function moveQuestion(i, d) {
  const j = i + d
  if (j < 0 || j >= form.questions.length) return
  ;[form.questions[i], form.questions[j]] = [form.questions[j], form.questions[i]]
}
function copyQuestion(i) {
  const src = form.questions[i]
  form.questions.splice(i + 1, 0, {
    id:       crypto.randomUUID(),
    text:     src.text,
    type:     src.type,
    required: src.required,
    min:      src.min,
    max:      src.max,
    scoring:  { ...src.scoring },
    options:  src.options.map(o => ({
      id:      crypto.randomUUID(),
      text:    o.text,
      scoring: { ...o.scoring },
    })),
  })
}
function onTypeChange(q) {
  if ((q.type === 'single' || q.type === 'multiple') && !q.options.length) {
    q.options = [
      { id: crypto.randomUUID(), text: '', scoring: {} },
      { id: crypto.randomUUID(), text: '', scoring: {} },
    ]
  }
}
function addOption(q)        { q.options.push({ id: crypto.randomUUID(), text: '', scoring: {} }) }
function removeOption(q, i)  { if (q.options.length > 2) q.options.splice(i, 1) }

// ── result_config helpers ──
function addLevel()     { form.result_config.levels.push({ min: 0, max: 0, label: '' }) }
function removeLevel(i) { form.result_config.levels.splice(i, 1) }
function addAxis() {
  // Явные дефолты -50/50 — иначе движок делит на (max-min) = 0
  form.result_config.axes.push({ id: '', label: '', min: -50, max: 50, left_label: '', right_label: '' })
}
function removeAxis(i)  { form.result_config.axes.splice(i, 1) }
function addResult() {
  form.result_config.results.push({ key: '', label: '', description: '', target: {} })
}
function removeResult(i) { form.result_config.results.splice(i, 1) }
function axisCenter(axis) {
  return String(Math.round(((axis.min ?? -50) + (axis.max ?? 50)) / 2))
}

// ── Сборка payload и сохранение ──
function mapType(t) {
  return { single: 'single_choice', multiple: 'multiple_choice', scale: 'scale', text: 'text' }[t] || t
}
function filterZeroScoring(scoring) {
  if (!scoring || typeof scoring !== 'object') return null
  const out = {}
  for (const [k, v] of Object.entries(scoring)) {
    if (v !== '' && v !== null && v !== undefined && !Number.isNaN(v)) out[k] = v
  }
  return Object.keys(out).length ? out : null
}

async function publish() {
  error.value = ''
  if (!form.title.trim())        { error.value = 'Введите название'; return }
  if (!form.questions.length)    { error.value = 'Добавьте хотя бы один вопрос'; return }
  for (const q of form.questions) {
    if (!q.text.trim())          { error.value = 'Заполните текст всех вопросов'; return }
    if ((q.type === 'single' || q.type === 'multiple') && q.options.length < 2) {
      error.value = 'Каждый вопрос с вариантами должен иметь минимум 2 варианта'; return
    }
  }

  const questions = form.questions.map(q => {
    const out = { id: q.id, text: q.text, type: mapType(q.type), required: q.required }
    if (q.type === 'single' || q.type === 'multiple') {
      out.options = q.options.map(o => {
        const opt = { id: o.id, text: o.text }
        const sc = filterZeroScoring(o.scoring)
        if (sc) opt.scoring = sc
        return opt
      })
    }
    if (q.type === 'scale') {
      out.min = q.min
      out.max = q.max
      const sc = filterZeroScoring(q.scoring)
      if (sc) out.scoring = sc
    }
    return out
  })

  const payload = {
    title:       form.title.trim(),
    description: form.description.trim(),
    questions,
    tags: form.tags,
  }

  if (resultType.value !== 'none') {
    const cfg = { type: resultType.value }
    if (resultType.value === 'score') {
      cfg.min = form.result_config.min
      cfg.max = form.result_config.max
      if (form.result_config.levels.length) cfg.levels = form.result_config.levels
    }
    if (['scales', 'string_map', 'combo'].includes(resultType.value)) {
      cfg.axes = form.result_config.axes
    }
    if (['string_map', 'combo'].includes(resultType.value)) {
      cfg.results = form.result_config.results
    }
    if (form.result_config.description?.trim()) {
      cfg.description = form.result_config.description.trim()
    }
    payload.result_config = cfg
  }

  publishing.value = true
  try {
    const test = editMode
      ? await api.updateTest(testId, payload)
      : await api.createTest(payload)
    router.push(`/tests/${test.id}`)
  } catch (e) {
    error.value = e.data?.error || e.message || (editMode ? 'Не удалось сохранить' : 'Не удалось создать тест')
  } finally {
    publishing.value = false
  }
}

async function doImport() {
  importError.value = ''
  if (!importUrl.value.trim()) {
    importError.value = 'Введите ссылку'
    return
  }
  importLoading.value = true
  try {
    const imported = await api.importGoogleForm(importUrl.value)
    // Заполняем форму импортированными данными
    prefillForm({
      title: imported.title,
      description: imported.description,
      questions: imported.questions,
      // при необходимости можно добавить tags, если они будут в ответе
    })
    // Очищаем стор, если там были данные (на случай, если пришли с ленты)
    importedStore.clear()
    showImport.value = false
    importUrl.value = ''
    // Можно добавить уведомление об успехе (опционально)
    // error.value = 'Форма успешно импортирована'  // но error используется для ошибок, лучше сделать отдельную переменную
  } catch (e) {
    importError.value = e.data?.error || e.message || 'Не удалось импортировать форму'
  } finally {
    importLoading.value = false
  }
}

</script>

<style scoped>
/* Стили полностью идентичны оригинальному CreateTestPage — вставьте блок <style> без изменений */
.back-link {
  display: inline-block; font-size: 0.78rem; letter-spacing: 0.06em;
  color: var(--text-muted); margin-bottom: 1.2rem; transition: color var(--transition);
}

.create-header-actions {
  display: flex;
  gap: 0.8rem;
  margin-bottom: 1rem;
}

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

.back-link:hover { color: var(--accent); }
.create-header { margin-bottom: 0; }
.create-title { font-family: var(--font-serif); font-size: 2rem; font-weight: 700; margin-bottom: 1.5rem; }
.section { margin-bottom: 0; }
.section-title { font-family: var(--font-serif); font-size: 1.3rem; font-weight: 600; margin-bottom: 1.2rem; }
.field { margin-bottom: 1.2rem; }
.field-label {
  display: block; font-family: var(--font-mono); font-size: 0.68rem;
  letter-spacing: 0.1em; text-transform: uppercase; color: var(--text-muted); margin-bottom: 0.4rem;
}
.field-hint { font-family: var(--font-mono); font-size: 0.65rem; color: var(--text-muted); margin-top: 0.3rem; }
.tags-input {
  border: 1px solid var(--border); border-radius: var(--radius); background: var(--bg-input);
  padding: 0.4rem 0.6rem; display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: center;
  transition: border-color var(--transition);
}
.tags-input:focus-within { border-color: var(--accent); }
.tags-input input { border: none; background: none; padding: 0.2rem 0.3rem; flex: 1; min-width: 120px; font-size: 0.9rem; }
.tags-input input:focus { border: none; }
.tag-chip { display: inline-flex; align-items: center; gap: 0.3rem; }
.chip-remove { font-size: 0.8rem; line-height: 1; color: var(--accent); opacity: 0.7; cursor: pointer; transition: opacity var(--transition); }
.chip-remove:hover { opacity: 1; }
.questions-list { display: flex; flex-direction: column; gap: 1rem; margin-bottom: 1rem; }
.question-card { border: 1px solid var(--border); border-radius: 6px; background: var(--bg-card); overflow: hidden; }
.question-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.6rem 1rem; border-bottom: 1px solid var(--border); background: rgba(255,255,255,0.02);
}
.question-num { font-family: var(--font-mono); font-size: 0.75rem; color: var(--accent); letter-spacing: 0.06em; }
.question-controls { display: flex; gap: 0.3rem; }
.ctrl-btn {
  font-family: var(--font-mono); font-size: 0.75rem; padding: 0.2rem 0.5rem;
  color: var(--text-muted); border: 1px solid var(--border); border-radius: var(--radius);
  transition: color var(--transition), border-color var(--transition);
}
.ctrl-btn:hover:not(:disabled) { color: var(--text); border-color: var(--text-muted); }
.ctrl-btn:disabled { opacity: 0.25; pointer-events: none; }
.ctrl-danger:hover:not(:disabled) { color: var(--danger); border-color: var(--danger); }
.question-body { padding: 1rem; display: flex; flex-direction: column; gap: 0.8rem; }
.question-meta { display: flex; gap: 1.5rem; align-items: flex-end; flex-wrap: wrap; }
.field-inline { display: flex; flex-direction: column; gap: 0.3rem; }
.field-inline select, .field-inline input { width: auto; }
.required-check { display: flex; align-items: center; gap: 0.4rem; cursor: pointer; padding-bottom: 0.65rem; }
.required-check input { width: auto; }
.required-check span { font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-muted); }
.options-list { display: flex; flex-direction: column; gap: 0.5rem; }
.option-row-wrap { display: flex; flex-direction: column; gap: 0.3rem; }
.option-row { display: flex; align-items: center; gap: 0.5rem; }
.option-bullet { font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-muted); min-width: 1.2rem; }
.scale-row { display: flex; gap: 1rem; flex-wrap: wrap; align-items: flex-end; }
.add-option-btn { font-size: 0.78rem; font-family: var(--font-mono); color: var(--text-muted); margin-top: 0.3rem; padding: 0; text-align: left; }
.add-option-btn:hover { color: var(--accent); }
.add-question-btn { margin-top: 0.5rem; }
.result-type-row { display: flex; gap: 1.5rem; flex-wrap: wrap; margin-bottom: 1.2rem; }
.result-type-option { display: flex; align-items: center; gap: 0.4rem; cursor: pointer; }
.result-type-option input { width: auto; }
.result-type-option span { font-family: var(--font-mono); font-size: 0.8rem; }
.result-config-block {
  padding: 1rem; border: 1px solid var(--border); border-radius: 6px;
  background: var(--bg-card); margin-top: 0.8rem; display: flex; flex-direction: column; gap: 0.6rem;
}
.levels-list { display: flex; flex-direction: column; gap: 0.5rem; }
.level-row { display: flex; gap: 0.5rem; align-items: center; }
.publish-row { padding: 0.5rem 0 1.5rem; }
.how-it-works {
  font-size: 0.82rem; color: var(--text-muted); line-height: 1.55;
  padding: 0.6rem 0.8rem; background: rgba(212,168,67,0.05);
  border-left: 2px solid var(--accent-dim); border-radius: 0 var(--radius) var(--radius) 0; margin-bottom: 0.8rem;
}
.how-it-works strong { color: var(--text); font-weight: 500; }
.how-it-works em { font-style: italic; color: var(--accent); }
.result-entry {
  display: flex; flex-direction: column; gap: 0.5rem; padding: 0.8rem;
  border: 1px solid var(--border); border-radius: var(--radius); margin-bottom: 0.6rem; background: rgba(255,255,255,0.01);
}
.result-entry-header { display: flex; gap: 0.5rem; align-items: center; }
.target-row {
  display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; padding: 0.4rem 0.5rem;
  background: rgba(78,184,122,0.04); border-left: 2px solid rgba(78,184,122,0.2); border-radius: 0 var(--radius) var(--radius) 0;
}
.axis-editor { display: flex; flex-direction: column; gap: 0.4rem; padding: 0.6rem; border: 1px solid var(--border); border-radius: var(--radius); margin-bottom: 0.5rem; }
.axis-editor-row { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.axis-range-row { padding-left: 0.2rem; opacity: 0.85; }
.scoring-row {
  display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap;
  padding: 0.4rem 0.5rem 0.4rem 1.7rem; background: rgba(212,168,67,0.04);
  border-left: 2px solid var(--accent-dim); border-radius: 0 var(--radius) var(--radius) 0;
}
.scoring-row-scale { padding-left: 0.5rem; margin-top: 0.3rem; }
.scoring-label { font-family: var(--font-mono); font-size: 0.68rem; letter-spacing: 0.06em; color: var(--text-muted); white-space: nowrap; }
.scoring-input { width: 80px !important; font-size: 0.85rem; padding: 0.3rem 0.5rem; }
.scoring-hint { font-family: var(--font-mono); font-size: 0.68rem; color: var(--text-muted); opacity: 0.6; }
.loading-state { display: flex; justify-content: center; padding: 3rem 0; }
.font-mono { font-family: var(--font-mono); }
</style>