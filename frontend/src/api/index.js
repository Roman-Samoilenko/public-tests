// api/index.js

const AUTH_BASE = '/api/auth'
const API_BASE = '/api'

// --- token utils ---

function getToken() {
  return localStorage.getItem('access_token')
}

function setToken(token) {
  localStorage.setItem('access_token', token)
}

function clearToken() {
  localStorage.removeItem('access_token')
}

// --- logout (единая точка) ---

async function fullLogout() {
  try {
    await fetch(`${AUTH_BASE}/logout`, {
      method: 'POST',
      credentials: 'include',
    })
  } catch { }

  clearToken()
  window.location.href = '/auth'
}

// --- refresh control ---

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
        setToken(data.access_token)
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

// --- core request ---

async function request(base, path, options = {}, retry = true) {
  const headers = { ...options.headers }

  // не ломаем FormData
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${base}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  })

  // --- 401 handling ---
  if (res.status === 401 && base === API_BASE && retry) {
    // если токена нет — смысла в refresh нет
    if (!getToken()) {
      await fullLogout()
      return res
    }

    const refreshed = await tryRefresh()

    if (refreshed) {
      return request(base, path, options, false)
    }

    await fullLogout()
    return res
  }

  return res
}

// --- json helper ---

async function json(base, path, options = {}) {
  const res = await request(base, path, options)

  if (!res.ok) {
    const err = await res.json().catch(() => ({
      error: res.statusText,
    }))

    throw Object.assign(
      new Error(err.error || 'request failed'),
      { status: res.status, data: err }
    )
  }

  return res.json()
}

// --- Auth service ---

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

// --- Main API ---

export const api = {
  // Профиль
  getProfile: () => json(API_BASE, '/profile'),

  updateProfile: (body) =>
    json(API_BASE, '/profile', {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  getAnswerHistory: (p) =>
    json(API_BASE, `/profile/answers?limit=${p?.limit || 20}&offset=${p?.offset || 0}`),

  // Тесты
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

  getTest: (id) =>
    json(API_BASE, `/tests/${id}`),

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

  // Мой ответ
  getMyAnswer: (testId) =>
    json(API_BASE, `/tests/${testId}/my-answer`),

  // Импорт
  importGoogleForm: (url) =>
    json(API_BASE, '/import/google-forms', {
      method: 'POST',
      body: JSON.stringify({ url }),
    }),

  // Админ
  setOfficial: (testId, official) =>
    json(API_BASE, `/admin/tests/${testId}/official`, {
      method: 'PUT',
      body: JSON.stringify({ official }),
    }),

  setStatus: (testId, status) =>
    json(API_BASE, `/admin/tests/${testId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),
}

export { clearToken, getToken, setToken }
