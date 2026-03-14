<template>
  <v-container class="fill-height" fluid>
    <v-row justify="center" align="center">
      <v-col cols="12" sm="8" md="5" lg="4">
        <v-card class="pa-6" elevation="8" rounded="xl">
          <div class="text-center mb-6">
            <v-icon size="64" color="warning" class="mb-3">mdi-lock-alert</v-icon>
            <h2 class="text-h5 text-primary font-weight-bold mb-2">تغيير كلمة المرور</h2>
            <p class="text-subtitle-2 text-medium-emphasis">يجب تغيير كلمة المرور الافتراضية قبل المتابعة</p>
          </div>

          <v-form @submit.prevent="handleChange" ref="form">
            <v-text-field
              v-model="oldPassword"
              label="كلمة المرور الحالية"
              prepend-inner-icon="mdi-lock-outline"
              :type="showOld ? 'text' : 'password'"
              :append-inner-icon="showOld ? 'mdi-eye-off' : 'mdi-eye'"
              @click:append-inner="showOld = !showOld"
              :rules="[v => !!v || 'مطلوب']"
              class="mb-3"
            />
            <v-text-field
              v-model="newPassword"
              label="كلمة المرور الجديدة"
              prepend-inner-icon="mdi-lock-plus"
              :type="showNew ? 'text' : 'password'"
              :append-inner-icon="showNew ? 'mdi-eye-off' : 'mdi-eye'"
              @click:append-inner="showNew = !showNew"
              :rules="[
                v => !!v || 'مطلوب',
                v => v.length >= 8 || 'يجب أن تكون 8 أحرف على الأقل',
                v => v !== oldPassword || 'يجب أن تكون مختلفة عن الحالية'
              ]"
              class="mb-3"
            />
            <v-text-field
              v-model="confirmPassword"
              label="تأكيد كلمة المرور الجديدة"
              prepend-inner-icon="mdi-lock-check"
              :type="showConfirm ? 'text' : 'password'"
              :append-inner-icon="showConfirm ? 'mdi-eye-off' : 'mdi-eye'"
              @click:append-inner="showConfirm = !showConfirm"
              :rules="[
                v => !!v || 'مطلوب',
                v => v === newPassword || 'كلمات المرور غير متطابقة'
              ]"
              class="mb-3"
            />
            <v-alert v-if="error" type="error" variant="tonal" class="mb-3" closable @click:close="error = ''">{{ error }}</v-alert>
            <v-alert v-if="success" type="success" variant="tonal" class="mb-3">{{ success }}</v-alert>
            <v-btn type="submit" color="primary" block size="large" :loading="loading">تغيير كلمة المرور</v-btn>
          </v-form>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const auth = useAuthStore()
const router = useRouter()
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showOld = ref(false)
const showNew = ref(false)
const showConfirm = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')
const form = ref(null)

async function handleChange() {
  const { valid } = await form.value.validate()
  if (!valid) return
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await api.changePassword({
      old_password: oldPassword.value,
      new_password: newPassword.value
    })
    success.value = 'تم تغيير كلمة المرور بنجاح، سيتم تحويلك...'
    auth.mustChangePassword = false
    if (auth.user) auth.user.must_change_password = false
    localStorage.setItem('user', JSON.stringify(auth.user))
    setTimeout(() => router.push('/dashboard'), 1500)
  } catch (e) {
    error.value = e.response?.data?.error || 'خطأ في تغيير كلمة المرور'
  } finally {
    loading.value = false
  }
}
</script>
