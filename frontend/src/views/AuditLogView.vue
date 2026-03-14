<template>
  <v-container fluid>
    <h1 class="text-h4 mb-6">سجل العمليات</h1>

    <v-card class="mb-4">
      <v-card-text>
        <v-row>
          <v-col cols="12" md="4">
            <v-select
              v-model="filterAction"
              :items="actionOptions"
              label="نوع العملية"
              clearable
              @update:model-value="loadLogs"
            />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field
              v-model="filterUser"
              label="بحث بالمستخدم"
              prepend-inner-icon="mdi-magnify"
              clearable
            />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <v-card>
      <v-data-table
        :headers="headers"
        :items="filteredLogs"
        :loading="loading"
        items-per-page="20"
        class="elevation-1"
      >
        <template #item.action="{ item }">
          <v-chip :color="actionColor(item.action)" size="small" variant="tonal">
            {{ actionLabel(item.action) }}
          </v-chip>
        </template>
        <template #item.user="{ item }">
          <div>
            <strong>{{ item.user?.full_name || item.user?.username || '-' }}</strong>
            <div class="text-caption text-medium-emphasis">{{ item.user?.role }}</div>
          </div>
        </template>
        <template #item.created_at="{ item }">
          {{ new Date(item.created_at).toLocaleString('ar-IQ') }}
        </template>
        <template #item.ip_address="{ item }">
          <code>{{ item.ip_address }}</code>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../services/api'

const logs = ref([])
const loading = ref(false)
const filterAction = ref(null)
const filterUser = ref('')

const actionOptions = [
  { title: 'تسجيل دخول', value: 'login' },
  { title: 'محاولة دخول فاشلة', value: 'login_failed' },
  { title: 'تسجيل خروج', value: 'logout' },
  { title: 'تغيير كلمة المرور', value: 'password_change' },
  { title: 'حظر حساب', value: 'account_blocked' },
  { title: 'إلغاء حظر حساب', value: 'account_unblocked' },
  { title: 'حظر حساب يدوي', value: 'account_blocked_manual' },
  { title: 'حظر IP', value: 'ip_blocked' },
  { title: 'إلغاء حظر IP', value: 'ip_unblocked' },
  { title: 'إنشاء مستخدم', value: 'user_created' },
  { title: 'إنهاء جلسة', value: 'session_terminated' },
]

const headers = [
  { title: 'التاريخ', key: 'created_at', width: '180px' },
  { title: 'المستخدم', key: 'user', sortable: false },
  { title: 'العملية', key: 'action', width: '150px' },
  { title: 'عنوان IP', key: 'ip_address', width: '150px' },
  { title: 'المتصفح', key: 'user_agent' },
  { title: 'التفاصيل', key: 'details' },
]

const filteredLogs = computed(() => {
  if (!filterUser.value) return logs.value
  const q = filterUser.value.toLowerCase()
  return logs.value.filter(l =>
    l.user?.full_name?.toLowerCase().includes(q) ||
    l.user?.username?.toLowerCase().includes(q)
  )
})

function actionColor(action) {
  const colors = {
    login: 'success', login_failed: 'error', logout: 'blue',
    password_change: 'orange', account_blocked: 'error', account_unblocked: 'success',
    account_blocked_manual: 'error', ip_blocked: 'error', ip_unblocked: 'success',
    user_created: 'info', session_terminated: 'warning',
    all_sessions_terminated: 'warning',
  }
  return colors[action] || 'grey'
}

function actionLabel(action) {
  const labels = {
    login: 'تسجيل دخول', login_failed: 'دخول فاشل', logout: 'تسجيل خروج',
    password_change: 'تغيير كلمة مرور', account_blocked: 'حظر حساب (تلقائي)',
    account_unblocked: 'إلغاء حظر حساب', account_blocked_manual: 'حظر حساب (يدوي)',
    ip_blocked: 'حظر IP', ip_unblocked: 'إلغاء حظر IP',
    user_created: 'إنشاء مستخدم', session_terminated: 'إنهاء جلسة',
    all_sessions_terminated: 'إنهاء جميع الجلسات',
  }
  return labels[action] || action
}

async function loadLogs() {
  loading.value = true
  try {
    const params = {}
    if (filterAction.value) params.action = filterAction.value
    const res = await api.getAuditLogs(params)
    logs.value = res.data.logs || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadLogs)
</script>
