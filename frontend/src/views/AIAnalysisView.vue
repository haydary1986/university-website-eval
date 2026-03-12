<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-robot" class="ml-2" />
      تحليل الذكاء الاصطناعي
    </h1>

    <v-card class="mb-6 pa-4" rounded="xl">
      <v-row>
        <v-col cols="12" md="5">
          <v-autocomplete v-model="selectedSubmission" :items="submissions" :item-title="submissionLabel" item-value="id" label="اختر التقديم" prepend-inner-icon="mdi-file-document" clearable />
        </v-col>
        <v-col cols="12" md="3">
          <v-radio-group v-model="provider" inline>
            <v-radio label="DeepSeek" value="deepseek" />
            <v-radio label="Gemini" value="gemini" />
          </v-radio-group>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" class="d-flex ga-3">
          <v-btn color="primary" @click="analyze" :loading="analyzing" :disabled="!selectedSubmission" prepend-icon="mdi-magnify-scan">
            تحليل التقديم
          </v-btn>
          <v-btn color="info" @click="suggest" :loading="suggesting" :disabled="!selectedSubmission" prepend-icon="mdi-lightbulb">
            اقتراح تحسينات
          </v-btn>
          <v-btn color="secondary" @click="compare" :loading="comparing" prepend-icon="mdi-compare">
            مقارنة الجامعات
          </v-btn>
        </v-col>
      </v-row>
    </v-card>

    <v-card v-if="result" class="pa-6" rounded="xl">
      <v-card-title class="d-flex align-center">
        <v-icon icon="mdi-robot" class="ml-2" color="primary" />
        نتائج التحليل
        <v-spacer />
        <v-chip size="small" color="info">{{ provider === 'deepseek' ? 'DeepSeek' : 'Gemini' }}</v-chip>
      </v-card-title>
      <v-card-text>
        <div class="ai-result" style="white-space: pre-wrap; line-height: 1.8;" v-html="formattedResult"></div>
      </v-card-text>
    </v-card>

    <v-overlay v-model="loadingOverlay" class="align-center justify-center" persistent>
      <v-card class="pa-6 text-center" rounded="xl" width="300">
        <v-progress-circular indeterminate color="primary" size="64" class="mb-4" />
        <div class="text-h6">جاري التحليل...</div>
        <div class="text-body-2 text-medium-emphasis">يرجى الانتظار</div>
      </v-card>
    </v-overlay>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../services/api'

const submissions = ref([])
const selectedSubmission = ref(null)
const provider = ref('deepseek')
const result = ref('')
const analyzing = ref(false)
const suggesting = ref(false)
const comparing = ref(false)
const loadingOverlay = ref(false)

function submissionLabel(item) {
  return `${item.university_name} - ${item.academic_year_name} (v${item.version})`
}

const formattedResult = computed(() => {
  if (!result.value) return ''
  return result.value
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br>')
})

async function analyze() {
  analyzing.value = true
  loadingOverlay.value = true
  try {
    const res = await api.analyzeSubmission(selectedSubmission.value, provider.value)
    result.value = res.data.analysis || res.data.result || JSON.stringify(res.data)
  } catch (e) {
    result.value = 'حدث خطأ أثناء التحليل: ' + (e.response?.data?.error || e.message)
  } finally {
    analyzing.value = false
    loadingOverlay.value = false
  }
}

async function suggest() {
  suggesting.value = true
  loadingOverlay.value = true
  try {
    const res = await api.suggestImprovements(selectedSubmission.value, provider.value)
    result.value = res.data.suggestions || res.data.result || JSON.stringify(res.data)
  } catch (e) {
    result.value = 'حدث خطأ: ' + (e.response?.data?.error || e.message)
  } finally {
    suggesting.value = false
    loadingOverlay.value = false
  }
}

async function compare() {
  comparing.value = true
  loadingOverlay.value = true
  try {
    const ids = submissions.value.slice(0, 10).map(s => s.id)
    const res = await api.compareUniversities(ids, provider.value)
    result.value = res.data.comparison || res.data.result || JSON.stringify(res.data)
  } catch (e) {
    result.value = 'حدث خطأ: ' + (e.response?.data?.error || e.message)
  } finally {
    comparing.value = false
    loadingOverlay.value = false
  }
}

onMounted(async () => {
  try {
    const res = await api.getSubmissions({ status: 'approved' })
    submissions.value = res.data?.submissions || res.data || []
    if (!submissions.value.length) {
      const allRes = await api.getSubmissions({})
      submissions.value = allRes.data?.submissions || allRes.data || []
    }
  } catch (e) {
    console.error(e)
  }
})
</script>
