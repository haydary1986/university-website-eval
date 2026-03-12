<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-clipboard-check" class="ml-2" />
      مراجعة التقديمات
    </h1>

    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row dense>
        <v-col cols="12" md="4">
          <v-select v-model="statusFilter" :items="statusOptions" label="الحالة" clearable hide-details @update:model-value="loadData" />
        </v-col>
      </v-row>
    </v-card>

    <v-card rounded="xl">
      <v-data-table :headers="headers" :items="submissions" :loading="loading" hover density="comfortable">
        <template v-slot:item.status="{ item }">
          <submission-status-badge :status="item.status" />
        </template>
        <template v-slot:item.total_score="{ item }">
          <v-chip v-if="item.total_score" :color="item.total_score >= 70 ? 'success' : 'warning'" size="small">{{ item.total_score }}</v-chip>
          <span v-else class="text-grey">-</span>
        </template>
        <template v-slot:item.created_at="{ item }">
          {{ new Date(item.created_at).toLocaleDateString('ar-IQ') }}
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn color="primary" size="small" variant="tonal" :to="`/submissions/${item.id}`" prepend-icon="mdi-eye">
            مراجعة
          </v-btn>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../services/api'
import SubmissionStatusBadge from '../components/SubmissionStatusBadge.vue'

const loading = ref(false)
const submissions = ref([])
const statusFilter = ref(null)

const statusOptions = [
  { title: 'مقدم', value: 'submitted' },
  { title: 'قيد المراجعة', value: 'under_review' },
  { title: 'معتمد', value: 'approved' },
  { title: 'مرفوض', value: 'rejected' },
]

const headers = [
  { title: 'الجامعة', key: 'university_name', sortable: true },
  { title: 'السنة الدراسية', key: 'academic_year_name', sortable: true },
  { title: 'النسخة', key: 'version' },
  { title: 'الحالة', key: 'status', sortable: true },
  { title: 'الدرجة', key: 'total_score', sortable: true },
  { title: 'التاريخ', key: 'created_at', sortable: true },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

async function loadData() {
  loading.value = true
  try {
    const params = {}
    if (statusFilter.value) params.status = statusFilter.value
    const res = await api.getAdminSubmissions(params)
    submissions.value = res.data?.submissions || res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
