<template>
  <div class="auth-wrap">
    <div class="auth-box">

      <!-- Логотип / заголовок -->
      <div class="auth-header">
        <span class="auth-logo">Public tests</span>
        <p class="auth-tagline">Платформа анонимных исследований</p>
      </div>

      <!-- Шаг 1: ввод контакта -->
      <template v-if="step === 'contact'">
        <div class="tab-row">
          <button :class="['tab', method === 'email' && 'active']" @click="method = 'email'">Email</button>
          <!-- <button :class="['tab', method === 'phone' && 'active']" @click="method = 'phone'">Телефон</button> -->
        </div>

        <div class="field">
          <label>{{ method === 'email' ? 'Электронная почта' : 'Номер телефона' }}</label>
          <input
            v-if="method === 'email'"
            v-model="contact"
            type="email"
            placeholder="you@example.com"
            @keydown.enter="sendCode"
          />
          <input
            v-else
            v-model="contact"
            type="tel"
            placeholder="+7 900 000 00 00"
            @keydown.enter="sendCode"
          />
        </div>

        <p v-if="error" class="error-msg">{{ error }}</p>

        <button class="btn" style="width:100%; justify-content:center" :disabled="loading" @click="sendCode">
          <span v-if="loading" class="spinner"></span>
          <span v-else>Получить код →</span>
        </button>

        <p class="auth-legal">
          Продолжая, вы соглашаетесь с
          <router-link to="/terms">условиями использования</router-link>
          и
          <router-link to="/privacy">политикой конфиденциальности</router-link>,
          а также даёте
          <router-link to="/consent">согласие на обработку персональных данных</router-link>.
        </p>
      </template>

      <!-- Шаг 2: ввод кода (и никнейма если новый пользователь) -->
      <template v-if="step === 'code'">
        <p class="step-hint">
          Код отправлен на <strong>{{ contact }}</strong>.
          <button class="btn-ghost" style="display:inline;padding:0;font-size:inherit;" @click="step = 'contact'">Изменить</button>
        </p>
        <p class="step-hint">
          Не забудьте проверить папку "Спам"
        </p>

        <div class="field">
          <label>Код подтверждения</label>
          <input
            v-model="code"
            type="text"
            inputmode="numeric"
            maxlength="6"
            placeholder="000000"
            class="code-input"
            @keydown.enter="verifyCode"
          />
        </div>

        <div v-if="needNickname" class="field">
          <label>Никнейм <span class="muted">(придумайте один раз)</span></label>
          <input v-model="nickname" type="text" placeholder="cool_username" maxlength="50" @keydown.enter="verifyCode" />
        </div>

        <p v-if="error" class="error-msg">{{ error }}</p>

        <button class="btn" style="width:100%; justify-content:center" :disabled="loading" @click="verifyCode">
          <span v-if="loading" class="spinner"></span>
          <span v-else>{{ needNickname ? 'Зарегистрироваться →' : 'Войти →' }}</span>
        </button>

        <button class="btn-ghost resend" @click="sendCode" :disabled="resendCooldown > 0">
          {{ resendCooldown > 0 ? `Повторить через ${resendCooldown}с` : 'Отправить код повторно' }}
        </button>
      </template>

    </div>

    <!-- Декоративный фон -->
    <div class="auth-bg" aria-hidden="true">
      <div class="auth-bg-line" v-for="i in 6" :key="i"></div>
    </div>
  </div>
</template>

<script setup>
import { onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { auth } from '../api/index.js'
import { authStore } from '../store/auth.js'

const router = useRouter()

const method      = ref('email')
const contact     = ref('')
const code        = ref('')
const nickname    = ref('')
const step        = ref('contact')
const needNickname = ref(false)
const loading     = ref(false)
const error       = ref('')
const resendCooldown = ref(0)

let cooldownTimer = null

function startCooldown() {
  resendCooldown.value = 15
  cooldownTimer = setInterval(() => {
    resendCooldown.value--
    if (resendCooldown.value <= 0) clearInterval(cooldownTimer)
  }, 1000)
}

onUnmounted(() => clearInterval(cooldownTimer))

async function sendCode() {
  error.value = ''
  if (!contact.value.trim()) {
    error.value = 'Введите ' + (method.value === 'email' ? 'email' : 'номер телефона')
    return
  }
  loading.value = true
  try {
    const body = method.value === 'email'
      ? { email: contact.value }
      : { phone: contact.value }
    await auth.sendCode(body)
    step.value = 'code'
    startCooldown()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function verifyCode() {
  error.value = ''
  if (code.value.length !== 6) {
    error.value = 'Введите 6-значный код'
    return
  }
  loading.value = true
  try {
    const body = method.value === 'email'
      ? { email: contact.value, code: code.value, nickname: nickname.value || undefined }
      : { phone: contact.value, code: code.value, nickname: nickname.value || undefined }

    const data = await auth.verify(body)

    if (data.new_user) {
      needNickname.value = true
      error.value = 'Придумайте никнейм для регистрации'
      loading.value = false
      return
    }

    authStore.login(data.access_token)
    router.push('/')
  } catch (e) {
    if (e.status === 409) {
      error.value = 'Этот никнейм уже занят'
    } else {
      error.value = e.message
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.auth-box {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 400px;
  padding: 2.5rem 2.5rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.auth-header { margin-bottom: 2rem; }
.auth-logo {
  font-family: var(--font-mono);
  font-size: 0.78rem;
  letter-spacing: 0.2em;
  color: var(--accent);
}
.auth-tagline {
  margin-top: 0.4rem;
  font-size: 1.25rem;
  font-family: var(--font-serif);
  color: var(--text);
  line-height: 1.3;
}

.tab-row {
  display: flex;
  gap: 0;
  margin-bottom: 1.5rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}
.tab {
  flex: 1;
  padding: 0.5rem;
  font-family: var(--font-mono);
  font-size: 0.78rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-muted);
  background: transparent;
  transition: background var(--transition), color var(--transition);
}
.tab.active { background: var(--accent-dim); color: var(--accent); }

.code-input {
  font-family: var(--font-mono);
  font-size: 1.6rem;
  letter-spacing: 0.3em;
  text-align: center;
}

.step-hint {
  font-size: 0.88rem;
  color: var(--text-muted);
  margin-bottom: 1.5rem;
}
.step-hint strong { color: var(--text); }

.resend {
  display: block;
  width: 100%;
  text-align: center;
  margin-top: 1rem;
  font-size: 0.82rem;
}

.auth-legal {
  margin-top: 1.2rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  line-height: 1.65;
  text-align: center;
  opacity: 0.75;
}
.auth-legal a {
  color: var(--accent);
  text-decoration: underline;
  text-decoration-color: var(--accent-dim);
  transition: opacity var(--transition);
}
.auth-legal a:hover { opacity: 0.75; }

/* Декоративный фон — горизонтальные линии */
.auth-bg {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-evenly;
  pointer-events: none;
}
.auth-bg-line {
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(212,168,67,0.06) 50%, transparent);
}
</style>
