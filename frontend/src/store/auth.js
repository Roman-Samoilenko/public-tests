// auth.js
import { computed, reactive } from 'vue'
import { clearTokens, getAccessToken, setTokens } from '../api/index.js'

function parseJWT(token) {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    // Декодируем base64 в бинарную строку (Latin-1)
    const binaryString = atob(base64);
    // Превращаем её в массив байт (Uint8Array)
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }
    // Декодируем байты как UTF-8
    const utf8String = new TextDecoder('utf-8').decode(bytes);
    return JSON.parse(utf8String);
  } catch {
    return null;
  }
}

const state = reactive({ user: null, initializing: true })

function applyToken(accessToken) {
  const claims = parseJWT(accessToken)
  if (claims) {
    state.user = { id: claims.user_id, nickname: claims.nickname, isAdmin: claims.is_admin }
    return true
  }
  return false
}

// Восстановление при загрузке страницы
async function init() {
  const accessToken = getAccessToken()
  if (accessToken) {
    const claims = parseJWT(accessToken)
    if (claims && claims.exp * 1000 > Date.now()) {
      // Токен ещё живой
      applyToken(accessToken)
    } else {
      // Токен просрочен — пробуем рефреш через cookie (не чистим сразу!)
      const refreshed = await authStore.refreshSession()
      if (!refreshed) clearTokens()
    }
  }
  state.initializing = false
}

export const authStore = {
  state,
  isLoggedIn: computed(() => !!state.user),
  isAdmin: computed(() => state.user?.isAdmin ?? false),

  login(accessToken) {
    setTokens(accessToken, null)  // refresh хранится в httpOnly cookie, не в localStorage
    applyToken(accessToken)
  },

  logout() {
    clearTokens()
    state.user = null
  },

  async refreshSession() {
    try {
      // Используем fetch с credentials: 'include' — браузер сам пошлёт httpOnly cookie
      const res = await fetch('/api/auth/refresh', {
        method: 'POST',
        credentials: 'include',
      })
      if (!res.ok) {
        this.logout()
        return false
      }
      const data = await res.json()
      setTokens(data.access_token, null)
      applyToken(data.access_token)
      return true
    } catch {
      this.logout()
      return false
    }
  }
}

init()