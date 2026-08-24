<template>
  <div id="shell">
    <nav v-if="authStore.state.user" class="nav">
      <router-link to="/" class="nav-brand">PublicTests</router-link>
      <div class="nav-links">
        <router-link to="/">Лента</router-link>
        <router-link to="/create">Создать</router-link>
        <router-link to="/profile">Профиль</router-link>
        <router-link v-if="authStore.state.user?.isAdmin" to="/admin">Админ</router-link>
        <button class="btn-ghost" @click="handleLogout">Выйти</button>
      </div>
    </nav>
    <main>
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" :key="$route.path" />
        </transition>
      </router-view>
    </main>
    <SiteFooter />
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { auth } from './api/index.js'
import SiteFooter from './components/SiteFooter.vue'
import { authStore } from './store/auth.js'
import { configStore } from './store/configStore'

const router = useRouter()

let refreshTimer = null

async function handleLogout() {
  await auth.logout().catch(() => {})
  authStore.logout()
  router.push('/auth')
}

onMounted(() => {
  // Запускаем таймер обновления токенов, если пользователь вошёл
  if (authStore.isLoggedIn.value) {
    refreshTimer = setInterval(async () => {
      await authStore.refreshSession()
    }, (configStore.accessTokenTTL / 2) * 1000)
  }
})

onUnmounted(() => {
  clearInterval(refreshTimer)
})
</script>

<style>
/* ── Дизайн-токены ───────────────────────────────────── */
:root {
  --bg:        #0c0c0e;
  --bg-card:   #13131a;
  --bg-input:  #1a1a24;
  --border:    rgba(255,255,255,0.08);
  --text:      #e8e6df;
  --text-muted:#7a7870;
  --accent:    #d4a843;
  --accent-dim:rgba(212,168,67,0.12);
  --danger:    #e05a4e;
  --success:   #4eb87a;

  --font-serif: 'Playfair Display', Georgia, serif;
  --font-mono:  'IBM Plex Mono', 'Courier New', monospace;
  --font-sans:  'IBM Plex Sans', sans-serif;

  --radius: 4px;
  --transition: 160ms ease;
}

/* ── Сброс и базовые стили ───────────────────────────── */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

html { font-size: 16px; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-sans);
  font-weight: 300;
  line-height: 1.6;
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
}

a {
  color: inherit;
  text-decoration: none;
}

button {
  font-family: var(--font-sans);
  cursor: pointer;
  border: none;
  background: none;
}

input, textarea, select {
  font-family: var(--font-sans);
  font-size: 0.95rem;
  background: var(--bg-input);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.65rem 0.9rem;
  width: 100%;
  outline: none;
  transition: border-color var(--transition);
}
input:focus, textarea:focus, select:focus {
  border-color: var(--accent);
}
input::placeholder, textarea::placeholder {
  color: var(--text-muted);
}
select option { background: var(--bg-card); }

/* ── Утилитарные компоненты ─────────────────────────── */



.spinner {
  width: 20px; height: 20px;
  border: 2px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  flex-shrink: 0;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Навигация ───────────────────────────────────────── */
.nav {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  height: 56px;
  background: rgba(12,12,14,0.92);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
}

.nav-brand {
  font-family: var(--font-mono);
  font-size: 0.82rem;
  font-weight: 500;
  letter-spacing: 0.18em;
  color: var(--accent);
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 1.8rem;
}
.nav-links a {
  font-size: 0.88rem;
  color: var(--text-muted);
  transition: color var(--transition);
  letter-spacing: 0.02em;
}
.nav-links a.router-link-active,
.nav-links a:hover { color: var(--text); }

/* ── Layout ──────────────────────────────────────────── */
main { min-height: calc(100vh - 56px); }

.page {
  max-width: 900px;
  margin: 0 auto;
  padding: 3rem 2rem;
}
.page-wide {
  max-width: 1100px;
  margin: 0 auto;
  padding: 3rem 2rem;
}

/* ── Анимация переходов ──────────────────────────────── */
.fade-enter-active, .fade-leave-active { transition: opacity 180ms ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* ── Скроллбар ───────────────────────────────────────── */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: var(--bg); }
::-webkit-scrollbar-thumb { background: #2a2a35; border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: #3a3a48; }
</style>

<style scoped>
#shell { display: flex; flex-direction: column; min-height: 100vh; }
main   { flex: 1; }
</style>
