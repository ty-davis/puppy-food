import { createRouter, createWebHistory } from 'vue-router'
import HomeView from './views/HomeView.vue'
import ConfigureView from './views/ConfigureView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/configure', component: ConfigureView },
  ],
})

export default router
