import { createApp } from 'vue'
import './main.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(router)

app.config.errorHandler = (err, _instance, info) => {
  console.error('[DIVI Error Boundary]', err)
  window.dispatchEvent(new CustomEvent('divi:app-error', {
    detail: { message: err instanceof Error ? err.message : String(err), info }
  }))
}

app.mount('#app')
