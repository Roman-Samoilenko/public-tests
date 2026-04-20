import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '../api/index.js'

const routes = [
  { path: '/auth',    component: () => import('../pages/AuthPage.vue'),       meta: { public: true } },
  { path: '/',        component: () => import('../pages/FeedPage.vue'),        meta: { auth: true } },
  { path: '/tests/:id', component: () => import('../pages/TestPage.vue'),      meta: { auth: true } },
  { path: '/profile', component: () => import('../pages/ProfilePage.vue'),     meta: { auth: true } },
  { path: '/create',  component: () => import('../pages/CreateTestPage.vue'),  meta: { auth: true } },
  { path: '/admin',   component: () => import('../pages/AdminPage.vue'),       meta: { auth: true } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const loggedIn = !!getToken()
  if (to.meta.auth && !loggedIn) return '/auth'
  if (to.meta.public && loggedIn) return '/'
})

export default router
