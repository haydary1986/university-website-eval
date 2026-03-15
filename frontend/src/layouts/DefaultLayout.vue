<template>
  <div>
    <v-navigation-drawer v-model="drawer" :rail="rail" permanent color="primary" theme="dark">
      <v-list-item class="pa-4">
        <template v-slot:prepend>
          <v-avatar size="40" rounded="0">
            <v-img src="/mohesr-logo.svg" />
          </v-avatar>
        </template>
        <v-list-item-title class="text-subtitle-1 font-weight-bold">نظام تقييم المواقع</v-list-item-title>
        <v-list-item-subtitle>وزارة التعليم العالي</v-list-item-subtitle>
        <template v-slot:append>
          <v-btn icon variant="text" @click="rail = !rail">
            <v-icon>{{ rail ? 'mdi-chevron-left' : 'mdi-chevron-right' }}</v-icon>
          </v-btn>
        </template>
      </v-list-item>

      <v-divider />

      <v-list nav density="comfortable">
        <v-list-item v-for="item in menuItems" :key="item.to" :to="item.to" :prepend-icon="item.icon" :title="item.title" rounded="xl" />
      </v-list>
    </v-navigation-drawer>

    <v-app-bar flat color="white" border="b">
      <v-app-bar-nav-icon @click="drawer = !drawer" />
      <v-toolbar-title class="text-primary font-weight-bold">
        نظام تقييم جودة المواقع الالكترونية الجامعية
      </v-toolbar-title>
      <v-spacer />
      <v-chip class="ml-2" color="primary" variant="tonal" prepend-icon="mdi-account">
        {{ auth.user?.full_name || auth.user?.username }}
      </v-chip>
      <v-chip class="ml-2" size="small" :color="roleColor">{{ roleLabel }}</v-chip>
      <v-btn icon variant="text" @click="handleLogout" class="ml-2">
        <v-icon>mdi-logout</v-icon>
        <v-tooltip activator="parent">تسجيل الخروج</v-tooltip>
      </v-btn>
    </v-app-bar>

    <v-main>
      <v-container fluid class="pa-6">
        <router-view />
      </v-container>
    </v-main>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const drawer = ref(true)
const rail = ref(false)

const roleColor = computed(() => {
  const map = { super_admin: 'error', admin: 'warning', university: 'info' }
  return map[auth.userRole] || 'grey'
})

const roleLabel = computed(() => {
  const map = { super_admin: 'مدير عام', admin: 'مراجع', university: 'جامعة' }
  return map[auth.userRole] || ''
})

const menuItems = computed(() => {
  const items = [
    { title: 'لوحة التحكم', icon: 'mdi-view-dashboard', to: '/dashboard' },
    { title: 'التقديمات', icon: 'mdi-file-document-multiple', to: '/submissions' },
  ]
  if (auth.isUniversity) {
    items.push({ title: 'تقديم جديد', icon: 'mdi-plus-circle', to: '/submissions/new' })
  }
  if (auth.isAdmin) {
    items.push(
      { title: 'مراجعة التقديمات', icon: 'mdi-clipboard-check', to: '/admin/review' },
      { title: 'الجامعات', icon: 'mdi-school', to: '/universities' },
      { title: 'الإحصائيات', icon: 'mdi-chart-bar', to: '/stats' },
      { title: 'تحليل الذكاء الاصطناعي', icon: 'mdi-robot', to: '/ai-analysis' },
    )
  }
  if (auth.isSuperAdmin) {
    items.push(
      { title: 'إدارة المستخدمين', icon: 'mdi-account-group', to: '/admin/users' },
      { title: 'السنوات الدراسية', icon: 'mdi-calendar-range', to: '/admin/academic-years' },
      { title: 'إدارة المعايير', icon: 'mdi-format-list-checks', to: '/admin/categories' },
      { title: 'مركز الأمان', icon: 'mdi-shield-lock', to: '/admin/security' },
      { title: 'سجل العمليات', icon: 'mdi-history', to: '/admin/audit-logs' },
      { title: 'الإعدادات', icon: 'mdi-cog', to: '/admin/settings' },
    )
  }
  return items
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>
