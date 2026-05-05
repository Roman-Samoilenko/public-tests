<template>
  <footer class="site-footer">
    <div class="footer-inner">

      <div class="footer-disclaimer">
        <span class="disclaimer-icon font-mono">!</span>
        <p>
          Администрация сайта не несёт ответственности за контент, размещённый пользователями.
          Вся ответственность за публикуемые материалы лежит на авторах.
        </p>
      </div>

      <div class="footer-nav">
        <router-link class="footer-link font-mono" to="/terms">Пользовательское соглашение</router-link>
        <span class="footer-sep font-mono">·</span>
        <router-link class="footer-link font-mono" to="/privacy">Политика конфиденциальности</router-link>
        <span class="footer-sep font-mono">·</span>
        <router-link class="footer-link font-mono" to="/consent">Согласие на обработку данных</router-link>
      </div>

      <div class="footer-copy font-mono">
        © {{ year }} Public tests — Платформа анонимных исследований
      </div>

    </div>
  </footer>
</template>

<script setup>
const year = new Date().getFullYear()
</script>

<style scoped>
.site-footer {
  margin-top: auto;
  border-top: 1px solid var(--border);
  background: var(--bg-card);
}
.footer-inner {
  max-width: 900px;
  margin: 0 auto;
  padding: 1.5rem 1.5rem 1.8rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.footer-disclaimer {
  display: flex;
  gap: 0.8rem;
  align-items: flex-start;
  padding: 0.8rem 1rem;
  background: rgba(212,168,67,0.04);
  border-left: 2px solid var(--accent-dim);
  border-radius: 0 var(--radius) var(--radius) 0;
}
.disclaimer-icon {
  font-family: var(--font-mono); font-size: 0.68rem; color: var(--accent);
  border: 1px solid var(--accent-dim); border-radius: 50%;
  width: 18px; height: 18px; min-width: 18px;
  display: flex; align-items: center; justify-content: center;
  margin-top: 1px; line-height: 1;
}
.footer-disclaimer p {
  font-size: 0.8rem; color: var(--text-muted); line-height: 1.6; margin: 0;
}

.footer-nav {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}
.footer-link {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  letter-spacing: 0.04em;
  color: var(--text-muted);
  text-decoration: none;
  transition: color var(--transition);
}
.footer-link:hover { color: var(--accent); }
.footer-sep {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  color: var(--border);
}

.footer-copy {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  letter-spacing: 0.06em;
  color: var(--text-muted);
  opacity: 0.5;
}
</style>
