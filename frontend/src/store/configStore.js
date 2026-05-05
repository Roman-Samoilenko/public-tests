import { reactive } from 'vue'

export const configStore = reactive({
    accessTokenTTL: 30 * 60,     // fallback 30m
    refreshTokenTTL: 720 * 3600, // fallback 30d
    loaded: false,
})

export async function loadConfig() {
    try {
        const res = await fetch('/api/auth/config')
        const data = await res.json()
        configStore.accessTokenTTL = data.access_token_ttl_seconds
        configStore.refreshTokenTTL = data.refresh_token_ttl_seconds
        configStore.loaded = true
    } catch (err) {
        console.warn('Failed to load config, using defaults')
        configStore.loaded = true
    }
}