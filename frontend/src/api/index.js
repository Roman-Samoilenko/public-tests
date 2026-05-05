const AUTH_BASE = '/api/auth'
const API_BASE = '/api'

const ACCESS_TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'

// ================== Токены ==================
export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setTokens(accessToken, refreshToken) {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  if (refreshToken !== undefined && refreshToken !== null) {
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  }
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

// ================== Выход (единая точка) ==================
async function fullLogout() {
  try {
    await fetch(`${AUTH_BASE}/logout`, {
      method: 'POST',
      credentials: 'include',
    })
  } catch { }
  clearTokens()
  window.location.href = '/auth'
}

// ================== Контроль рефреша ==================
let refreshPromise = null

async function tryRefresh() {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const res = await fetch(`${AUTH_BASE}/refresh`, {
          method: 'POST',
          credentials: 'include',
        })
        if (!res.ok) return false
        const data = await res.json()
        setTokens(data.access_token, getRefreshToken()) // сохраняем новый access и старый refresh
        return true
      } catch {
        return false
      } finally {
        refreshPromise = null
      }
    })()
  }
  return refreshPromise
}

// ================== Базовый запрос ==================
async function request(base, path, options = {}, retry = true) {
  const headers = { ...options.headers }

  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  const token = getAccessToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  let res = await fetch(`${base}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  })

  // Обработка 401 (только для API, не для auth)
  if (res.status === 401 && base === API_BASE && retry) {
    if (!getAccessToken()) {
      await fullLogout()
      return res
    }

    const refreshed = await tryRefresh()
    if (refreshed) {
      // повторяем запрос с новым токеном
      const newHeaders = { ...options.headers }
      if (!(options.body instanceof FormData)) {
        newHeaders['Content-Type'] = 'application/json'
      }
      newHeaders['Authorization'] = `Bearer ${getAccessToken()}`
      return fetch(`${base}${path}`, {
        ...options,
        headers: newHeaders,
        credentials: 'include',
      })
    }

    await fullLogout()
    return res
  }

  return res
}

// ================== JSON-хелпер ==================
async function json(base, path, options = {}) {
  const res = await request(base, path, options)

  if (!res.ok) {
    let errBody
    try {
      errBody = await res.json()
    } catch {
      errBody = { error: res.statusText }
    }
    throw Object.assign(
      new Error(errBody.error || 'request failed'),
      { status: res.status, data: errBody }
    )
  }

  return res.json()
}

// ================== Auth-сервис ==================
export const auth = {
  sendCode: (body) =>
    json(AUTH_BASE, '/send-code', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  verify: (body) =>
    json(AUTH_BASE, '/verify', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  refresh: () =>
    json(AUTH_BASE, '/refresh', {
      method: 'POST',
    }),

  logout: async () => {
    await fullLogout()
  },
}

// ================== API-сервис ==================
export const api = {
  getProfile: () => json(API_BASE, '/profile'),

  updateProfile: (body) =>
    json(API_BASE, '/profile', {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  getAnswerHistory: (p) =>
    json(API_BASE, `/profile/answers?limit=${p?.limit || 20}&offset=${p?.offset || 0}`),

  getTests: (params = {}) => {
    const q = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) {
      if (v === undefined || v === null || v === '') continue
      if (Array.isArray(v)) {
        if (v.length) q.set(k, v.join(','))
      } else {
        q.set(k, v)
      }
    }
    return json(API_BASE, '/tests?' + q.toString())
  },

  getTest: (id) => json(API_BASE, `/tests/${id}`),

  createTest: (body) =>
    json(API_BASE, '/tests', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  submitAnswer: (testId, body) =>
    json(API_BASE, `/tests/${testId}/answers`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  voteTest: (testId, vote) =>
    json(API_BASE, `/tests/${testId}/vote`, {
      method: 'POST',
      body: JSON.stringify({ vote }),
    }),

  getComments: (testId) =>
    json(API_BASE, `/tests/${testId}/comments`),

  addComment: (testId, text) =>
    json(API_BASE, `/tests/${testId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content: text }),
    }),

  deleteComment: (testId, id) =>
    json(API_BASE, `/tests/${testId}/comments/${id}`, {
      method: 'DELETE',
    }),

  updateTest: (id, body) =>
    json(API_BASE, `/tests/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),

  getMyAnswer: (testId) =>
    json(API_BASE, `/tests/${testId}/my-answer`),

  importGoogleForm: (url) =>
    json(API_BASE, '/import/google-forms', {
      method: 'POST',
      body: JSON.stringify({ url }),
    }),

  setOfficial: (testId, official) =>
    json(API_BASE, `/admin/tests/${testId}/official`, {
      method: 'PATCH',
      body: JSON.stringify({ official }),
    }),

  setStatus: (testId, status) =>
    json(API_BASE, `/admin/tests/${testId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),
}