<template>
  <div class="page">

    <div class="create-header">
      <router-link to="/" class="back-link font-mono">← Лента</router-link>
      <h1 class="create-title">Создать тест</h1>
    </div>

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
            <span
              v-for="tag in form.tags"
              :key="tag"
              class="tag tag-chip"
            >{{ tag }}<button class="chip-remove" @click="removeTag(tag)">×</button></span>
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
        <div
          v-for="(q, i) in form.questions"
          :key="q.id"
          class="question-card"
        >
          <div class="question-header">
            <span class="question-num font-mono">{{ i + 1 }}</span>
            <div class="question-controls">
              <button class="ctrl-btn" :disabled="i === 0" @click="moveQuestion(i, -1)" title="Вверх">↑</button>
              <button class="ctrl-btn" :disabled="i === form.questions.length - 1" @click="moveQuestion(i, 1)" title="Вниз">↓</button>
              <button
                class="ctrl-btn ctrl-danger"
                :disabled="form.questions.length === 1"
                @click="removeQuestion(i)"
                title="Удалить"
              >✕</button>
            </div>
          </div>

          <div class="question-body">
            <textarea
              v-model="q.text"
              rows="2"
              placeholder="Текст вопроса *"
              style="resize:vertical"
            ></textarea>

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

            <!-- Варианты ответов (single / multiple) -->
            <div v-if="q.type === 'single' || q.type === 'multiple'" class="options-list">
              <div
                v-for="(opt, oi) in q.options"
                :key="opt.id"
                class="option-row"
              >
                <span class="option-bullet font-mono">{{ oi + 1 }}.</span>
                <input v-model="opt.text" type="text" placeholder="Вариант ответа" />
                <button
                  class="ctrl-btn ctrl-danger"
                  :disabled="q.options.length <= 2"
                  @click="removeOption(q, oi)"
                >×</button>
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
          </div>
        </div>
      </div>

      <button class="btn btn-outline add-question-btn" @click="addQuestion">+ Добавить вопрос</button>
    </section>

    <div class="divider"></div>

    <!-- ─── Секция 3: Конфигурация результата ─── -->
    <section class="section">
      <h2 class="section-title">Конфигурация результата</h2>
      <p class="section-sub">v1 — только структура, вычисление появится в v2</p>

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

      <!-- Шкалы -->
      <div v-if="resultType === 'scales' || resultType === 'combo'" class="result-config-block">
        <p class="field-label font-mono" style="margin-bottom:0.5rem">Оси</p>
        <div v-for="(axis, ai) in form.result_config.axes" :key="ai" class="option-row">
          <input v-model="axis.id"    type="text" placeholder="id (латиница)" style="width:150px" />
          <input v-model="axis.label" type="text" placeholder="Название оси" />
          <button class="ctrl-btn ctrl-danger" @click="removeAxis(ai)">×</button>
        </div>
        <button class="btn-ghost add-option-btn" @click="addAxis">+ ось</button>
      </div>

      <!-- Тип / Комбо — результаты -->
      <div v-if="resultType === 'string_map' || resultType === 'combo'" class="result-config-block">
        <p class="field-label font-mono" style="margin-bottom:0.5rem">Результаты</p>
        <div v-for="(res, ri) in form.result_config.results" :key="ri" class="result-row">
          <input v-model="res.key"         type="text" placeholder="key" style="width:120px" />
          <input v-model="res.label"       type="text" placeholder="Название" style="width:160px" />
          <input v-model="res.description" type="text" placeholder="Описание" style="flex:1" />
          <button class="ctrl-btn ctrl-danger" @click="removeResult(ri)">×</button>
        </div>
        <button class="btn-ghost add-option-btn" @click="addResult">+ результат</button>
      </div>

      <!-- TODO v2: scoring (веса ответов) будет добавлено в v2 -->
    </section>

    <div class="divider"></div>

    <!-- Публикация -->
    <div class="publish-row">
      <button class="btn" :disabled="publishing" @click="publish">
        <span v-if="publishing" class="spinner"></span>
        <span v-else>Опубликовать тест →</span>
      </button>
    </div>

  </div>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api/index.js'

const router = useRouter()

const error     = ref('')
const publishing = ref(false)
const tagInput  = ref('')

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
      { id: crypto.randomUUID(), text: '' },
      { id: crypto.randomUUID(), text: '' },
    ],
    min: 1,
    max: 10,
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
  },
})

// Сбрасываем result_config при смене типа
watch(resultType, (val) => {
  form.result_config = {
    type: val === 'none' ? undefined : val,
    min: 0, max: 100,
    levels: [],
    axes: [],
    results: [],
  }
})

// ── Теги ──
function addTag() {
  const raw = tagInput.value.replace(/,/g, '').trim()
  if (raw && !form.tags.includes(raw)) form.tags.push(raw)
  tagInput.value = ''
}
function removeTag(tag) {
  form.tags = form.tags.filter(t => t !== tag)
}

// ── Вопросы ──
function addQuestion() {
  form.questions.push(makeQuestion())
}
function removeQuestion(i) {
  if (form.questions.length > 1) form.questions.splice(i, 1)
}
function moveQuestion(i, dir) {
  const j = i + dir
  if (j < 0 || j >= form.questions.length) return
  const tmp = form.questions[i]
  form.questions[i] = form.questions[j]
  form.questions[j] = tmp
}
function onTypeChange(q) {
  if (q.type === 'single' || q.type === 'multiple') {
    if (!q.options.length) {
      q.options = [
        { id: crypto.randomUUID(), text: '' },
        { id: crypto.randomUUID(), text: '' },
      ]
    }
  }
}
function addOption(q) {
  q.options.push({ id: crypto.randomUUID(), text: '' })
}
function removeOption(q, i) {
  if (q.options.length > 2) q.options.splice(i, 1)
}

// ── result_config helpers ──
function addLevel()  { form.result_config.levels.push({ min: 0, max: 0, label: '' }) }
function removeLevel(i) { form.result_config.levels.splice(i, 1) }
function addAxis()   { form.result_config.axes.push({ id: '', label: '' }) }
function removeAxis(i) { form.result_config.axes.splice(i, 1) }
function addResult() { form.result_config.results.push({ key: '', label: '', description: '' }) }
function removeResult(i) { form.result_config.results.splice(i, 1) }

// ── Публикация ──
async function publish() {
  error.value = ''

  if (!form.title.trim()) { error.value = 'Введите название'; return }
  if (!form.questions.length) { error.value = 'Добавьте хотя бы один вопрос'; return }
  for (const q of form.questions) {
    if (!q.text.trim()) { error.value = 'Заполните текст всех вопросов'; return }
    if ((q.type === 'single' || q.type === 'multiple') && q.options.length < 2) {
      error.value = 'Каждый вопрос с вариантами должен иметь минимум 2 варианта'; return
    }
  }

  // Преобразуем вопросы в формат бэкенда
  const questions = form.questions.map(q => {
    const out = { id: q.id, text: q.text, type: mapType(q.type), required: q.required }
    if (q.type === 'single' || q.type === 'multiple') {
      out.options = q.options.map(o => ({ id: o.id, text: o.text }))
    }
    if (q.type === 'scale') {
      out.min = q.min
      out.max = q.max
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
    if (resultType.value === 'scales' || resultType.value === 'combo') {
      cfg.axes = form.result_config.axes
    }
    if (resultType.value === 'string_map' || resultType.value === 'combo') {
      cfg.results = form.result_config.results
    }
    payload.result_config = cfg
  }

  publishing.value = true
  try {
    const test = await api.createTest(payload)
    router.push(`/tests/${test.id}`)
  } catch (e) {
    error.value = e.data?.error || e.message || 'Не удалось создать тест'
  } finally {
    publishing.value = false
  }
}

function mapType(t) {
  return { single: 'single_choice', multiple: 'multiple_choice', scale: 'scale', text: 'text' }[t] || t
}
</script>

<style scoped>
.back-link {
  display: inline-block;
  font-size: 0.78rem;
  letter-spacing: 0.06em;
  color: var(--text-muted);
  margin-bottom: 1.2rem;
  transition: color var(--transition);
}
.back-link:hover { color: var(--accent); }

.create-header { margin-bottom: 0; }
.create-title {
  font-family: var(--font-serif);
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
}

.section { margin-bottom: 0; }
.section-title {
  font-family: var(--font-serif);
  font-size: 1.3rem;
  font-weight: 600;
  margin-bottom: 1.2rem;
}
.section-sub { font-size: 0.85rem; color: var(--text-muted); margin-top: -0.8rem; margin-bottom: 1rem; }

.field { margin-bottom: 1.2rem; }
.field-label {
  display: block;
  font-family: var(--font-mono);
  font-size: 0.68rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.4rem;
}
.field-hint {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  color: var(--text-muted);
  margin-top: 0.3rem;
}

/* Теги */
.tags-input {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-input);
  padding: 0.4rem 0.6rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
  transition: border-color var(--transition);
}
.tags-input:focus-within { border-color: var(--accent); }
.tags-input input {
  border: none;
  background: none;
  padding: 0.2rem 0.3rem;
  flex: 1;
  min-width: 120px;
  font-size: 0.9rem;
}
.tags-input input:focus { border: none; }
.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}
.chip-remove {
  font-size: 0.8rem;
  line-height: 1;
  color: var(--accent);
  opacity: 0.7;
  cursor: pointer;
  transition: opacity var(--transition);
}
.chip-remove:hover { opacity: 1; }

/* Вопросы */
.questions-list { display: flex; flex-direction: column; gap: 1rem; margin-bottom: 1rem; }
.question-card {
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg-card);
  overflow: hidden;
}
.question-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1rem;
  border-bottom: 1px solid var(--border);
  background: rgba(255,255,255,0.02);
}
.question-num {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--accent);
  letter-spacing: 0.06em;
}
.question-controls { display: flex; gap: 0.3rem; }
.ctrl-btn {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  color: var(--text-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  transition: color var(--transition), border-color var(--transition);
}
.ctrl-btn:hover:not(:disabled) { color: var(--text); border-color: var(--text-muted); }
.ctrl-btn:disabled { opacity: 0.25; pointer-events: none; }
.ctrl-danger:hover:not(:disabled) { color: var(--danger); border-color: var(--danger); }

.question-body { padding: 1rem; display: flex; flex-direction: column; gap: 0.8rem; }

.question-meta {
  display: flex;
  gap: 1.5rem;
  align-items: flex-end;
  flex-wrap: wrap;
}
.field-inline { display: flex; flex-direction: column; gap: 0.3rem; }
.field-inline select, .field-inline input { width: auto; }

.required-check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  cursor: pointer;
  padding-bottom: 0.65rem;
}
.required-check input { width: auto; }
.required-check span { font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-muted); }

.options-list { display: flex; flex-direction: column; gap: 0.5rem; }
.option-row { display: flex; align-items: center; gap: 0.5rem; }
.option-bullet { font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-muted); min-width: 1.2rem; }

.scale-row { display: flex; gap: 1rem; flex-wrap: wrap; align-items: flex-end; }

.add-option-btn {
  font-size: 0.78rem;
  font-family: var(--font-mono);
  color: var(--text-muted);
  margin-top: 0.3rem;
  padding: 0;
  text-align: left;
}
.add-option-btn:hover { color: var(--accent); }

.add-question-btn { margin-top: 0.5rem; }

/* Конфигурация результата */
.result-type-row {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
  margin-bottom: 1.2rem;
}
.result-type-option {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  cursor: pointer;
}
.result-type-option input { width: auto; }
.result-type-option span { font-family: var(--font-mono); font-size: 0.8rem; }

.result-config-block {
  padding: 1rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg-card);
  margin-top: 0.8rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.levels-list { display: flex; flex-direction: column; gap: 0.5rem; }
.level-row { display: flex; gap: 0.5rem; align-items: center; }
.result-row { display: flex; gap: 0.5rem; align-items: center; }

/* Публикация */
.publish-row { padding: 0.5rem 0 1.5rem; }

.font-mono { font-family: var(--font-mono); }
</style>
