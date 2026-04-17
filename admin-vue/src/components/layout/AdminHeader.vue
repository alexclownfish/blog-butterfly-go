<template>
  <header class="header">
    <div>
      <div class="header-eyebrow">🛸 Content management hub</div>
      <h1>{{ pageTitle }}</h1>
      <p>{{ pageDescription }}</p>
    </div>

    <div class="header-actions">
      <div v-if="authStore.forcePasswordChange" class="status-pill status-pill--warning">等待修改密码</div>
      <div v-else class="status-pill">系统在线</div>
      <el-button class="ghost-btn" @click="handleLogout">退出登录</el-button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const pageTitle = computed(() => {
  if (route.path === '/change-password') return '修改密码'
  if (route.path.startsWith('/articles')) return '文章管理'
  if (route.path.startsWith('/dashboard')) return '工作台'
  return '内容管理后台'
})

const pageDescription = computed(() => {
  if (route.path === '/change-password') {
    return '先完成安全校验，把默认密码升级掉，再继续快乐搬砖。'
  }
  if (route.path.startsWith('/articles')) {
    return '统一处理文章发布、编辑、筛选与基础内容维护，优先保障真实接口联调可用。'
  }
  if (route.path.startsWith('/dashboard')) {
    return '这是新一代 admin-vue 后台工作台的起点，先把内容主链路跑通，再做效率增强。'
  }
  return '围绕内容生产与运营效率设计，逐步替代旧后台。'
})

function handleLogout() {
  authStore.logout()
  router.replace('/login')
}
</script>
