<template>
  <v-container fluid>
    <h1 class="text-h4 mb-6">
      <v-icon class="ml-2">mdi-shield-lock</v-icon>
      مركز الأمان
    </h1>

    <!-- Security Stats Cards -->
    <v-row class="mb-6">
      <v-col cols="12" sm="6" md="3">
        <v-card color="success" variant="tonal" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="48" color="success" class="ml-4">mdi-login</v-icon>
            <div>
              <div class="text-h4 font-weight-bold">{{ stats.login_success_24h || 0 }}</div>
              <div class="text-caption">دخول ناجح (24 ساعة)</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="error" variant="tonal" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="48" color="error" class="ml-4">mdi-login-variant</v-icon>
            <div>
              <div class="text-h4 font-weight-bold">{{ stats.login_failed_24h || 0 }}</div>
              <div class="text-caption">محاولات فاشلة (24 ساعة)</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="warning" variant="tonal" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="48" color="warning" class="ml-4">mdi-account-lock</v-icon>
            <div>
              <div class="text-h4 font-weight-bold">{{ stats.blocked_users || 0 }}</div>
              <div class="text-caption">حسابات محظورة</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card color="info" variant="tonal" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="48" color="info" class="ml-4">mdi-monitor-cellphone</v-icon>
            <div>
              <div class="text-h4 font-weight-bold">{{ stats.active_sessions || 0 }}</div>
              <div class="text-caption">جلسات نشطة</div>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mb-6">
      <v-col cols="12" sm="6" md="3">
        <v-card variant="outlined" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="36" color="error" class="ml-3">mdi-ip-network</v-icon>
            <div>
              <div class="text-h5 font-weight-bold">{{ stats.blocked_ips || 0 }}</div>
              <div class="text-caption">عناوين IP محظورة</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card variant="outlined" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="36" color="orange" class="ml-3">mdi-alert</v-icon>
            <div>
              <div class="text-h5 font-weight-bold">{{ stats.login_failed_week || 0 }}</div>
              <div class="text-caption">فاشلة (هذا الأسبوع)</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card variant="outlined" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="36" color="primary" class="ml-3">mdi-key-change</v-icon>
            <div>
              <div class="text-h5 font-weight-bold">{{ stats.password_changes_24h || 0 }}</div>
              <div class="text-caption">تغيير كلمة مرور (24س)</div>
            </div>
          </div>
        </v-card>
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <v-card variant="outlined" class="pa-4">
          <div class="d-flex align-center">
            <v-icon size="36" color="success" class="ml-3">mdi-account-group</v-icon>
            <div>
              <div class="text-h5 font-weight-bold">{{ stats.total_users || 0 }}</div>
              <div class="text-caption">إجمالي المستخدمين</div>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Tabs for different security sections -->
    <v-card>
      <v-tabs v-model="activeTab" color="primary" grow>
        <v-tab value="attempts"><v-icon class="ml-1">mdi-login-variant</v-icon> محاولات الدخول</v-tab>
        <v-tab value="blocked"><v-icon class="ml-1">mdi-account-lock</v-icon> الحسابات المحظورة</v-tab>
        <v-tab value="ips"><v-icon class="ml-1">mdi-ip-network</v-icon> عناوين IP</v-tab>
        <v-tab value="sessions"><v-icon class="ml-1">mdi-monitor</v-icon> الجلسات النشطة</v-tab>
      </v-tabs>

      <v-card-text>
        <!-- Login Attempts Tab -->
        <div v-if="activeTab === 'attempts'">
          <v-row class="mb-4">
            <v-col cols="12" md="3">
              <v-select v-model="attemptFilter" :items="attemptFilterOptions" label="نوع المحاولة" clearable @update:model-value="loadAttempts" />
            </v-col>
            <v-col cols="12" md="3">
              <v-text-field v-model="attemptSearch" label="بحث باسم المستخدم" prepend-inner-icon="mdi-magnify" clearable @keyup.enter="loadAttempts" />
            </v-col>
            <v-col cols="12" md="3">
              <v-text-field v-model="attemptIP" label="بحث بعنوان IP" prepend-inner-icon="mdi-ip" clearable @keyup.enter="loadAttempts" />
            </v-col>
            <v-col cols="12" md="3" class="d-flex align-center">
              <v-btn color="primary" @click="loadAttempts" prepend-icon="mdi-refresh">تحديث</v-btn>
            </v-col>
          </v-row>

          <v-data-table :headers="attemptHeaders" :items="attempts" :loading="loadingAttempts" items-per-page="20">
            <template #item.success="{ item }">
              <v-icon :color="item.success ? 'success' : 'error'" size="small">
                {{ item.success ? 'mdi-check-circle' : 'mdi-close-circle' }}
              </v-icon>
            </template>
            <template #item.reason="{ item }">
              <v-chip :color="reasonColor(item.reason)" size="small" variant="tonal">{{ reasonLabel(item.reason) }}</v-chip>
            </template>
            <template #item.created_at="{ item }">
              {{ formatDate(item.created_at) }}
            </template>
            <template #item.ip_address="{ item }">
              <code>{{ item.ip_address }}</code>
              <v-btn icon size="x-small" variant="text" color="error" @click="blockIPDialog(item.ip_address)" class="mr-1">
                <v-icon size="small">mdi-block-helper</v-icon>
                <v-tooltip activator="parent">حظر هذا العنوان</v-tooltip>
              </v-btn>
            </template>
          </v-data-table>
        </div>

        <!-- Blocked Users Tab -->
        <div v-if="activeTab === 'blocked'">
          <v-alert v-if="blockedUsers.length === 0" type="success" variant="tonal" class="mb-4">
            لا توجد حسابات محظورة حالياً
          </v-alert>
          <v-data-table v-else :headers="blockedHeaders" :items="blockedUsers" items-per-page="20">
            <template #item.blocked_until="{ item }">
              <span v-if="item.blocked_until">{{ formatDate(item.blocked_until) }}</span>
              <v-chip v-else color="error" size="small">دائم</v-chip>
            </template>
            <template #item.failed_attempts="{ item }">
              <v-chip color="error" size="small" variant="tonal">{{ item.failed_attempts }}</v-chip>
            </template>
            <template #item.last_failed_at="{ item }">
              {{ item.last_failed_at ? formatDate(item.last_failed_at) : '-' }}
            </template>
            <template #item.actions="{ item }">
              <v-btn color="success" size="small" variant="tonal" @click="handleUnblock(item)" prepend-icon="mdi-lock-open">
                إلغاء الحظر
              </v-btn>
            </template>
          </v-data-table>
        </div>

        <!-- Blocked IPs Tab -->
        <div v-if="activeTab === 'ips'">
          <v-btn color="error" class="mb-4" @click="showBlockIPDialog = true" prepend-icon="mdi-plus">حظر عنوان IP</v-btn>

          <v-alert v-if="blockedIPs.length === 0" type="success" variant="tonal" class="mb-4">
            لا توجد عناوين IP محظورة حالياً
          </v-alert>
          <v-data-table v-else :headers="ipHeaders" :items="blockedIPs" items-per-page="20">
            <template #item.expires_at="{ item }">
              <span v-if="item.expires_at">{{ formatDate(item.expires_at) }}</span>
              <v-chip v-else color="error" size="small">دائم</v-chip>
            </template>
            <template #item.created_at="{ item }">
              {{ formatDate(item.created_at) }}
            </template>
            <template #item.actions="{ item }">
              <v-btn color="success" size="small" variant="tonal" @click="handleUnblockIP(item.ip_address)" prepend-icon="mdi-lock-open">
                إلغاء الحظر
              </v-btn>
            </template>
          </v-data-table>
        </div>

        <!-- Active Sessions Tab -->
        <div v-if="activeTab === 'sessions'">
          <v-data-table :headers="sessionHeaders" :items="sessions" :loading="loadingSessions" items-per-page="20">
            <template #item.user="{ item }">
              <div>
                <strong>{{ item.user?.full_name || item.user?.username || '-' }}</strong>
                <div class="text-caption text-medium-emphasis">{{ item.user?.role }}</div>
              </div>
            </template>
            <template #item.ip_address="{ item }">
              <code>{{ item.ip_address }}</code>
            </template>
            <template #item.created_at="{ item }">
              {{ formatDate(item.created_at) }}
            </template>
            <template #item.expires_at="{ item }">
              {{ formatDate(item.expires_at) }}
            </template>
            <template #item.actions="{ item }">
              <v-btn color="error" size="small" variant="tonal" @click="handleTerminateSession(item.id)" prepend-icon="mdi-close">
                إنهاء
              </v-btn>
            </template>
          </v-data-table>
        </div>
      </v-card-text>
    </v-card>

    <!-- Top Threats Section -->
    <v-row class="mt-6" v-if="stats.top_failed_ips?.length || stats.top_failed_users?.length">
      <v-col cols="12" md="6" v-if="stats.top_failed_ips?.length">
        <v-card>
          <v-card-title class="text-subtitle-1">
            <v-icon class="ml-1" color="error">mdi-alert-circle</v-icon>
            أكثر العناوين فشلاً (هذا الأسبوع)
          </v-card-title>
          <v-list density="compact">
            <v-list-item v-for="ip in stats.top_failed_ips" :key="ip.ip_address">
              <template v-slot:prepend>
                <code class="ml-2">{{ ip.ip_address }}</code>
              </template>
              <template v-slot:append>
                <v-chip color="error" size="small" variant="tonal">{{ ip.count }} محاولة</v-chip>
                <v-btn icon size="x-small" color="error" variant="text" class="mr-1" @click="blockIPDialog(ip.ip_address)">
                  <v-icon>mdi-block-helper</v-icon>
                </v-btn>
              </template>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
      <v-col cols="12" md="6" v-if="stats.top_failed_users?.length">
        <v-card>
          <v-card-title class="text-subtitle-1">
            <v-icon class="ml-1" color="warning">mdi-account-alert</v-icon>
            أكثر المستخدمين فشلاً (هذا الأسبوع)
          </v-card-title>
          <v-list density="compact">
            <v-list-item v-for="u in stats.top_failed_users" :key="u.username">
              <template v-slot:prepend>
                <strong class="ml-2">{{ u.username }}</strong>
              </template>
              <template v-slot:append>
                <v-chip color="warning" size="small" variant="tonal">{{ u.count }} محاولة</v-chip>
              </template>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>

    <!-- Block IP Dialog -->
    <v-dialog v-model="showBlockIPDialog" max-width="500">
      <v-card>
        <v-card-title>حظر عنوان IP</v-card-title>
        <v-card-text>
          <v-text-field v-model="blockIPForm.ip_address" label="عنوان IP" dir="ltr" />
          <v-text-field v-model="blockIPForm.reason" label="السبب" />
          <v-select v-model="blockIPForm.duration" :items="durationOptions" label="مدة الحظر" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="showBlockIPDialog = false">إلغاء</v-btn>
          <v-btn color="error" @click="handleBlockIP">حظر</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '../services/api'

const activeTab = ref('attempts')
const stats = ref({})
const attempts = ref([])
const blockedUsers = ref([])
const blockedIPs = ref([])
const sessions = ref([])
const loadingAttempts = ref(false)
const loadingSessions = ref(false)

const attemptFilter = ref(null)
const attemptSearch = ref('')
const attemptIP = ref('')
const showBlockIPDialog = ref(false)

const blockIPForm = ref({ ip_address: '', reason: '', duration: 0 })

const attemptFilterOptions = [
  { title: 'الكل', value: null },
  { title: 'ناجحة فقط', value: 'true' },
  { title: 'فاشلة فقط', value: 'false' },
]

const durationOptions = [
  { title: 'ساعة واحدة', value: 1 },
  { title: '6 ساعات', value: 6 },
  { title: '24 ساعة', value: 24 },
  { title: 'أسبوع', value: 168 },
  { title: 'دائم', value: 0 },
]

const attemptHeaders = [
  { title: 'التاريخ', key: 'created_at', width: '170px' },
  { title: 'المستخدم', key: 'username', width: '150px' },
  { title: 'النتيجة', key: 'success', width: '80px', align: 'center' },
  { title: 'السبب', key: 'reason', width: '150px' },
  { title: 'عنوان IP', key: 'ip_address', width: '180px' },
  { title: 'المتصفح', key: 'user_agent' },
]

const blockedHeaders = [
  { title: 'المستخدم', key: 'username', width: '150px' },
  { title: 'الاسم', key: 'full_name' },
  { title: 'محاولات فاشلة', key: 'failed_attempts', width: '120px', align: 'center' },
  { title: 'آخر محاولة', key: 'last_failed_at', width: '170px' },
  { title: 'محظور حتى', key: 'blocked_until', width: '170px' },
  { title: 'الإجراءات', key: 'actions', width: '150px', sortable: false },
]

const ipHeaders = [
  { title: 'عنوان IP', key: 'ip_address', width: '150px' },
  { title: 'السبب', key: 'reason' },
  { title: 'تاريخ الحظر', key: 'created_at', width: '170px' },
  { title: 'ينتهي في', key: 'expires_at', width: '170px' },
  { title: 'الإجراءات', key: 'actions', width: '150px', sortable: false },
]

const sessionHeaders = [
  { title: 'المستخدم', key: 'user', sortable: false },
  { title: 'عنوان IP', key: 'ip_address', width: '150px' },
  { title: 'بدأت في', key: 'created_at', width: '170px' },
  { title: 'تنتهي في', key: 'expires_at', width: '170px' },
  { title: 'الإجراءات', key: 'actions', width: '120px', sortable: false },
]

function formatDate(d) {
  if (!d) return '-'
  return new Date(d).toLocaleString('ar-IQ')
}

function reasonColor(reason) {
  const colors = {
    success: 'success', invalid_password: 'error', user_not_found: 'grey',
    account_blocked: 'warning', ip_blocked: 'error', ip_rate_limited: 'error',
    account_auto_blocked: 'error'
  }
  return colors[reason] || 'grey'
}

function reasonLabel(reason) {
  const labels = {
    success: 'ناجح', invalid_password: 'كلمة مرور خاطئة', user_not_found: 'مستخدم غير موجود',
    account_blocked: 'حساب محظور', ip_blocked: 'IP محظور', ip_rate_limited: 'تجاوز الحد',
    account_auto_blocked: 'حظر تلقائي'
  }
  return labels[reason] || reason
}

async function loadStats() {
  try {
    const res = await api.getSecurityOverview()
    stats.value = res.data
  } catch (e) { console.error(e) }
}

async function loadAttempts() {
  loadingAttempts.value = true
  try {
    const params = {}
    if (attemptFilter.value !== null) params.success = attemptFilter.value
    if (attemptSearch.value) params.username = attemptSearch.value
    if (attemptIP.value) params.ip = attemptIP.value
    const res = await api.getLoginAttempts(params)
    attempts.value = res.data.attempts || []
  } catch (e) { console.error(e) }
  finally { loadingAttempts.value = false }
}

async function loadBlockedUsers() {
  try {
    const res = await api.getUsers()
    blockedUsers.value = (res.data || []).filter(u => u.is_blocked)
  } catch (e) { console.error(e) }
}

async function loadBlockedIPs() {
  try {
    const res = await api.getBlockedIPs()
    blockedIPs.value = res.data.blocked_ips || []
  } catch (e) { console.error(e) }
}

async function loadSessions() {
  loadingSessions.value = true
  try {
    const res = await api.getAllSessions()
    sessions.value = res.data.sessions || []
  } catch (e) { console.error(e) }
  finally { loadingSessions.value = false }
}

async function handleUnblock(user) {
  try {
    await api.unblockUser(user.id)
    await loadBlockedUsers()
    await loadStats()
  } catch (e) { console.error(e) }
}

async function handleUnblockIP(ip) {
  try {
    await api.unblockIP(ip)
    await loadBlockedIPs()
    await loadStats()
  } catch (e) { console.error(e) }
}

async function handleBlockIP() {
  try {
    await api.blockIP(blockIPForm.value)
    showBlockIPDialog.value = false
    blockIPForm.value = { ip_address: '', reason: '', duration: 0 }
    await loadBlockedIPs()
    await loadStats()
  } catch (e) { console.error(e) }
}

function blockIPDialog(ip) {
  blockIPForm.value.ip_address = ip
  showBlockIPDialog.value = true
}

async function handleTerminateSession(id) {
  try {
    await api.terminateSession(id)
    await loadSessions()
    await loadStats()
  } catch (e) { console.error(e) }
}

watch(activeTab, (tab) => {
  if (tab === 'attempts') loadAttempts()
  if (tab === 'blocked') loadBlockedUsers()
  if (tab === 'ips') loadBlockedIPs()
  if (tab === 'sessions') loadSessions()
})

onMounted(() => {
  loadStats()
  loadAttempts()
})
</script>
