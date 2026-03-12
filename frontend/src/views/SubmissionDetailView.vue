<template>
  <div>
    <div class="d-flex align-center mb-6">
      <v-btn icon variant="text" @click="$router.back()" class="ml-2">
        <v-icon>mdi-arrow-right</v-icon>
      </v-btn>
      <h1 class="text-h4 font-weight-bold text-primary">تفاصيل التقديم</h1>
      <v-spacer />
      <submission-status-badge v-if="submission" :status="submission.status" class="ml-2" />
      <v-btn v-if="auth.isAdmin && submission?.version > 1" color="info" variant="tonal" :to="`/submissions/${submission.id}/diff`" prepend-icon="mdi-compare" class="mr-2">
        مقارنة النسخ
      </v-btn>
    </div>

    <v-skeleton-loader v-if="loading" type="card, card, card" />

    <template v-if="submission && !loading">
      <!-- Info Header -->
      <v-card class="mb-6 pa-4" rounded="xl">
        <v-row>
          <v-col cols="12" md="3">
            <div class="text-caption text-medium-emphasis">الجامعة</div>
            <div class="font-weight-bold">{{ submission.university_name }}</div>
          </v-col>
          <v-col cols="12" md="3">
            <div class="text-caption text-medium-emphasis">السنة الدراسية</div>
            <div class="font-weight-bold">{{ submission.academic_year_name }}</div>
          </v-col>
          <v-col cols="12" md="2">
            <div class="text-caption text-medium-emphasis">النسخة</div>
            <div class="font-weight-bold">{{ submission.version }}</div>
          </v-col>
          <v-col cols="12" md="2">
            <div class="text-caption text-medium-emphasis">الدرجة الكلية</div>
            <div class="text-h5 font-weight-bold" :class="scoreColor">{{ submission.total_score || '-' }} / 1000</div>
          </v-col>
          <v-col cols="12" md="2">
            <div class="text-caption text-medium-emphasis">المخول</div>
            <div>{{ submission.authorized_person }}</div>
            <div class="text-caption">{{ submission.authorized_phone }}</div>
          </v-col>
        </v-row>
      </v-card>

      <!-- Categories & Items -->
      <v-expansion-panels multiple variant="accordion">
        <v-expansion-panel v-for="cat in groupedItems" :key="cat.id" rounded="xl" class="mb-3">
          <v-expansion-panel-title>
            <div class="d-flex align-center w-100">
              <v-avatar color="primary" size="36" class="ml-3">
                <span class="text-white font-weight-bold">{{ cat.number }}</span>
              </v-avatar>
              <div class="flex-grow-1">
                <div class="font-weight-bold">{{ cat.name_ar }}</div>
              </div>
              <v-chip color="secondary" size="small" class="mr-2">
                {{ cat.score }} / {{ cat.weight }}
              </v-chip>
            </div>
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <v-card v-for="item in cat.items" :key="item.id" variant="outlined" class="mb-3 pa-4" rounded="lg">
              <div class="d-flex align-center mb-2">
                <div class="flex-grow-1 font-weight-medium">{{ item.criteria_name }}</div>
                <v-chip color="primary" variant="tonal" size="small">أقصى: {{ item.max_score }}</v-chip>
              </div>

              <div v-if="item.evidence" class="mb-2 pa-2 bg-grey-lighten-4 rounded">
                <div class="text-caption text-medium-emphasis">الدليل:</div>
                <div>{{ item.evidence }}</div>
              </div>

              <div v-if="item.evidence_file" class="mb-2">
                <v-chip size="small" prepend-icon="mdi-paperclip" color="info" variant="tonal">
                  ملف مرفق
                </v-chip>
              </div>

              <!-- Admin Scoring -->
              <div v-if="auth.isAdmin && canReview" class="mt-3 pa-3 bg-blue-lighten-5 rounded">
                <v-row dense>
                  <v-col cols="12" md="4">
                    <v-text-field v-model.number="scores[item.criteria_id]" type="number" :max="item.max_score" :min="0" label="الدرجة" density="compact" hide-details />
                  </v-col>
                  <v-col cols="12" md="8">
                    <v-text-field v-model="comments[item.criteria_id]" label="ملاحظة" density="compact" hide-details />
                  </v-col>
                </v-row>
              </div>

              <!-- Show existing score -->
              <div v-if="item.score !== undefined && item.score !== null && !canReview" class="mt-2">
                <v-chip :color="item.score >= item.max_score * 0.7 ? 'success' : 'warning'" size="small">
                  الدرجة: {{ item.score }} / {{ item.max_score }}
                </v-chip>
                <span v-if="item.admin_comment" class="text-caption mr-2">{{ item.admin_comment }}</span>
              </div>
            </v-card>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>

      <!-- Admin Actions -->
      <v-card v-if="auth.isAdmin && canReview" class="mt-6 pa-4" rounded="xl">
        <div class="d-flex justify-center ga-4">
          <v-btn color="success" size="large" @click="submitReview('approve')" :loading="reviewing" prepend-icon="mdi-check-circle">
            اعتماد
          </v-btn>
          <v-btn color="error" size="large" @click="submitReview('reject')" :loading="reviewing" prepend-icon="mdi-close-circle">
            رفض
          </v-btn>
          <v-btn color="info" size="large" @click="submitReview('save')" :loading="reviewing" prepend-icon="mdi-content-save">
            حفظ الدرجات
          </v-btn>
        </div>
      </v-card>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'
import SubmissionStatusBadge from '../components/SubmissionStatusBadge.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const reviewing = ref(false)
const submission = ref(null)
const categories = ref([])
const scores = reactive({})
const comments = reactive({})

const canReview = computed(() => {
  return submission.value && ['submitted', 'under_review'].includes(submission.value.status)
})

const scoreColor = computed(() => {
  const s = submission.value?.total_score
  if (!s) return ''
  if (s >= 700) return 'text-success'
  if (s >= 500) return 'text-warning'
  return 'text-error'
})

const groupedItems = computed(() => {
  if (!categories.value.length || !submission.value?.items) return []
  return categories.value.map(cat => {
    const catItems = (submission.value.items || []).filter(i => i.category_id === cat.id)
    const score = catItems.reduce((sum, i) => sum + (i.score || 0), 0)
    return { ...cat, items: catItems, score }
  }).filter(c => c.items.length > 0)
})

onMounted(async () => {
  loading.value = true
  try {
    const [subRes, catRes] = await Promise.all([
      auth.isAdmin ? api.getAdminSubmission(route.params.id) : api.getSubmission(route.params.id),
      api.getCategories()
    ])
    submission.value = subRes.data
    categories.value = catRes.data || []

    // Pre-fill scores
    if (submission.value.items) {
      submission.value.items.forEach(item => {
        if (item.score !== undefined) scores[item.criteria_id] = item.score
        if (item.admin_comment) comments[item.criteria_id] = item.admin_comment
      })
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})

async function submitReview(action) {
  reviewing.value = true
  try {
    const reviewItems = Object.entries(scores).map(([criteriaId, score]) => ({
      criteria_id: parseInt(criteriaId),
      score: score,
      admin_comment: comments[criteriaId] || '',
    }))

    await api.reviewSubmission(route.params.id, { items: reviewItems })

    if (action === 'approve') await api.approveSubmission(route.params.id)
    else if (action === 'reject') await api.rejectSubmission(route.params.id)

    router.push('/admin/review')
  } catch (e) {
    console.error(e)
  } finally {
    reviewing.value = false
  }
}
</script>
