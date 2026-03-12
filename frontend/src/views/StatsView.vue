<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-chart-bar" class="ml-2" />
      الإحصائيات
    </h1>

    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row dense>
        <v-col cols="12" md="4">
          <v-select v-model="selectedYear" :items="academicYears" item-title="name" item-value="id" label="السنة الدراسية" clearable @update:model-value="loadStats" />
        </v-col>
      </v-row>
    </v-card>

    <!-- Stat Cards -->
    <v-row class="mb-6">
      <v-col v-for="card in statCards" :key="card.title" cols="12" sm="6" md="3">
        <v-card class="pa-4" rounded="xl">
          <div class="d-flex align-center">
            <v-avatar :color="card.color" size="48" rounded="lg" class="ml-3">
              <v-icon :icon="card.icon" color="white" />
            </v-avatar>
            <div>
              <div class="text-h5 font-weight-bold">{{ card.value }}</div>
              <div class="text-body-2 text-medium-emphasis">{{ card.title }}</div>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Charts -->
    <v-row>
      <v-col cols="12" md="6">
        <v-card class="pa-4" rounded="xl">
          <v-card-title>أعلى 10 جامعات</v-card-title>
          <v-card-text>
            <Bar v-if="topUniversitiesData" :data="topUniversitiesData" :options="barOptions" />
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card class="pa-4" rounded="xl">
          <v-card-title>التقديمات حسب الحالة</v-card-title>
          <v-card-text>
            <Pie v-if="statusData" :data="statusData" :options="pieOptions" />
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card class="pa-4" rounded="xl">
          <v-card-title>حكومية vs أهلية</v-card-title>
          <v-card-text>
            <Bar v-if="typeComparisonData" :data="typeComparisonData" :options="barOptions" />
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card class="pa-4" rounded="xl">
          <v-card-title>متوسط الدرجات حسب الفئة</v-card-title>
          <v-card-text>
            <Radar v-if="categoryData" :data="categoryData" :options="radarOptions" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Bar, Pie, Radar } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, BarElement, ArcElement, RadialLinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler } from 'chart.js'
import api from '../services/api'

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, RadialLinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const academicYears = ref([])
const selectedYear = ref(null)
const overview = ref({})
const uniStats = ref([])
const catStats = ref([])

const statCards = computed(() => [
  { title: 'إجمالي الجامعات', value: overview.value.total_universities || 0, icon: 'mdi-school', color: 'primary' },
  { title: 'إجمالي التقديمات', value: overview.value.total_submissions || 0, icon: 'mdi-file-document', color: 'info' },
  { title: 'متوسط الدرجات', value: Math.round(overview.value.average_score || 0), icon: 'mdi-chart-line', color: 'success' },
  { title: 'أعلى درجة', value: overview.value.max_score || 0, icon: 'mdi-trophy', color: 'secondary' },
])

const topUniversitiesData = computed(() => {
  const top = [...uniStats.value].sort((a, b) => b.score - a.score).slice(0, 10)
  if (!top.length) return null
  return {
    labels: top.map(u => u.name),
    datasets: [{ label: 'الدرجة', data: top.map(u => u.score), backgroundColor: '#1a237e' }]
  }
})

const statusData = computed(() => {
  const s = overview.value.status_counts
  if (!s) return null
  return {
    labels: ['مسودة', 'مقدم', 'قيد المراجعة', 'معتمد', 'مرفوض'],
    datasets: [{ data: [s.draft || 0, s.submitted || 0, s.under_review || 0, s.approved || 0, s.rejected || 0], backgroundColor: ['#9e9e9e', '#1976d2', '#f57c00', '#388e3c', '#d32f2f'] }]
  }
})

const typeComparisonData = computed(() => {
  const t = overview.value.type_comparison
  if (!t) return null
  return {
    labels: ['حكومية', 'أهلية'],
    datasets: [{ label: 'متوسط الدرجة', data: [t.government || 0, t.private || 0], backgroundColor: ['#1a237e', '#ffd600'] }]
  }
})

const categoryData = computed(() => {
  if (!catStats.value.length) return null
  return {
    labels: catStats.value.map(c => c.name),
    datasets: [{
      label: 'متوسط الدرجة %',
      data: catStats.value.map(c => c.average_percentage || 0),
      backgroundColor: 'rgba(26, 35, 126, 0.2)',
      borderColor: '#1a237e',
      pointBackgroundColor: '#1a237e',
    }]
  }
})

const barOptions = { responsive: true, plugins: { legend: { display: false } } }
const pieOptions = { responsive: true }
const radarOptions = { responsive: true, scales: { r: { beginAtZero: true, max: 100 } } }

async function loadStats() {
  try {
    const params = selectedYear.value ? { academic_year_id: selectedYear.value } : {}
    const [overviewRes, uniRes, catRes] = await Promise.all([
      api.getStatsOverview(params),
      api.getStatsUniversities(params),
      api.getStatsCategories(params),
    ])
    overview.value = overviewRes.data || {}
    uniStats.value = uniRes.data || []
    catStats.value = catRes.data || []
  } catch (e) {
    console.error(e)
  }
}

onMounted(async () => {
  const yearRes = await api.getAcademicYears()
  academicYears.value = yearRes.data || []
  loadStats()
})
</script>
