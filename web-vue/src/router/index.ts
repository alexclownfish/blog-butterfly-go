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
    },
    {
      path: '/tags/:name',
      name: 'tag-detail',
      component: () => import('@/views/TagDetailPage.vue')
    }
  ]
})

export default router
