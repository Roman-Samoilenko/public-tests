import { reactive, computed } from 'vue'
import { getToken, setToken, clearToken } from '../api/index.js'

// Парсим user из JWT payload (без верификации — только для UI).
function parseJWT(token) {
  try {
    const payload = token.split('.')[1]
    return JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
  } catch {
    return null
  }
}

const state = reactive({
  user: null,
})

// При загрузке страницы восстанавливаем пользователя из токена
const token = getToken()
if (token) {
  const claims = parseJWT(token)
  if (claims && claims.exp * 1000 > Date.now()) {
    state.user = { id: claims.user_id, nickname: claims.nickname, isAdmin: claims.is_admin }
  } else {
    clearToken()
  }
}

export const authStore = {
  state,

  isLoggedIn: computed(() => !!state.user),
  isAdmin:    computed(() => state.user?.isAdmin ?? false),

  login(accessToken, user) {
    setToken(accessToken)
    state.user = {
      id:       user.id,
      nickname: user.nickname,
      isAdmin:  user.is_admin,
    }
  },

  logout() {
    clearToken()
    state.user = null
  },
}
