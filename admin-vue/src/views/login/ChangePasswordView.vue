<template>
  <div class="change-password-view">
    <div class="change-password-card">
      <div class="card-eyebrow">🔐 首次登录安全检查</div>
      <h2>先修改默认密码</h2>
      <p class="card-desc">
        当前账号还在使用初始密码。为了后台安全，修改完成后才能继续进入内容工作台。
      </p>

      <el-alert
        v-if="successMessage"
        :title="successMessage"
        type="success"
        :closable="false"
        class="change-password-alert"
      />

      <el-alert
        v-if="errorMessage"
        :title="errorMessage"
        type="error"
        :closable="false"
        class="change-password-alert"
      />

      <el-form :model="form" label-position="top" @submit.prevent="handleSubmit">
        <el-form-item label="当前密码">
          <el-input
            v-model="form.old_password"
            type="password"
            show-password
            autocomplete="current-password"
            placeholder="请输入当前密码"
          />
        </el-form-item>

        <el-form-item label="新密码">
          <el-input
            v-model="form.new_password"
            type="password"
            show-password
            autocomplete="new-password"
            placeholder="请设置新的登录密码"
          />
        </el-form-item>

        <el-form-item label="确认新密码">
          <el-input
            v-model="form.confirm_password"
            type="password"
            show-password
            autocomplete="new-password"
            placeholder="请再次输入新密码"
            @keyup.enter="handleSubmit"
          />
        </el-form-item>

        <div class="password-tips">
          建议使用 8 位以上、包含大小写字母/数字/符号的组合。别再让默认密码当赛博裸奔侠啦。
        </div>

        <el-button type="primary" class="submit-btn" :loading="authStore.loading" @click="handleSubmit">
          保存新密码并进入后台
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const errorMessage = ref('')
const successMessage = ref('')

async function handleSubmit() {
  errorMessage.value = ''
  successMessage.value = ''

  if (!form.old_password.trim() || !form.new_password.trim() || !form.confirm_password.trim()) {
    errorMessage.value = '请完整填写当前密码、新密码和确认密码'
    return
  }

  if (form.new_password !== form.confirm_password) {
    errorMessage.value = '两次输入的新密码不一致'
    return
  }

  try {
    await authStore.changePassword({
      old_password: form.old_password,
      new_password: form.new_password,
      confirm_password: form.confirm_password
    })

    authStore.setForcePasswordChange(false)
    successMessage.value = '密码修改成功，正在带你跳转到工作台…'
    form.old_password = ''
    form.new_password = ''
    form.confirm_password = ''
    await router.replace('/dashboard')
  } catch (error: any) {
    errorMessage.value =
      error?.response?.data?.error ||
      error?.response?.data?.message ||
      error?.message ||
      '密码修改失败，请稍后重试'
  }
}
</script>

<style scoped>
.change-password-view {
  min-height: calc(100vh - 220px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.change-password-card {
  width: min(520px, 100%);
  padding: 28px;
  border-radius: 24px;
  background: rgba(9, 14, 31, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 22px 60px rgba(15, 23, 42, 0.35);
}

.card-eyebrow {
  margin-bottom: 10px;
  color: #8b5cf6;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

h2 {
  margin: 0;
  font-size: 28px;
  color: #f8fafc;
}

.card-desc {
  margin: 12px 0 20px;
  color: rgba(226, 232, 240, 0.82);
  line-height: 1.7;
}

.change-password-alert {
  margin-bottom: 16px;
}

.password-tips {
  margin: 6px 0 18px;
  color: rgba(148, 163, 184, 0.95);
  font-size: 13px;
  line-height: 1.7;
}

.submit-btn {
  width: 100%;
}
</style>
