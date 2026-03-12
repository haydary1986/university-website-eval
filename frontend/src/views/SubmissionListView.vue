<template>
  <div>
    <div class="d-flex align-center mb-6">
      <h1 class="text-h4 font-weight-bold text-primary">
        <v-icon icon="mdi-file-document-multiple" class="ml-2" />
        التقديمات
      </h1>
      <v-spacer />
      <v-btn v-if="auth.isUniversity" color="primary" to="/submissions/new" prepend-icon="mdi-plus">تقديم جديد</v-btn>
    </div>

    <!-- Filters -->
    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row dense>
        <v-col cols="12" md="3">
          <v-text-field v-model="filters.search" label="بحث بالجامعة" prepend-inner-icon="mdi-magnify" clearable hide-details @input="loadData" />
        </v-col>
        <v-col cols="12" md="3">
          <v-select v-model="filters.academic_year_id" :items="academicYears" item-title="name" item-value="id" label="السنة الدراسية" clearable hide-details @update:model-value="loadData" />
        </v-col>
        <v-col cols="12" md="3">
          <v-select v-model="filters.status" :items="statusOptions" label="الحالة" clearable hide-details @update:model-value="loadData" />
        </v-col>
        <v-col cols="12" md="3">
          <v-select v-model="filters.type" :items="typeOptions" label="النوع" clearable hide-details @update:model-value="loadData" />
        </v-col>
      </v-row>
    </v-card>

    <v-card rounded="xl">
      <v-data-table :headers="headers" :items="submissions" :loading="loading" hover density="comfortable" :items-per-page="15">
        <template v-slot:item.status="{ item }">
          <submission-status-badge :status="item.status" />
        </template>
        <template v-slot:item.total_score="{ item }">
          <v-chip v-if="item.total_score" :color="item.total_score >= 70 ? 'success' : item.total_score >= 50 ? 'warning' : 'error'" size="small">
            {{ item.total_score }}
          </v-chip>
          <span v-else class="text-grey">-</span>
        </template>
        <template v-slot:item.created_at="{ item }">
          {{ new Date(item.created_at).toLocaleDateString('ar-IQ') }}
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon variant="text" size="small" :to="`/submissions/${item.id}`">
            <v-icon>mdi-eye</v-icon>
            <v-tooltip activator="parent">عرض</v-tooltip>
          </v-btn>
          <v-btn v-if="auth.isAdmin && item.version > 1" icon variant="text" size="small" :to="`/submissions/${item.id}/diff`">
            <v-icon>mdi-compare</v-icon>
            <v-tooltip activator="parent">مقارنة النسخ</v-tooltip>
          </v-btn>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'
import SubmissionStatusBadge from '../components/SubmissionStatusBadge.vue'

const auth = useAuthStore()
const loading = ref(false)
const submissions = ref([])
const academicYears = ref([])

const filters = reactive({ search: '', academic_year_id: null, status: null, type: null })

const statusOptions = [
  { title: 'مسودة', value: 'draft' },
  { title: 'مقدم', value: 'submitted' },
  { title: 'قيد المراجعة', value: 'under_review' },
  { title: 'معتمد', value: 'approved' },
  { title: 'مرفوض', value: 'rejected' },
]

const typeOptions = [
  { title: 'حكومية', value: 'government' },
  { title: 'أهلية', value: 'private' },
]

const headers = [
  { title: 'الجامعة', key: 'university_name', sortable: true },
  { title: 'السنة الدراسية', key: 'academic_year_name', sortable: true },
  { title: 'النسخة', key: 'version', sortable: true },
  { title: 'الحالة', key: 'status', sortable: true },
  { title: 'الدرجة', key: 'total_score', sortable: true },
  { title: 'التاريخ', key: 'created_at', sortable: true },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

async function loadData() {
  loading.value = true
  try {
    const params = {}
    if (filters.search) params.search = filters.search
    if (filters.academic_year_id) params.academic_year_id = filters.academic_year_id
    if (filters.status) params.status = filters.status
    if (filters.type) params.type = filters.type
    const res = await api.getSubmissions(params)
    submissions.value = res.data?.submissions || res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const yearRes = await api.getAcademicYears()
  academicYears.value = yearRes.data || []
  loadData()
})
</script>
