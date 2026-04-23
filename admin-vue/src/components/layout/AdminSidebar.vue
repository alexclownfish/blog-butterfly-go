<template>
  <aside class="sidebar">
    <div class="sidebar-brand">
      <div class="brand-mark">A</div>
      <div>
        <strong>Alexcld CMS</strong>
        <span>内容管理后台</span>
      </div>
    </div>

    <div class="sidebar-group">
      <div class="sidebar-label">导航</div>

      <RouterLink to="/dashboard" class="sidebar-link">
        <span>🏠</span>
        <span>工作台</span>
      </RouterLink>

      <RouterLink to="/articles" class="sidebar-link">
        <span>📄</span>
        <span>文章管理</span>
      </RouterLink>

      <RouterLink to="/articles/csdn-sync" class="sidebar-link">
        <span>🪄</span>
        <span>CSDN 同步导入</span>
      </RouterLink>

      <RouterLink to="/categories" class="sidebar-link">
        <span>🗂️</span>
        <span>分类管理</span>
      </RouterLink>

      <RouterLink to="/images" class="sidebar-link">
        <span>🖼️</span>
        <span>素材管理</span>
      </RouterLink>
    </div>

    <div class="sidebar-footnote">
      <div class="foot-card">
        <small>当前阶段</small>
        <strong>Vue 后台初始化</strong>
      </div>
      <div class="foot-card">
        <small>主线目标</small>
        <strong>文章管理真实联调</strong>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
</script>
