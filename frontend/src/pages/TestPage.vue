<template>
  <div class="page" v-if="test">

    <!-- Шапка теста -->
    <div class="test-header">
      <router-link to="/" class="back-link font-mono">← Лента</router-link>
      <div class="test-tags">
        <span v-if="test.is_official" class="tag official">Официальный</span>
        <span class="tag">{{ test.questions?.length || 0 }} вопросов</span>
        <span v-for="tag in test.tags" :key="tag" class="tag">{{ tag }}</span>
      </div>
      <h1 class="test-title">{{ test.title }}</h1>
      <p v-if="test.description" class="test-desc">{{ test.description }}</p>
      <div class="test-meta">
        <span class="font-mono">{{ test.pass_count }} прохождений</span>
        <div class="vote-row">
          <button :class="['vote-btn', voted === 1 && 'active-up']" @click="vote(1)">▲</button>
          <span class="vote-score font-mono" :class="{ pos: test.rating > 0, neg: test.rating < 0 }">
            {{ test.rating > 0 ? '+' : '' }}{{ test.rating }}
          </span>
          <button :class="['vote-btn down', voted === -1 && 'active-down']" @click="vote(-1)">▼</button>
        </div>
      </div>
    </div>

    <div class="divider"></div>

    <!-- Уже пройден -->
    <div v-if="alreadyDone" class="done-banner">
      <span>✓ Вы уже прошли этот тест</span>
      <button class="btn-ghost" @click="retake">Пройти заново</button>
    </div>

    <!-- Форма прохождения -->
    <form v-else @submit.prevent="submit">
      <QuestionRenderer
        v-for="(q, i) in test.questions"
        :key="q.id"
        :question="q"
        :index="i"
        v-model="answers[q.id]"
      />

      <div class="submit-row">
        <p v-if="submitError" class="error-msg">{{ submitError }}</p>
        <button type="submit" class="btn" :disabled="submitting">
          <span v-if="submitting" class="spinner"></span>
          <span v-else>Отправить ответы →</span>
        </button>
      </div>
    </form>

    <!-- Результат -->
    <div v-if="result" class="result-box">
      <p class="result-label font-mono">Результат</p>
      <p class="result-msg">Ответы сохранены.</p>
      <!-- TODO v2: ResultDisplay компонент для детального отображения result_config -->
    </div>

    <div class="divider"></div>

    <!-- Комментарии -->
    <div class="comments">
      <h2 class="comments-title font-serif">Комментарии</h2>

      <div class="comment-form">
        <textarea v-model="commentText" rows="2" placeholder="Написать комментарий..." maxlength="1000"></textarea>
        <button class="btn btn-outline" :disabled="!commentText.trim()" @click="addComment">Отправить</button>
      </div>

      <div v-if="commentsLoading" class="loading-state">
        <div class="spinner"></div>
      </div>

      <div v-else-if="!comments.length" class="empty-comments font-mono">
        Комментариев пока нет
      </div>

      <div v-else class="comment-list">
        <div v-for="c in comments" :key="c.id" class="comment">
          <div class="comment-header">
            <span class="comment-author font-mono">{{ c.nickname }}</span>
            <span class="comment-date font-mono">{{ formatDate(c.created_at) }}</span>
            <button
              v-if="c.user_id === currentUserId"
              class="btn-ghost comment-delete"
              @click="deleteComment(c.id)"
            >удалить</button>
          </div>
          <p class="comment-text">{{ c.content }}</p>
        </div>
      </div>
    </div>

  </div>

  <!-- Загрузка -->
  <div v-else-if="pageLoading" class="page loading-state" style="min-height:60vh">
    <div class="spinner" style="width:36px;height:36px"></div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import QuestionRenderer from '../components/QuestionRenderer.vue'
import { api } from '../api/index.js'
import { authStore } from '../store/auth.js'

const route  = useRoute()
const testId = Number(route.params.id)

const test        = ref(null)
const pageLoading = ref(true)
const answers     = ref({})
const submitting  = ref(false)
const submitError = ref('')
const result      = ref(null)
const alreadyDone = ref(false)
const prevAnswer  = ref(null)
const voted       = ref(0)

const comments        = ref([])
const commentsLoading = ref(false)
const commentText     = ref('')

const currentUserId = authStore.state.user?.id

async function loadTest() {
  try {
    test.value = await api.getTest(testId)
  } catch {
    test.value = null
  } finally {
    pageLoading.value = false
  }
}

async function loadComments() {
  commentsLoading.value = true
  try {
    const data = await api.getComments(testId)
    // бэкенд возвращает []domain.Comment напрямую
    comments.value = Array.isArray(data) ? data : (data.items || [])
  } catch { /* ignore */ }
  finally { commentsLoading.value = false }
}

function retake() {
  if (prevAnswer.value?.answers) {
    try {
      answers.value = typeof prevAnswer.value.answers === 'string'
        ? JSON.parse(prevAnswer.value.answers)
        : prevAnswer.value.answers
    } catch {
      answers.value = {}
    }
  }
  alreadyDone.value = false
}

async function submit() {
  submitError.value = ''

  const missing = test.value.questions
    .filter(q => q.required && !answers.value[q.id])
  if (missing.length) {
    submitError.value = `Ответьте на обязательные вопросы (${missing.map((_, i) => i + 1).join(', ')})`
    return
  }

  submitting.value = true
  try {
    const data = await api.submitAnswer(testId, { answers: answers.value })
    result.value = data
    alreadyDone.value = true
    // Инкрементируем счётчик только если это новый ответ
    if (!prevAnswer.value) test.value.pass_count++
    prevAnswer.value = data
  } catch (e) {
    submitError.value = e.data?.error || 'Не удалось сохранить ответы'
  } finally {
    submitting.value = false
  }
}

async function vote(val) {
  const newVote = voted.value === val ? 0 : val
  try {
    await api.voteTest(testId, newVote)
    test.value.rating += newVote - voted.value
    voted.value = newVote
  } catch { /* ignore */ }
}

async function addComment() {
  const text = commentText.value.trim()
  if (!text) return
  try {
    await api.addComment(testId, text)
    commentText.value = ''
    await loadComments()
  } catch { /* ignore */ }
}

async function deleteComment(id) {
  try {
    await api.deleteComment(testId, id)
    comments.value = comments.value.filter(c => c.id !== id)
  } catch { /* ignore */ }
}

function formatDate(iso) {
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

onMounted(async () => {
  await loadTest()

  // Загружаем предыдущий ответ
  const pa = await api.getMyAnswer(testId).catch(() => null)
  if (pa) {
    prevAnswer.value = pa
    alreadyDone.value = true
  }

  // Поддержка ?retake=1
  if (route.query.retake === '1' && prevAnswer.value) {
    retake()
  }

  loadComments()
})
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

.test-header { margin-bottom: 0; }
.test-tags { display: flex; gap: 0.5rem; margin-bottom: 0.8rem; flex-wrap: wrap; }
.test-title {
  font-family: var(--font-serif);
  font-size: 2rem;
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 0.7rem;
}
.test-desc { color: var(--text-muted); margin-bottom: 1rem; line-height: 1.6; }

.test-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}
.test-meta .font-mono { font-family: var(--font-mono); font-size: 0.8rem; color: var(--text-muted); }

.vote-row { display: flex; align-items: center; gap: 0.6rem; }
.vote-btn {
  font-size: 0.75rem;
  color: var(--text-muted);
  padding: 0.2rem 0.4rem;
  border-radius: var(--radius);
  transition: color var(--transition), background var(--transition);
}
.vote-btn:hover { color: var(--text); background: var(--bg-input); }
.vote-btn.active-up { color: var(--success); }
.vote-btn.active-down { color: var(--danger); }
.vote-score { font-family: var(--font-mono); font-size: 0.9rem; min-width: 2rem; text-align: center; }
.vote-score.pos { color: var(--success); }
.vote-score.neg { color: var(--danger); }

.done-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.2rem;
  background: rgba(78,184,122,0.08);
  border: 1px solid rgba(78,184,122,0.2);
  border-radius: 6px;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  color: var(--success);
}

.submit-row { padding: 1.5rem 0 0.5rem; }

.result-box {
  padding: 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  margin-top: 1.5rem;
}
.result-label {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
}
.result-msg { color: var(--text-muted); font-size: 0.9rem; }

/* Комментарии */
.comments-title {
  font-family: var(--font-serif);
  font-size: 1.3rem;
  margin-bottom: 1.2rem;
}
.comment-form {
  display: flex;
  gap: 0.8rem;
  align-items: flex-end;
  margin-bottom: 1.5rem;
}
.comment-form textarea { flex: 1; resize: none; }

.empty-comments {
  color: var(--text-muted);
  font-size: 0.8rem;
  letter-spacing: 0.04em;
  padding: 1rem 0;
}

.comment-list { display: flex; flex-direction: column; gap: 0.8rem; }
.comment {
  padding: 0.9rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
}
.comment-header {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  margin-bottom: 0.4rem;
}
.comment-author { font-family: var(--font-mono); font-size: 0.78rem; color: var(--accent); }
.comment-date   { font-family: var(--font-mono); font-size: 0.72rem; color: var(--text-muted); }
.comment-delete { font-size: 0.75rem; margin-left: auto; }
.comment-text   { font-size: 0.9rem; line-height: 1.5; }

.loading-state { display: flex; justify-content: center; padding: 3rem 0; }
.font-mono { font-family: var(--font-mono); }
.font-serif { font-family: var(--font-serif); }
</style>
