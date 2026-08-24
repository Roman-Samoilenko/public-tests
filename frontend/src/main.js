import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'
import { loadConfig } from './store/configStore'
import './assets/global.css'
import BackLink from './components/BackLink.vue'
import ImportGoogleForm from './components/ImportForm.vue'

async function init() {
    await loadConfig()
    createApp(App).use(router).mount('#app')
}

init()