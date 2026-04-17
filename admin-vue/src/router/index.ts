import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const changePasswordPath = '/change-password'

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
      path: changePasswordPath,
      name: 'change-password',
      component: () => import('@/views/login/ChangePasswordView.vue'),
      meta: {
        requiresAuth: true,
        title: '修改密码'
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
        },
        {
          path: 'categories',
          name: 'categories',
          component: () => import('@/views/categories/CategoryListView.vue'),
          meta: {
            title: '分类管理'
          }
        }
      ]
    }
  ]
})

router.beforeEach((to) => {
  const authStore = useAuthStore()
  const isLoggedIn = Boolean(authStore.token)
  const forcePasswordChange = authStore.forcePasswordChange

  if (to.meta.requiresAuth && !isLoggedIn) {
    return '/login'
  }

  if (forcePasswordChange && isLoggedIn && to.path !== changePasswordPath) {
    return changePasswordPath
  }

  if (to.meta.guestOnly && isLoggedIn) {
    return forcePasswordChange ? changePasswordPath : '/dashboard'
  }

  return true
})

router.afterEach((to) => {
  const title = to.meta?.title ? `${String(to.meta.title)} | ${import.meta.env.VITE_APP_TITLE || 'Alexcld CMS'}` : (import.meta.env.VITE_APP_TITLE || 'Alexcld CMS')
  document.title = title
})

export default router
