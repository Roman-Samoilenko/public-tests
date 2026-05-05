import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'
import { loadConfig } from './store/configStore'

async function init() {
    await loadConfig()
    createApp(App).use(router).mount('#app')
}

init()