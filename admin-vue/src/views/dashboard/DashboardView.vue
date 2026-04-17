<template>
  <section class="page-section">
    <div class="page-grid">
      <div class="panel-card hero-card">
        <div class="card-eyebrow">Phase 1</div>
        <h2>Admin-Vue 初始化已起航</h2>
        <p>
          当前目标不是把后台一次性重写成宇宙战舰，而是先稳定跑通登录、文章管理与真实接口联调。
        </p>
        <RouterLink to="/articles">
          <el-button type="primary">进入文章管理</el-button>
        </RouterLink>
      </div>

      <div class="panel-card">
        <small>当前重点</small>
        <strong>文章主链路</strong>
        <p>列表、筛选、分页、编辑、新建、删除。</p>
      </div>

      <div class="panel-card">
        <small>下一阶段</small>
        <strong>创作效率增强</strong>
        <p>Markdown 编辑器、图床联动、自动保存草稿。</p>
      </div>

      <div class="panel-card">
        <small>联调策略</small>
        <strong>真实 API 优先</strong>
        <p>直接对接现有 Go 后端，不走纯静态假数据。</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
</script>
