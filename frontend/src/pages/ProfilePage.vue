<template>
  <div class="page">

    <div class="profile-header">
      <div>
        <p class="profile-handle font-mono">{{ authStore.state.user?.nickname }}</p>
        <h1 class="profile-title">Мой профиль</h1>
      </div>
      <span v-if="authStore.state.user?.isAdmin" class="tag">admin</span>
    </div>

    <div class="divider"></div>

    <!-- Демографический профиль -->
    <section class="section">
      <div class="section-header">
        <div>
          <h2 class="section-title">Демографический профиль</h2>
          <p class="section-sub">Используется для анализа корреляций. Заполнение добровольно.</p>
        </div>
        <button v-if="!editingProfile" class="btn btn-outline" @click="startEdit">Редактировать</button>
        <div v-else class="edit-actions">
          <button class="btn" :disabled="profileSaving" @click="saveProfile">
            <span v-if="profileSaving" class="spinner"></span>
            <span v-else>Сохранить</span>
          </button>
          <button class="btn-ghost" @click="cancelEdit">Отмена</button>
        </div>
      </div>

      <!-- Просмотр -->
      <div v-if="!editingProfile" class="profile-grid">
        <div v-for="field in profileFields" :key="field.key" class="profile-field">
          <span class="field-label font-mono">{{ field.label }}</span>
          <span class="field-value">{{ displayValue(field) }}</span>
        </div>
      </div>

      <!-- Редактирование -->
      <div v-else class="edit-grid">
        <div class="edit-field">
          <label class="font-mono">Возраст</label>
          <input v-model.number="draft.age" type="number" min="1" max="149" placeholder="—" />
        </div>

        <div class="edit-field">
          <label class="font-mono">Пол</label>
          <select v-model="draft.gender">
            <option value="">—</option>
            <option value="M">Мужской</option>
            <option value="F">Женский</option>
            <option value="O">Другой</option>
          </select>
        </div>

        <div class="edit-field">
          <label class="font-mono">Доход ($/мес.)</label>
          <input v-model.number="draft.income" type="number" min="0" placeholder="—" />
        </div>

        <div class="edit-field">
          <label class="font-mono">Дети</label>
          <input v-model.number="draft.children" type="number" min="0" placeholder="—" />
        </div>

        <div class="edit-field">
          <label class="font-mono">Религия</label>
          <select v-model="draft.religion">
            <option value="">—</option>
            <option value="ch">Христианство</option>
            <option value="isl">Ислам</option>
            <option value="ind">Индуизм</option>
            <option value="dao">Даосизм</option>
            <option value="conf">Конфуцианство</option>
            <option value="jew">Иудаизм</option>
            <option value="ate">Атеизм</option>
            <option value="agn">Агностицизм</option>
            <option value="oth">Другое</option>
          </select>
        </div>

        <div class="edit-field">
          <label class="font-mono">Образование</label>
          <select v-model="draft.education">
            <option value="">—</option>
            <option value="secondary">Среднее</option>
            <option value="vocational">Среднее специальное</option>
            <option value="bachelor">Бакалавр</option>
            <option value="higher">Высшее</option>
            <option value="postgrad">Учёная степень</option>
          </select>
        </div>
      </div>

      <p v-if="profileError" class="error-msg">{{ profileError }}</p>
    </section>

    <div class="divider"></div>

    <!-- История ответов -->
    <section class="section">
      <h2 class="section-title">История ответов</h2>
      <p class="section-sub">Тесты, которые вы прошли</p>

      <div v-if="historyLoading" class="loading-state">
        <div class="spinner" style="width:28px;height:28px"></div>
      </div>

      <div v-else-if="!history.length" class="empty-state">
        <p class="font-mono">Вы ещё не прошли ни одного теста</p>
      </div>

      <div v-else class="history-list">
        <div
          v-for="item in history"
          :key="item.answer_id"
          class="history-item"
        >
          <!-- Клик по названию переходит на тест -->
          <router-link :to="`/tests/${item.test_id}`" class="history-main">
            <span class="history-title">{{ item.test_title }}</span>
            <span v-if="item.result" class="history-result font-mono">✓</span>
            <span
              v-else-if="item.score !== null && item.score !== undefined"
              class="history-score font-mono"
            >
              {{ item.score }} очков
            </span>
          </router-link>

          <span class="history-date font-mono">
            {{ formatDate(item.updated_at || item.created_at) }}
            <span
              v-if="item.updated_at && item.updated_at !== item.created_at"
              style="color:var(--text-muted)"
            > (изм.)</span>
          </span>

          <!-- Кнопка пересдать -->
          <router-link
            :to="`/tests/${item.test_id}?retake=1`"
            class="btn-ghost btn-retake font-mono"
          >Пересдать →</router-link>
        </div>
      </div>

      <!-- Пагинация истории -->
      <div v-if="historyTotal > historyLimit" class="pagination" style="margin-top:1.5rem">
        <button class="page-btn" :disabled="historyOffset === 0" @click="changeHistoryPage(-1)">← Назад</button>
        <span class="font-mono" style="font-size:0.8rem;color:var(--text-muted)">
          {{ Math.floor(historyOffset / historyLimit) + 1 }} / {{ Math.ceil(historyTotal / historyLimit) }}
        </span>
        <button class="page-btn" :disabled="historyOffset + historyLimit >= historyTotal" @click="changeHistoryPage(1)">Вперёд →</button>
      </div>
    </section>

  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api } from '../api/index.js'
import { authStore } from '../store/auth.js'

// --- Профиль ---
const profile        = ref({})
const editingProfile = ref(false)
const draft          = ref({})
const profileSaving  = ref(false)
const profileError   = ref('')

const profileFields = [
  { key: 'age',       label: 'Возраст',     format: v => v ?? '—' },
  { key: 'gender',    label: 'Пол',         format: v => ({ M: 'Мужской', F: 'Женский', O: 'Другой' }[v] || '—') },
  { key: 'income',    label: 'Доход',       format: v => v ? `$${v}` : '—' },
  { key: 'children',  label: 'Дети',        format: v => v ?? '—' },
  {
    key: 'religion',
    label: 'Религия',
    format: v => ({
      ch: 'Христианство', isl: 'Ислам', ind: 'Индуизм',
      dao: 'Даосизм', conf: 'Конфуцианство', jew: 'Иудаизм',
      ate: 'Атеизм', agn: 'Агностицизм', oth: 'Другое',
    }[v] || v || '—'),
  },
  {
    key: 'education',
    label: 'Образование',
    format: v => ({
      secondary: 'Среднее', vocational: 'Ср. специальное',
      bachelor: 'Бакалавр', higher: 'Высшее', postgrad: 'Учёная степень',
    }[v] || v || '—'),
  },
]

function displayValue(field) {
  return field.format(profile.value[field.key])
}

function startEdit() {
  draft.value = { ...profile.value }
  editingProfile.value = true
  profileError.value = ''
}
function cancelEdit() {
  editingProfile.value = false
}
async function saveProfile() {
  profileError.value = ''
  profileSaving.value = true
  try {
    const body = {}
    for (const [k, v] of Object.entries(draft.value)) {
      if (v !== '' && v !== null && v !== undefined) body[k] = v
    }
    profile.value = await api.updateProfile(body)
    editingProfile.value = false
  } catch (e) {
    profileError.value = e.data?.error || e.message
  } finally {
    profileSaving.value = false
  }
}

// --- История ---
const history        = ref([])
const historyTotal   = ref(0)
const historyLoading = ref(false)
const historyLimit   = 15
const historyOffset  = ref(0)

async function loadHistory() {
  historyLoading.value = true
  try {
    const data = await api.getAnswerHistory({ limit: historyLimit, offset: historyOffset.value })
    history.value = data.items || []
    historyTotal.value = data.total || 0
  } catch { /* ignore */ }
  finally { historyLoading.value = false }
}

function changeHistoryPage(dir) {
  historyOffset.value = Math.max(0, historyOffset.value + dir * historyLimit)
  loadHistory()
}

function formatDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

onMounted(async () => {
  try { profile.value = await api.getProfile() } catch { /* ignore */ }
  loadHistory()
})
</script>

<style scoped>
.profile-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 0;
}
.profile-handle {
  font-family: var(--font-mono);
  font-size: 0.78rem;
  color: var(--accent);
  letter-spacing: 0.1em;
  margin-bottom: 0.3rem;
}
.profile-title {
  font-family: var(--font-serif);
  font-size: 2rem;
  font-weight: 700;
}

.section { margin-bottom: 0; }
.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}
.section-title {
  font-family: var(--font-serif);
  font-size: 1.3rem;
  font-weight: 600;
  margin-bottom: 0.2rem;
}
.section-sub { font-size: 0.85rem; color: var(--text-muted); }
.edit-actions { display: flex; gap: 0.8rem; align-items: center; }

/* Профиль — просмотр */
.profile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 1rem;
}
.profile-field {
  padding: 0.9rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}

.field-value { font-size: 1rem; }

/* Профиль — редактирование */
.edit-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}
.edit-field label {
  display: block;
  font-family: var(--font-mono);
  font-size: 0.68rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.35rem;
}

/* История */
.history-list { display: flex; flex-direction: column; gap: 0.5rem; }
.history-item {
  display: flex;
  align-items: center;
  padding: 0.9rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  gap: 1rem;
  transition: border-color var(--transition);
}
.history-item:hover { border-color: rgba(212,168,67,0.3); }

.history-main {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
  min-width: 0;
  color: inherit;
}
.history-title { font-size: 0.92rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.history-result { font-family: var(--font-mono); font-size: 0.78rem; color: var(--success); white-space: nowrap; }
.history-score  { font-family: var(--font-mono); font-size: 0.78rem; color: var(--accent); white-space: nowrap; }
.history-date   { font-family: var(--font-mono); font-size: 0.72rem; color: var(--text-muted); white-space: nowrap; }

.font-mono { font-family: var(--font-mono); }
</style>
