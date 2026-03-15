<template>
  <div>
    <h1 class="text-h5 text-md-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-clipboard-check" class="ml-2" />
      مراجعة التقديمات
    </h1>

    <!-- Filters -->
    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row dense>
        <v-col cols="12" sm="6" md="3">
          <v-select v-model="statusFilter" :items="statusOptions" label="الحالة" clearable hide-details @update:model-value="loadData" />
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-select v-model="yearFilter" :items="academicYears" item-title="name" item-value="id" label="السنة الدراسية" clearable hide-details @update:model-value="loadData" />
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-select v-model="typeFilter" :items="[{title:'حكومية',value:'government'},{title:'أهلية',value:'private'}]" label="نوع الجامعة" clearable hide-details @update:model-value="filterLocal" />
        </v-col>
        <v-col cols="12" sm="6" md="3">
          <v-text-field v-model="searchQuery" label="بحث بالاسم..." prepend-inner-icon="mdi-magnify" clearable hide-details @update:model-value="filterLocal" />
        </v-col>
      </v-row>
    </v-card>

    <v-card rounded="xl">
      <v-data-table :headers="headers" :items="filteredSubmissions" :loading="loading" hover density="comfortable" :search="searchQuery" items-per-page="20">
        <template v-slot:item.university_name="{ item }">
          <div>
            {{ item.university_name }}
            <v-chip v-if="item.university_type" :color="item.university_type === 'government' ? 'primary' : 'secondary'" size="x-small" variant="tonal" class="mr-1">
              {{ item.university_type === 'government' ? 'حكومية' : 'أهلية' }}
            </v-chip>
          </div>
        </template>
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
import { ref, computed, onMounted } from 'vue'
import api from '../services/api'
import SubmissionStatusBadge from '../components/SubmissionStatusBadge.vue'

const loading = ref(false)
const submissions = ref([])
const academicYears = ref([])
const statusFilter = ref(null)
const yearFilter = ref(null)
const typeFilter = ref(null)
const searchQuery = ref('')

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

const filteredSubmissions = computed(() => {
  let items = submissions.value
  if (typeFilter.value) {
    items = items.filter(s => s.university_type === typeFilter.value)
  }
  return items
})

function filterLocal() {
  // triggered by local filters; computed handles it
}

async function loadData() {
  loading.value = true
  try {
    const params = {}
    if (statusFilter.value) params.status = statusFilter.value
    if (yearFilter.value) params.academic_year_id = yearFilter.value
    const res = await api.getAdminSubmissions(params)
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
