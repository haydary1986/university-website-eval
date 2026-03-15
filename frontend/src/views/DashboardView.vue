<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">لوحة التحكم</h1>

    <v-skeleton-loader v-if="loading" type="card, card, card" />

    <template v-else>
      <!-- Admin Stats -->
      <template v-if="auth.isAdmin">
        <v-row class="mb-4">
          <v-col v-for="card in statCards" :key="card.title" cols="12" sm="6" md="3">
            <v-card class="pa-4" rounded="xl" :to="card.to">
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

        <!-- Status Distribution -->
        <v-row class="mb-4">
          <v-col cols="12" md="6">
            <v-card rounded="xl" class="pa-4">
              <div class="text-subtitle-1 font-weight-bold mb-3">توزيع حالات التقديمات</div>
              <div v-for="s in statusDistribution" :key="s.label" class="d-flex align-center mb-2">
                <v-chip :color="s.color" size="small" class="ml-3" style="min-width:80px">{{ s.label }}</v-chip>
                <v-progress-linear :model-value="s.pct" :color="s.color" height="20" rounded class="flex-grow-1">
                  <template v-slot:default>
                    <span class="text-caption font-weight-bold">{{ s.count }}</span>
                  </template>
                </v-progress-linear>
              </div>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card rounded="xl" class="pa-4">
              <div class="text-subtitle-1 font-weight-bold mb-3">مقارنة أنواع الجامعات</div>
              <v-row>
                <v-col cols="6" class="text-center">
                  <v-avatar color="primary" size="72" class="mb-2">
                    <span class="text-h5 text-white font-weight-bold">{{ stats.gov_count || 0 }}</span>
                  </v-avatar>
                  <div class="text-body-2">حكومية</div>
                </v-col>
                <v-col cols="6" class="text-center">
                  <v-avatar color="deep-purple" size="72" class="mb-2">
                    <span class="text-h5 text-white font-weight-bold">{{ stats.private_count || 0 }}</span>
                  </v-avatar>
                  <div class="text-body-2">أهلية</div>
                </v-col>
              </v-row>
              <div class="text-center mt-2">
                <v-chip color="success" variant="tonal" size="small">
                  متوسط الدرجات: {{ stats.average_score || 0 }}
                </v-chip>
                <v-chip color="info" variant="tonal" size="small" class="mr-2">
                  أعلى درجة: {{ stats.max_score || 0 }}
                </v-chip>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </template>

      <!-- University Stats -->
      <template v-if="auth.isUniversity">
        <v-row class="mb-6">
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
      </template>

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
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'
import SubmissionStatusBadge from '../components/SubmissionStatusBadge.vue'

const auth = useAuthStore()
const loading = ref(true)
const stats = ref({})
const recentSubmissions = ref([])
const uniStats = ref({})

const statCards = computed(() => [
  { title: 'إجمالي الجامعات', value: stats.value.total_universities || 0, icon: 'mdi-school', color: 'primary', to: '/universities' },
  { title: 'إجمالي التقديمات', value: stats.value.total_submissions || 0, icon: 'mdi-file-document-multiple', color: 'info', to: '/admin/review' },
  { title: 'بانتظار المراجعة', value: stats.value.pending_reviews || 0, icon: 'mdi-clock-outline', color: 'warning', to: '/admin/review' },
  { title: 'معتمدة', value: stats.value.approved_submissions || 0, icon: 'mdi-check-circle', color: 'success', to: '/stats' },
])

const statusDistribution = computed(() => {
  const total = stats.value.total_submissions || 1
  return [
    { label: 'مقدم', count: stats.value.submitted_count || 0, color: 'info', pct: ((stats.value.submitted_count || 0) / total) * 100 },
    { label: 'قيد المراجعة', count: stats.value.under_review_count || 0, color: 'warning', pct: ((stats.value.under_review_count || 0) / total) * 100 },
    { label: 'معتمد', count: stats.value.approved_submissions || 0, color: 'success', pct: ((stats.value.approved_submissions || 0) / total) * 100 },
    { label: 'مرفوض', count: stats.value.rejected_count || 0, color: 'error', pct: ((stats.value.rejected_count || 0) / total) * 100 },
  ]
})

const headers = [
  { title: 'الجامعة', key: 'university_name', sortable: true },
  { title: 'السنة الدراسية', key: 'academic_year_name', sortable: true },
  { title: 'النسخة', key: 'version', sortable: true },
  { title: 'الحالة', key: 'status', sortable: true },
  { title: 'الدرجة', key: 'total_score', sortable: true },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

onMounted(async () => {
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
