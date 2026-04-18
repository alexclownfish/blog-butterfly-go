<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

const route = useRoute()
const theme = ref<'dark' | 'light'>('dark')

const themeLabel = computed(() => theme.value === 'dark' ? '暗色流光' : '柔光纸感')
const themeHint = computed(() => theme.value === 'dark' ? '点击切到明亮模式' : '点击切回暗色模式')
const themeIcon = computed(() => theme.value === 'dark' ? '🌙' : '☀️')
const shellMode = computed(() => (route.name === 'article-detail' || route.name === 'tag-detail') ? 'detail' : 'home')

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
}

watch(theme, (value) => {
  document.body.dataset.theme = value
  localStorage.setItem('web-vue-theme', value)
}, { immediate: true })

watch(() => route.fullPath, () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
})

onMounted(() => {
  const storedTheme = localStorage.getItem('web-vue-theme')
  if (storedTheme === 'dark' || storedTheme === 'light') {
    theme.value = storedTheme
    return
  }

  const prefersLight = window.matchMedia?.('(prefers-color-scheme: light)').matches
  theme.value = prefersLight ? 'light' : 'dark'
})
</script>

<template>
  <div class="aurora-bg" aria-hidden="true">
    <div class="aurora-layer one"></div>
    <div class="aurora-layer two"></div>
    <div class="aurora-layer three"></div>
    <div class="aurora-beam"></div>
  </div>

  <div class="site-shell" :data-mode="shellMode">
    <div class="container">
      <header class="topbar">
        <div class="brand">
          <div class="brand-mark">A</div>
          <div class="brand-copy">
            <strong>children don't lie</strong>
            <span>导航 / 技术记录 / 运维 / 开发</span>
          </div>
        </div>
        <div class="topbar-meta">
          <button class="theme-toggle" type="button" aria-label="切换主题" @click="toggleTheme">
            <span class="theme-toggle-icon">{{ themeIcon }}</span>
            <span class="theme-toggle-text">
              <strong>{{ themeLabel }}</strong>
              <small>{{ themeHint }}</small>
            </span>
          </button>
          <RouterLink class="pill" to="/">首页</RouterLink>
          <RouterLink class="pill" to="/tags/运维">标签</RouterLink>
          <a class="pill" href="/categories/">分类</a>
          <a class="pill" href="/archives/">归档</a>
          <a class="pill" href="/about/">博主</a>
        </div>
      </header>

      <RouterView />

      <footer class="footer-note">Alexcld Home · Vue runtime edition · 从原站视觉语言平滑迁移</footer>
    </div>
  </div>
</template>
