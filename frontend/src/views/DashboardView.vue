<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">لوحة التحكم</h1>

    <!-- Admin Stats -->
    <v-row v-if="auth.isAdmin" class="mb-6">
      <v-col v-for="card in statCards" :key="card.title" cols="12" sm="6" md="3">
        <v-card class="pa-4" rounded="xl">
          <div class="d-flex align-center">
            <v-avatar :color="card.color" size="56" rounded="lg" class="ml-4">
              <v-icon :icon="card.icon" size="28" color="white" />
            </v-avatar>
            <div>
              <div class="text-h4 font-weight-bold">{{ card.value }}</div>
              <div class="text-body-2 text-medium-emphasis">{{ card.title }}</div>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- University Stats -->
    <v-row v-if="auth.isUniversity" class="mb-6">
      <v-col cols="12" md="4">
        <v-card class="pa-6 text-center" rounded="xl" color="primary" theme="dark">
          <v-icon icon="mdi-star" size="48" class="mb-2" />
          <div class="text-h3 font-weight-bold">{{ uniStats.latestScore || '-' }}</div>
          <div>آخر درجة</div>
        </v-card>
      </v-col>
      <v-col cols="12" md="4">
        <v-card class="pa-6 text-center" rounded="xl">
          <v-icon icon="mdi-file-document" size="48" class="mb-2" color="primary" />
          <div class="text-h3 font-weight-bold text-primary">{{ uniStats.totalSubmissions || 0 }}</div>
          <div>عدد التقديمات</div>
        </v-card>
      </v-col>
      <v-col cols="12" md="4">
        <v-card class="pa-6 text-center" rounded="xl">
          <v-icon icon="mdi-clock" size="48" class="mb-2" color="warning" />
          <div class="text-h3 font-weight-bold text-warning">{{ uniStats.latestStatus || '-' }}</div>
          <div>حالة آخر تقديم</div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Recent Submissions -->
    <v-card rounded="xl">
      <v-card-title class="d-flex align-center pa-4">
        <v-icon icon="mdi-history" class="ml-2" />
        آخر التقديمات
        <v-spacer />
        <v-btn color="primary" variant="tonal" to="/submissions" size="small">عرض الكل</v-btn>
      </v-card-title>
      <v-data-table :headers="headers" :items="recentSubmissions" :loading="loading" density="comfortable" hover>
        <template v-slot:item.status="{ item }">
          <submission-status-badge :status="item.status" />
        </template>
        <template v-slot:item.total_score="{ item }">
          <v-chip :color="item.total_score >= 70 ? 'success' : item.total_score >= 50 ? 'warning' : 'error'" size="small" v-if="item.total_score">
            {{ item.total_score }}
          </v-chip>
          <span v-else class="text-grey">-</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon variant="text" size="small" :to="`/submissions/${item.id}`">
            <v-icon>mdi-eye</v-icon>
          </v-btn>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'
import SubmissionStatusBadge from '../components/SubmissionStatusBadge.vue'

const auth = useAuthStore()
const loading = ref(false)
const stats = ref({})
const recentSubmissions = ref([])
const uniStats = ref({})

const statCards = computed(() => [
  { title: 'إجمالي الجامعات', value: stats.value.total_universities || 0, icon: 'mdi-school', color: 'primary' },
  { title: 'إجمالي التقديمات', value: stats.value.total_submissions || 0, icon: 'mdi-file-document-multiple', color: 'info' },
  { title: 'بانتظار المراجعة', value: stats.value.pending_reviews || 0, icon: 'mdi-clock-outline', color: 'warning' },
  { title: 'متوسط الدرجات', value: stats.value.average_score || 0, icon: 'mdi-chart-line', color: 'success' },
])

const headers = [
  { title: 'الجامعة', key: 'university_name', sortable: true },
  { title: 'السنة الدراسية', key: 'academic_year_name', sortable: true },
  { title: 'النسخة', key: 'version', sortable: true },
  { title: 'الحالة', key: 'status', sortable: true },
  { title: 'الدرجة', key: 'total_score', sortable: true },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

onMounted(async () => {
  loading.value = true
  try {
    if (auth.isAdmin) {
      const res = await api.getStatsOverview()
      stats.value = res.data
    }
    const subRes = await api.getSubmissions({ limit: 10 })
    recentSubmissions.value = subRes.data || []
    if (auth.isUniversity && recentSubmissions.value.length > 0) {
      const latest = recentSubmissions.value[0]
      uniStats.value = {
        latestScore: latest.total_score,
        totalSubmissions: recentSubmissions.value.length,
        latestStatus: statusMap[latest.status] || latest.status,
      }
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})

const statusMap = {
  draft: 'مسودة',
  submitted: 'مقدم',
  under_review: 'قيد المراجعة',
  approved: 'معتمد',
  rejected: 'مرفوض'
}
</script>
