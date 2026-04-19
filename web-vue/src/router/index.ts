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
    },
    {
      path: '/categories/',
      name: 'categories',
      component: () => import('@/views/CategoriesPage.vue')
    },
    {
      path: '/archives/',
      name: 'archives',
      component: () => import('@/views/ArchivesPage.vue')
    },
    {
      path: '/about/',
      name: 'about',
      component: () => import('@/views/AboutPage.vue')
    }
  ]
})

export default router
