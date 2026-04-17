import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login/LoginView.vue'),
      meta: {
        guestOnly: true,
        title: '登录'
      }
    },
    {
      path: '/',
      component: () => import('@/components/layout/AdminLayout.vue'),
      meta: {
        requiresAuth: true
      },
      children: [
        {
          path: '',
          redirect: '/dashboard'
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: {
            title: '工作台'
          }
        },
        {
          path: 'articles',
          name: 'articles',
          component: () => import('@/views/articles/ArticleListView.vue'),
          meta: {
            title: '文章管理'
          }
        }
      ]
    }
  ]
})

router.beforeEach((to) => {
  const token = getToken()
  const isLoggedIn = Boolean(token)

  if (to.meta.requiresAuth && !isLoggedIn) {
    return '/login'
  }

  if (to.meta.guestOnly && isLoggedIn) {
    return '/dashboard'
  }

  return true
})

router.afterEach((to) => {
  const title = to.meta?.title ? `${String(to.meta.title)} | ${import.meta.env.VITE_APP_TITLE || 'Alexcld CMS'}` : (import.meta.env.VITE_APP_TITLE || 'Alexcld CMS')
  document.title = title
})

export default router
