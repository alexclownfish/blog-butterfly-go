<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-eyebrow">✨ Admin Vue Bootstrap</div>
      <h1>欢迎回来</h1>
      <p class="login-desc">
        登录新后台工作台，先把文章主链路跑顺，再慢慢升级成创作效率怪兽。
      </p>

      <el-alert
        v-if="errorMessage"
        :title="errorMessage"
        type="error"
        :closable="false"
        class="login-alert"
      />

      <el-form :model="form" @submit.prevent="handleSubmit" class="login-form">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" clearable />
        </el-form-item>

        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            clearable
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <el-button
          type="primary"
          size="large"
          :loading="authStore.loading"
          class="login-btn"
          @click="handleSubmit"
        >
          登录后台
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  username: '',
  password: ''
})

const errorMessage = ref('')

onMounted(() => {
  authStore.setForcePasswordChange(false)
})

async function handleSubmit() {
  errorMessage.value = ''

  if (!form.username.trim() || !form.password.trim()) {
    errorMessage.value = '请输入用户名和密码'
    return
  }

  try {
    const result = await authStore.login({
      username: form.username.trim(),
      password: form.password
    })
    router.replace(result.forcePasswordChange ? '/change-password' : '/dashboard')
  } catch (error: any) {
    errorMessage.value =
      error?.response?.data?.error ||
      error?.response?.data?.message ||
      error?.message ||
      '登录失败，请检查账号密码'
  }
}
</script>
