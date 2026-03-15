<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-chart-bar" class="ml-2" />
      الإحصائيات والتصنيفات
    </h1>

    <!-- Filters -->
    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row dense>
        <v-col cols="12" md="3">
          <v-select v-model="selectedYear" :items="academicYears" item-title="name" item-value="id" label="السنة الدراسية" clearable @update:model-value="loadAll" />
        </v-col>
        <v-col cols="12" md="3">
          <v-select v-model="selectedType" :items="typeOptions" label="نوع الجامعة" clearable @update:model-value="loadAll" />
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

    <!-- Tabs -->
    <v-card rounded="xl">
      <div class="d-flex align-center pa-3">
        <v-spacer />
        <v-menu>
          <template v-slot:activator="{ props }">
            <v-btn v-bind="props" color="success" variant="tonal" size="small" prepend-icon="mdi-download">
              تصدير CSV
            </v-btn>
          </template>
          <v-list density="compact">
            <v-list-item @click="exportData('rankings')">
              <v-list-item-title>تصدير التصنيف العام</v-list-item-title>
            </v-list-item>
            <v-list-item @click="exportData('category-rankings')">
              <v-list-item-title>تصدير تصنيف الفقرات</v-list-item-title>
            </v-list-item>
            <v-list-item @click="exportData('submissions')">
              <v-list-item-title>تصدير التقديمات</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

      <v-tabs v-model="activeTab" color="primary" grow>
        <v-tab value="overall"><v-icon class="ml-1">mdi-trophy</v-icon> التصنيف العام</v-tab>
        <v-tab value="categories"><v-icon class="ml-1">mdi-format-list-numbered</v-icon> تصنيف حسب الفقرات</v-tab>
        <v-tab value="profile"><v-icon class="ml-1">mdi-school</v-icon> ملف جامعة</v-tab>
        <v-tab value="charts"><v-icon class="ml-1">mdi-chart-line</v-icon> الرسوم البيانية</v-tab>
      </v-tabs>

      <v-card-text>
        <!-- Overall Rankings -->
        <div v-if="activeTab === 'overall'">
          <v-data-table :headers="overallHeaders" :items="rankings" :loading="loading" items-per-page="20" hover>
            <template #item.rank="{ index }">
              <v-chip v-if="index < 3" :color="['amber', 'grey-lighten-1', 'brown-lighten-1'][index]" size="small" variant="elevated" class="font-weight-bold">
                {{ index + 1 }}
                <v-icon size="14" class="mr-1">mdi-medal</v-icon>
              </v-chip>
              <span v-else class="font-weight-bold">{{ index + 1 }}</span>
            </template>
            <template #item.university_type="{ item }">
              <v-chip :color="item.university_type === 'government' ? 'primary' : 'secondary'" size="small" variant="tonal">
                {{ item.university_type === 'government' ? 'حكومية' : 'أهلية' }}
              </v-chip>
            </template>
            <template #item.total_score="{ item }">
              <div class="d-flex align-center">
                <v-progress-linear :model-value="item.total_score / 10" :color="scoreColor(item.total_score)" height="8" rounded class="ml-2" style="max-width:100px" />
                <strong>{{ item.total_score }}</strong>
              </div>
            </template>
            <template #item.actions="{ item }">
              <v-btn size="small" variant="tonal" color="primary" @click="showProfile(item.university_id)">
                <v-icon size="small" class="ml-1">mdi-eye</v-icon> الملف
              </v-btn>
            </template>
          </v-data-table>
        </div>

        <!-- Category Rankings -->
        <div v-if="activeTab === 'categories'">
          <v-expansion-panels v-model="openPanels" multiple>
            <v-expansion-panel v-for="cat in categoryRankings" :key="cat.category_id">
              <v-expansion-panel-title>
                <div class="d-flex align-center w-100">
                  <v-chip color="primary" size="small" variant="elevated" class="ml-3">{{ cat.category_id }}</v-chip>
                  <strong class="ml-3">{{ cat.category_name }}</strong>
                  <v-spacer />
                  <v-chip size="small" variant="tonal" color="info" class="ml-2">الوزن: {{ cat.category_weight }}</v-chip>
                  <v-chip size="small" variant="tonal">{{ cat.universities?.length || 0 }} جامعة</v-chip>
                </div>
              </v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-table density="compact" hover>
                  <thead>
                    <tr>
                      <th width="60">الترتيب</th>
                      <th>الجامعة</th>
                      <th width="80">النوع</th>
                      <th width="120">الدرجة</th>
                      <th width="200">النسبة</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="uni in cat.universities" :key="uni.university_id" :class="{ 'bg-amber-lighten-5': uni.rank <= 3 }">
                      <td>
                        <v-chip v-if="uni.rank <= 3" :color="['amber', 'grey-lighten-1', 'brown-lighten-1'][uni.rank - 1]" size="x-small" variant="elevated">
                          {{ uni.rank }}
                          <v-icon size="12" class="mr-1">mdi-medal</v-icon>
                        </v-chip>
                        <span v-else>{{ uni.rank }}</span>
                      </td>
                      <td>
                        <a href="#" @click.prevent="showProfile(uni.university_id)" class="text-primary text-decoration-none font-weight-medium">
                          {{ uni.university_name }}
                        </a>
                      </td>
                      <td>
                        <v-chip :color="uni.university_type === 'government' ? 'primary' : 'secondary'" size="x-small" variant="tonal">
                          {{ uni.university_type === 'government' ? 'حكومية' : 'أهلية' }}
                        </v-chip>
                      </td>
                      <td><strong>{{ uni.score }}</strong> / {{ cat.category_weight }}</td>
                      <td>
                        <div class="d-flex align-center">
                          <v-progress-linear :model-value="uni.percentage" :color="scoreColor(uni.percentage * 10)" height="8" rounded class="ml-2" />
                          <span class="text-caption">{{ Math.round(uni.percentage) }}%</span>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </v-table>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </div>

        <!-- University Profile -->
        <div v-if="activeTab === 'profile'">
          <v-autocomplete v-model="selectedUniId" :items="allUniversities" item-title="name" item-value="id" label="اختر الجامعة" prepend-inner-icon="mdi-school" class="mb-4" @update:model-value="loadProfile" />

          <div v-if="profile">
            <!-- University Header -->
            <v-card class="mb-4 pa-4" color="primary" theme="dark" rounded="xl">
              <v-row align="center">
                <v-col cols="12" md="6">
                  <h2 class="text-h5 font-weight-bold">{{ profile.university?.name }}</h2>
                  <v-chip :color="profile.university?.type === 'government' ? 'white' : 'amber'" variant="tonal" size="small" class="mt-1">
                    {{ profile.university?.type === 'government' ? 'حكومية' : 'أهلية' }}
                  </v-chip>
                </v-col>
                <v-col cols="4" md="2" class="text-center">
                  <div class="text-h3 font-weight-bold">{{ profile.overall_rank || '-' }}</div>
                  <div class="text-caption">الترتيب العام</div>
                </v-col>
                <v-col cols="4" md="2" class="text-center">
                  <div class="text-h3 font-weight-bold">{{ profile.total_ranked || '-' }}</div>
                  <div class="text-caption">إجمالي المصنفين</div>
                </v-col>
                <v-col cols="4" md="2" class="text-center">
                  <div class="text-h3 font-weight-bold">{{ latestScore }}</div>
                  <div class="text-caption">آخر درجة</div>
                </v-col>
              </v-row>
            </v-card>

            <!-- Category Performance -->
            <v-card class="mb-4" rounded="xl">
              <v-card-title>الأداء حسب الفقرات</v-card-title>
              <v-card-text>
                <v-table density="comfortable" hover>
                  <thead>
                    <tr>
                      <th>الفقرة</th>
                      <th width="120">الدرجة</th>
                      <th width="80">الوزن</th>
                      <th width="200">النسبة</th>
                      <th width="100">الترتيب</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="cat in profileCategories" :key="cat.category_id">
                      <td class="font-weight-medium">{{ cat.category_name }}</td>
                      <td><strong>{{ cat.score }}</strong></td>
                      <td>{{ cat.weight }}</td>
                      <td>
                        <div class="d-flex align-center">
                          <v-progress-linear :model-value="cat.percentage" :color="scoreColor(cat.percentage * 10)" height="10" rounded class="ml-2" />
                          <span class="font-weight-bold">{{ Math.round(cat.percentage) }}%</span>
                        </div>
                      </td>
                      <td>
                        <v-chip v-if="getRank(cat.category_id)" :color="getRank(cat.category_id).rank <= 3 ? 'amber' : getRank(cat.category_id).rank <= 10 ? 'info' : 'grey'" size="small" variant="tonal">
                          {{ getRank(cat.category_id).rank }} / {{ getRank(cat.category_id).total_in_rank }}
                        </v-chip>
                      </td>
                    </tr>
                  </tbody>
                </v-table>
              </v-card-text>
            </v-card>

            <!-- Radar Chart -->
            <v-card class="mb-4" rounded="xl" v-if="profileRadarData">
              <v-card-title>خريطة الأداء</v-card-title>
              <v-card-text style="max-height: 400px">
                <Radar :data="profileRadarData" :options="radarOptions" />
              </v-card-text>
            </v-card>

            <!-- Year over Year -->
            <v-card rounded="xl" v-if="profile.yearly_scores?.length > 1">
              <v-card-title>التطور عبر السنوات</v-card-title>
              <v-card-text>
                <Bar :data="yearlyData" :options="barOptions" v-if="yearlyData" />
              </v-card-text>
            </v-card>
          </div>
        </div>

        <!-- Charts Tab -->
        <div v-if="activeTab === 'charts'">
          <v-row>
            <v-col cols="12" md="6">
              <v-card class="pa-4" rounded="xl">
                <v-card-title>أعلى 10 جامعات</v-card-title>
                <v-card-text>
                  <Bar v-if="topUniversitiesData" :data="topUniversitiesData" :options="horizontalBarOptions" />
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="12" md="6">
              <v-card class="pa-4" rounded="xl">
                <v-card-title>حكومية مقابل أهلية</v-card-title>
                <v-card-text>
                  <Pie v-if="typeData" :data="typeData" :options="pieOptions" />
                </v-card-text>
              </v-card>
            </v-col>
            <v-col cols="12" md="6">
              <v-card class="pa-4" rounded="xl">
                <v-card-title>متوسط الدرجات حسب الفقرة</v-card-title>
                <v-card-text>
                  <Radar v-if="avgCategoryData" :data="avgCategoryData" :options="radarOptions" />
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
          </v-row>
        </div>
      </v-card-text>
    </v-card>
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
const selectedType = ref(null)
const activeTab = ref('overall')
const loading = ref(false)
const openPanels = ref([])

const overview = ref({})
const rankings = ref([])
const catStats = ref([])
const categoryRankings = ref([])
const allUniversities = ref([])
const selectedUniId = ref(null)
const profile = ref(null)

const typeOptions = [
  { title: 'حكومية', value: 'government' },
  { title: 'أهلية', value: 'private' },
]

const overallHeaders = [
  { title: '#', key: 'rank', width: '60px', sortable: false },
  { title: 'الجامعة', key: 'university_name' },
  { title: 'النوع', key: 'university_type', width: '100px' },
  { title: 'الدرجة', key: 'total_score', width: '200px' },
  { title: 'السنة', key: 'academic_year', width: '120px' },
  { title: '', key: 'actions', width: '100px', sortable: false },
]

const statCards = computed(() => [
  { title: 'إجمالي الجامعات', value: overview.value.total_universities || 0, icon: 'mdi-school', color: 'primary' },
  { title: 'تقديمات معتمدة', value: overview.value.approved_submissions || 0, icon: 'mdi-check-circle', color: 'success' },
  { title: 'متوسط الدرجات', value: Math.round(overview.value.average_score || 0), icon: 'mdi-chart-line', color: 'info' },
  { title: 'أعلى درجة', value: Math.round(overview.value.max_score || 0), icon: 'mdi-trophy', color: 'amber-darken-2' },
])

const latestScore = computed(() => {
  if (!profile.value?.submissions?.length) return '-'
  return profile.value.submissions[0].total_score
})

const profileCategories = computed(() => profile.value?.category_scores || [])

function getRank(categoryId) {
  return profile.value?.category_ranks?.find(r => r.category_id === categoryId)
}

function scoreColor(score) {
  if (score >= 700) return 'success'
  if (score >= 500) return 'info'
  if (score >= 300) return 'warning'
  return 'error'
}

function getParams() {
  const p = {}
  if (selectedYear.value) p.academic_year_id = selectedYear.value
  if (selectedType.value) p.type = selectedType.value
  return p
}

// Chart data
const topUniversitiesData = computed(() => {
  const top = [...rankings.value].slice(0, 10)
  if (!top.length) return null
  return {
    labels: top.map(u => u.university_name),
    datasets: [{
      label: 'الدرجة',
      data: top.map(u => u.total_score),
      backgroundColor: top.map(u => u.university_type === 'government' ? '#1a237e' : '#ffd600'),
    }]
  }
})

const typeData = computed(() => {
  const tc = overview.value.type_comparison
  if (!tc) return null
  return {
    labels: ['حكومية', 'أهلية'],
    datasets: [{ data: [Math.round(tc.government || 0), Math.round(tc.private || 0)], backgroundColor: ['#1a237e', '#ffd600'] }]
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

const avgCategoryData = computed(() => {
  if (!catStats.value.length) return null
  return {
    labels: catStats.value.map(c => c.category_name),
    datasets: [{
      label: 'متوسط الدرجة %',
      data: catStats.value.map(c => c.max_possible > 0 ? Math.round(c.avg_score * 100 / c.max_possible) : 0),
      backgroundColor: 'rgba(26, 35, 126, 0.2)',
      borderColor: '#1a237e',
      pointBackgroundColor: '#1a237e',
    }]
  }
})

const profileRadarData = computed(() => {
  if (!profileCategories.value.length) return null
  return {
    labels: profileCategories.value.map(c => c.category_name),
    datasets: [{
      label: 'النسبة %',
      data: profileCategories.value.map(c => Math.round(c.percentage)),
      backgroundColor: 'rgba(26, 35, 126, 0.2)',
      borderColor: '#1a237e',
      pointBackgroundColor: '#1a237e',
    }]
  }
})

const yearlyData = computed(() => {
  if (!profile.value?.yearly_scores?.length) return null
  return {
    labels: profile.value.yearly_scores.map(y => y.academic_year),
    datasets: [{
      label: 'الدرجة الكلية',
      data: profile.value.yearly_scores.map(y => y.total_score),
      backgroundColor: '#1a237e',
    }]
  }
})

const horizontalBarOptions = {
  indexAxis: 'y',
  responsive: true,
  plugins: { legend: { display: false } },
  scales: { x: { beginAtZero: true } }
}
const barOptions = { responsive: true, plugins: { legend: { display: false } } }
const pieOptions = { responsive: true }
const radarOptions = { responsive: true, scales: { r: { beginAtZero: true, max: 100 } } }

async function loadAll() {
  loading.value = true
  try {
    const params = getParams()
    const [overviewRes, rankingsRes, catStatsRes, catRankingsRes] = await Promise.all([
      api.getStatsOverview(params),
      api.getStatsUniversities(params),
      api.getStatsCategories(params),
      api.getCategoryRankings(params),
    ])
    overview.value = overviewRes.data || {}
    rankings.value = rankingsRes.data || []
    catStats.value = catStatsRes.data || []
    categoryRankings.value = catRankingsRes.data?.category_rankings || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function showProfile(uniId) {
  selectedUniId.value = uniId
  activeTab.value = 'profile'
  loadProfile()
}

async function loadProfile() {
  if (!selectedUniId.value) { profile.value = null; return }
  try {
    const params = selectedYear.value ? { academic_year_id: selectedYear.value } : {}
    const res = await api.getUniversityProfile(selectedUniId.value, params)
    profile.value = res.data
  } catch (e) {
    console.error(e)
    profile.value = null
  }
}

async function exportData(type_) {
  const params = {}
  if (selectedYear.value) params.academic_year_id = selectedYear.value

  try {
    let res
    if (type_ === 'rankings') res = await api.exportRankings(params)
    else if (type_ === 'category-rankings') res = await api.exportCategoryRankings(params)
    else res = await api.exportSubmissions(params)

    const url = window.URL.createObjectURL(new Blob([res.data]))
    const a = document.createElement('a')
    a.href = url
    a.download = `${type_}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch (e) {
    console.error('Export failed', e)
  }
}

onMounted(async () => {
  const [yearRes, uniRes] = await Promise.all([
    api.getAcademicYears(),
    api.getUniversities(),
  ])
  academicYears.value = yearRes.data || []
  allUniversities.value = uniRes.data || []
  loadAll()
})
</script>
