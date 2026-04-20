<template>
  <router-link :to="`/tests/${test.id}`" class="card">
    <div class="card-top">
      <div class="card-badges">
        <span v-if="test.is_official" class="tag official">Официальный</span>
        <span class="tag">{{ questionCount }} вопр.</span>
      </div>
      <div class="card-vote">
        <span class="vote-score" :class="{ pos: test.rating > 0, neg: test.rating < 0 }">
          {{ test.rating > 0 ? '+' : '' }}{{ test.rating }}
        </span>
      </div>
    </div>

    <h3 class="card-title">{{ test.title }}</h3>
    <p v-if="test.description" class="card-desc">{{ truncate(test.description, 120) }}</p>

    <!-- Пользовательские теги -->
    <div v-if="test.tags?.length" class="card-tags">
      <span
        v-for="tag in test.tags.slice(0, 4)"
        :key="tag"
        class="tag tag-pill"
        @click.prevent="$emit('filter-tag', tag)"
      >{{ tag }}</span>
    </div>

    <div class="card-footer">
      <span class="card-meta">{{ test.pass_count }} прохождений</span>
      <span class="card-meta font-mono">💬 {{ test.comment_count || 0 }}</span>
      <span class="card-meta">{{ formatDate(test.created_at) }}</span>
    </div>
  </router-link>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  test: { type: Object, required: true },
})

defineEmits(['filter-tag'])

const questionCount = computed(() => {
  if (Array.isArray(props.test.questions)) return props.test.questions.length
  // questions может быть JSON-строкой из RawMessage
  if (typeof props.test.questions === 'string') {
    try { return JSON.parse(props.test.questions).length } catch { return 0 }
  }
  return 0
})

function truncate(str, len) {
  return str.length > len ? str.slice(0, len) + '…' : str
}

function formatDate(iso) {
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<style scoped>
.card {
  display: block;
  padding: 1.4rem 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  transition: border-color var(--transition), transform var(--transition);
  cursor: pointer;
}
.card:hover {
  border-color: rgba(212,168,67,0.3);
  transform: translateY(-2px);
}

.card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.8rem;
}
.card-badges { display: flex; gap: 0.4rem; flex-wrap: wrap; }

.vote-score {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-muted);
}
.vote-score.pos { color: var(--success); }
.vote-score.neg { color: var(--danger); }

.card-title {
  font-family: var(--font-serif);
  font-size: 1.1rem;
  font-weight: 600;
  line-height: 1.3;
  margin-bottom: 0.5rem;
  color: var(--text);
}

.card-desc {
  font-size: 0.88rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin-bottom: 0.8rem;
}

.card-tags {
  display: flex;
  gap: 0.3rem;
  flex-wrap: wrap;
  margin-bottom: 0.8rem;
}
.tag-pill {
  cursor: pointer;
  transition: opacity var(--transition);
}
.tag-pill:hover { opacity: 0.75; }

.card-footer {
  display: flex;
  gap: 1.2rem;
  flex-wrap: wrap;
}
.card-meta {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  color: var(--text-muted);
  letter-spacing: 0.03em;
}
</style>
