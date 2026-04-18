import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomePage.vue')
    },
    {
      path: '/posts/:id.html',
      name: 'article-detail',
      component: () => import('@/views/ArticleDetailPage.vue')
    }
  ]
})

export default router
