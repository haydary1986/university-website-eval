<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-file-plus" class="ml-2" />
      {{ isEdit ? 'تعديل التقديم' : 'تقديم جديد' }}
    </h1>

    <v-alert v-if="existingVersions > 0" type="info" variant="tonal" class="mb-4" icon="mdi-information">
      هذا التقديم هو النسخة رقم {{ existingVersions + 1 }}. سيتم حفظه كنسخة جديدة مع إمكانية مقارنة التغييرات.
    </v-alert>

    <v-form ref="form" @submit.prevent="handleSubmit">
      <!-- University & Contact Info -->
      <v-card class="mb-6 pa-4" rounded="xl">
        <v-card-title class="text-h6 text-primary">
          <v-icon icon="mdi-account-tie" class="ml-2" />
          معلومات المخول
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12" md="4">
              <v-select v-model="formData.academic_year_id" :items="academicYears" item-title="name" item-value="id" label="السنة الدراسية" :rules="[v => !!v || 'مطلوب']" prepend-inner-icon="mdi-calendar" />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field v-model="formData.authorized_person" label="اسم المخول" :rules="[v => !!v || 'مطلوب']" prepend-inner-icon="mdi-account" />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field v-model="formData.authorized_phone" label="رقم الهاتف" prepend-inner-icon="mdi-phone" />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field v-model="formData.authorized_email" label="البريد الالكتروني" type="email" prepend-inner-icon="mdi-email" />
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>

      <!-- Categories & Criteria -->
      <v-expansion-panels v-model="openPanels" multiple variant="accordion">
        <v-expansion-panel v-for="cat in categories" :key="cat.id" rounded="xl" class="mb-3">
          <v-expansion-panel-title>
            <div class="d-flex align-center w-100">
              <v-avatar color="primary" size="36" class="ml-3">
                <span class="text-white font-weight-bold">{{ cat.number }}</span>
              </v-avatar>
              <div class="flex-grow-1">
                <div class="font-weight-bold">{{ cat.name_ar }}</div>
                <div class="text-caption text-medium-emphasis">الوزن: {{ cat.weight }} درجة</div>
              </div>
              <v-chip color="secondary" size="small" class="mr-2">
                {{ getCategoryFilledCount(cat) }} / {{ cat.criteria?.length || 0 }}
              </v-chip>
            </div>
          </v-expansion-panel-title>
          <v-expansion-panel-text>
            <v-card v-for="cr in cat.criteria" :key="cr.id" variant="outlined" class="mb-3 pa-4" rounded="lg">
              <div class="d-flex align-center mb-3">
                <div class="flex-grow-1">
                  <div class="font-weight-medium">{{ cr.name_ar }}</div>
                  <div class="text-caption text-medium-emphasis mt-1" v-if="cr.description" style="white-space: pre-line;">{{ cr.description }}</div>
                </div>
                <v-chip color="primary" variant="tonal" size="small">أقصى درجة: {{ cr.max_score }}</v-chip>
              </div>
              <v-textarea v-model="getItem(cr.id).evidence" label="الدليل (رابط أو وصف)" rows="2" auto-grow hide-details class="mb-2" />
              <v-file-input v-model="getItem(cr.id).file" label="ارفاق ملف (اختياري)" prepend-icon="mdi-paperclip" hide-details density="compact" />
            </v-card>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>

      <!-- Actions -->
      <v-card class="mt-6 pa-4" rounded="xl">
        <div class="d-flex justify-center ga-4">
          <v-btn color="grey" variant="tonal" size="large" @click="saveDraft" :loading="saving" prepend-icon="mdi-content-save">
            حفظ كمسودة
          </v-btn>
          <v-btn color="primary" size="large" @click="handleSubmit" :loading="submitting" prepend-icon="mdi-send">
            تقديم للمراجعة
          </v-btn>
        </div>
      </v-card>
    </v-form>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const route = useRoute()
const form = ref(null)
const categories = ref([])
const academicYears = ref([])
const openPanels = ref([])
const saving = ref(false)
const submitting = ref(false)
const existingVersions = ref(0)
const submissionId = ref(null)
const isEdit = ref(false)

const formData = reactive({
  academic_year_id: null,
  authorized_person: '',
  authorized_phone: '',
  authorized_email: '',
})

const items = reactive({})

function getItem(criteriaId) {
  if (!items[criteriaId]) {
    items[criteriaId] = reactive({ evidence: '', file: null })
  }
  return items[criteriaId]
}

function getCategoryFilledCount(cat) {
  if (!cat.criteria) return 0
  return cat.criteria.filter(cr => items[cr.id]?.evidence).length
}

onMounted(async () => {
  try {
    const [catRes, yearRes] = await Promise.all([
      api.getCategories(),
      api.getAcademicYears()
    ])
    categories.value = catRes.data?.categories || catRes.data || []
    academicYears.value = yearRes.data?.academic_years || yearRes.data || []

    // Set active year as default
    const activeYear = academicYears.value.find(y => y.is_active)
    if (activeYear) formData.academic_year_id = activeYear.id

    // Check if editing existing
    if (route.query.edit) {
      isEdit.value = true
      submissionId.value = route.query.edit
      const subRes = await api.getSubmission(route.query.edit)
      const sub = subRes.data
      formData.academic_year_id = sub.academic_year_id
      formData.authorized_person = sub.authorized_person
      formData.authorized_phone = sub.authorized_phone
      formData.authorized_email = sub.authorized_email
      if (sub.items) {
        sub.items.forEach(item => {
          items[item.criteria_id] = reactive({ evidence: item.evidence || '', file: null })
        })
      }
    }

    // Check existing versions
    const subRes = await api.getSubmissions({ academic_year_id: formData.academic_year_id })
    const subs = subRes.data?.submissions || subRes.data || []
    existingVersions.value = subs.length
  } catch (e) {
    console.error(e)
  }
})

async function saveSubmission(status) {
  const itemsArray = []
  for (const [criteriaId, item] of Object.entries(items)) {
    if (item.evidence) {
      let filePath = ''
      if (item.file) {
        try {
          const uploadRes = await api.uploadFile(item.file)
          filePath = uploadRes.data.path
        } catch (e) { console.error('Upload failed:', e) }
      }
      itemsArray.push({
        criteria_id: parseInt(criteriaId),
        evidence: item.evidence,
        evidence_file: filePath,
      })
    }
  }

  const payload = {
    ...formData,
    items: itemsArray,
  }

  if (submissionId.value && isEdit.value) {
    await api.updateSubmission(submissionId.value, payload)
    if (status === 'submitted') await api.submitSubmission(submissionId.value)
    return submissionId.value
  } else {
    const res = await api.createSubmission(payload)
    const id = res.data.id
    if (status === 'submitted') await api.submitSubmission(id)
    return id
  }
}

async function saveDraft() {
  saving.value = true
  try {
    const id = await saveSubmission('draft')
    router.push(`/submissions/${id}`)
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function handleSubmit() {
  const { valid } = await form.value.validate()
  if (!valid) return
  submitting.value = true
  try {
    const id = await saveSubmission('submitted')
    router.push(`/submissions/${id}`)
  } catch (e) {
    console.error(e)
  } finally {
    submitting.value = false
  }
}
</script>
