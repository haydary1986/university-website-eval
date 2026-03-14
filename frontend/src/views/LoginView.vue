<template>
  <v-container class="fill-height" fluid>
    <v-row justify="center" align="center">
      <v-col cols="12" sm="8" md="5" lg="4">
        <v-card class="pa-6" elevation="8" rounded="xl">
          <div class="text-center mb-6">
            <v-avatar size="100" rounded="0" class="mb-4">
              <v-img src="/mohesr-logo.svg" />
            </v-avatar>
            <h2 class="text-h5 text-primary font-weight-bold mb-2">نظام تقييم جودة المواقع الالكترونية</h2>
            <p class="text-subtitle-1 text-medium-emphasis">وزارة التعليم العالي والبحث العلمي</p>
          </div>

          <v-form @submit.prevent="handleLogin" ref="form">
            <v-text-field v-model="username" label="اسم المستخدم" prepend-inner-icon="mdi-account" :rules="[v => !!v || 'مطلوب']" class="mb-3" :disabled="isBlocked" />
            <v-text-field v-model="password" label="كلمة المرور" prepend-inner-icon="mdi-lock" :type="showPass ? 'text' : 'password'" :append-inner-icon="showPass ? 'mdi-eye-off' : 'mdi-eye'" @click:append-inner="showPass = !showPass" :rules="[v => !!v || 'مطلوب']" class="mb-3" :disabled="isBlocked" />

            <v-alert v-if="error" :type="isBlocked ? 'warning' : 'error'" variant="tonal" class="mb-3" closable @click:close="error = ''">
              <div>{{ error }}</div>
              <div v-if="remainingAttempts !== null && remainingAttempts > 0" class="text-caption mt-1">
                المحاولات المتبقية: <strong>{{ remainingAttempts }}</strong>
              </div>
            </v-alert>

            <v-btn type="submit" color="primary" block size="large" :loading="loading" :disabled="isBlocked" class="mb-2">
              <v-icon class="ml-1">mdi-login</v-icon>
              تسجيل الدخول
            </v-btn>
          </v-form>

          <div class="text-center mt-4">
            <v-icon size="small" color="success" class="ml-1">mdi-shield-check</v-icon>
            <span class="text-caption text-medium-emphasis">محمي بنظام أمان متعدد الطبقات</span>
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const username = ref('')
const password = ref('')
const showPass = ref(false)
const loading = ref(false)
const error = ref('')
const isBlocked = ref(false)
const remainingAttempts = ref(null)
const form = ref(null)

async function handleLogin() {
  const { valid } = await form.value.validate()
  if (!valid) return
  loading.value = true
  error.value = ''
  isBlocked.value = false
  remainingAttempts.value = null
  try {
    const data = await auth.login(username.value, password.value)
    if (data.must_change_password) {
      router.push('/change-password')
    } else {
      router.push('/dashboard')
    }
  } catch (e) {
    const resp = e.response?.data
    error.value = resp?.error || 'خطأ في تسجيل الدخول'
    if (resp?.blocked) {
      isBlocked.value = true
    }
    if (resp?.remaining_attempts !== undefined) {
      remainingAttempts.value = resp.remaining_attempts
    }
  } finally {
    loading.value = false
  }
}
</script>
