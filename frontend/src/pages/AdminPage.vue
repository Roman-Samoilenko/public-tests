<template>
  <div class="page">

    <div class="admin-header">
      <h1 class="admin-title">Администрирование</h1>
      <p class="admin-sub">Управление тестами платформы</p>
    </div>

    <div class="divider"></div>

    <div v-if="loading" class="loading-state">
      <div class="spinner" style="width:32px;height:32px"></div>
    </div>

    <div v-else-if="!tests.length" class="empty-state">
      <p class="font-mono">Тестов нет</p>
    </div>

    <div v-else class="admin-table">
      <div class="table-header">
        <span class="col-title font-mono">Название</span>
        <span class="col-status font-mono">Статус</span>
        <span class="col-actions font-mono">Действия</span>
      </div>

      <div
        v-for="t in tests"
        :key="t.id"
        class="table-row"
        :class="{ blocked: t.status === 'blocked' }"
      >
        <div class="col-title">
          <router-link :to="`/tests/${t.id}`" class="test-link">{{ t.title }}</router-link>
          <div class="test-meta font-mono">
            <span>id: {{ t.id }}</span>
            <span>{{ t.pass_count }} прохождений</span>
            <span v-if="t.is_official" class="tag official" style="font-size:0.62rem">Официальный</span>
          </div>
        </div>

        <div class="col-status">
          <span
            class="status-badge font-mono"
            :class="t.status === 'blocked' ? 'status-blocked' : 'status-published'"
          >{{ t.status }}</span>
        </div>

        <div class="col-actions">
          <!-- Официальный -->
          <button
            class="action-btn"
            :class="t.is_official ? 'action-active' : ''"
            @click="toggleOfficial(t)"
            :disabled="actionLoading === t.id + '_official'"
          >
            <span v-if="actionLoading === t.id + '_official'" class="spinner" style="width:14px;height:14px"></span>
            <span v-else>{{ t.is_official ? '★ Снять официальный' : '☆ Пометить официальным' }}</span>
          </button>

          <!-- Статус -->
          <button
            class="action-btn"
            :class="t.status === 'blocked' ? 'action-unblock' : 'action-block'"
            @click="toggleStatus(t)"
            :disabled="actionLoading === t.id + '_status'"
          >
            <span v-if="actionLoading === t.id + '_status'" class="spinner" style="width:14px;height:14px"></span>
            <span v-else>{{ t.status === 'blocked' ? '✓ Разблокировать' : '✕ Заблокировать' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Пагинация -->
<div v-if="total > params.limit" class="pagination" style="margin-top:1.5rem">
  <button class="page-btn" :disabled="params.offset === 0" @click="changePage(-1)">← Назад</button>
  <span class="font-mono" style="font-size:0.8rem;color:var(--text-muted)">
    {{ Math.floor(params.offset / params.limit) + 1 }} / {{ Math.ceil(total / params.limit) }}
  </span>
  <button class="page-btn" :disabled="params.offset + params.limit >= total" @click="changePage(1)">Вперёд →</button>
</div>

    <p v-if="actionError" class="error-msg" style="margin-top:1rem">{{ actionError }}</p>

  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api } from '../api/index.js'
import { useTests } from '../composables/useTests.js'

// Используем композабл с параметрами админки
const { tests, total, loading, params, loadTests, changePage } = useTests({
  limit: 50,
  status: 'all', // загружаем все тесты, включая заблокированные
})

// Состояния для действий (оставляем локальными)
const actionLoading = ref('')
const actionError = ref('')

// Действия (без изменений, только используют api)
async function toggleOfficial(t) {
  actionError.value = ''
  actionLoading.value = t.id + '_official'
  try {
    await api.setOfficial(t.id, !t.is_official)
    t.is_official = !t.is_official
  } catch (e) {
    actionError.value = e.data?.error || e.message
  } finally {
    actionLoading.value = ''
  }
}

async function toggleStatus(t) {
  actionError.value = ''
  actionLoading.value = t.id + '_status'
  const newStatus = t.status === 'blocked' ? 'published' : 'blocked'
  try {
    await api.setStatus(t.id, newStatus)
    t.status = newStatus
  } catch (e) {
    actionError.value = e.data?.error || e.message
  } finally {
    actionLoading.value = ''
  }
}

onMounted(() => loadTests())
</script>

<style scoped>
.admin-header { margin-bottom: 0; }
.admin-title {
  font-family: var(--font-serif);
  font-size: 2rem;
  font-weight: 700;
}
.admin-sub { color: var(--text-muted); font-size: 0.88rem; margin-top: 0.3rem; }

.admin-table {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border);
  border-radius: 6px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 120px 280px;
  gap: 1rem;
  padding: 0.6rem 1rem;
  background: rgba(255,255,255,0.02);
  border-bottom: 1px solid var(--border);
}
.table-header span {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 120px 280px;
  gap: 1rem;
  padding: 0.9rem 1rem;
  align-items: center;
  border-bottom: 1px solid var(--border);
  transition: background var(--transition);
}
.table-row:last-child { border-bottom: none; }
.table-row:hover { background: rgba(255,255,255,0.02); }
.table-row.blocked { opacity: 0.6; }

.test-link {
  font-size: 0.92rem;
  color: var(--text);
  transition: color var(--transition);
}
.test-link:hover { color: var(--accent); }
.test-meta {
  display: flex;
  gap: 0.8rem;
  margin-top: 0.2rem;
  flex-wrap: wrap;
  align-items: center;
}
.test-meta span { font-family: var(--font-mono); font-size: 0.65rem; color: var(--text-muted); }

.status-badge {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  letter-spacing: 0.06em;
  padding: 0.2rem 0.6rem;
  border-radius: 2px;
}
.status-published { background: rgba(78,184,122,0.12); color: var(--success); }
.status-blocked   { background: rgba(224,90,78,0.12);  color: var(--danger); }

.col-actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-family: var(--font-mono);
  font-size: 0.7rem;
  letter-spacing: 0.04em;
  padding: 0.3rem 0.7rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-muted);
  transition: color var(--transition), border-color var(--transition);
}
.action-btn:hover:not(:disabled) { color: var(--text); border-color: var(--text-muted); }
.action-btn:disabled { opacity: 0.4; pointer-events: none; }
.action-active { color: var(--accent); border-color: var(--accent); }
.action-block:hover:not(:disabled)   { color: var(--danger); border-color: var(--danger); }
.action-unblock:hover:not(:disabled) { color: var(--success); border-color: var(--success); }


.font-mono { font-family: var(--font-mono); }
</style>
