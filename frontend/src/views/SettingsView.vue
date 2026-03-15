<template>
  <div>
    <h2 class="text-h5 font-weight-bold mb-6">الإعدادات</h2>

    <v-alert v-if="successMsg" type="success" closable class="mb-4" @click:close="successMsg = ''">{{ successMsg }}</v-alert>
    <v-alert v-if="errorMsg" type="error" closable class="mb-4" @click:close="errorMsg = ''">{{ errorMsg }}</v-alert>

    <v-skeleton-loader v-if="loading" type="card, card, card" />

    <template v-else>
      <!-- Site Settings -->
      <v-card class="mb-6" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="ml-2" color="primary">mdi-web</v-icon>
          إعدادات الموقع و SEO
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="settings.site_title"
            label="عنوان النظام"
            variant="outlined"
            density="comfortable"
            class="mb-4"
            hint="يظهر في شريط العنوان ونتائج البحث"
            persistent-hint
          />
          <v-textarea
            v-model="settings.site_description"
            label="وصف النظام (SEO)"
            variant="outlined"
            density="comfortable"
            rows="3"
            hint="يظهر في نتائج محركات البحث"
            persistent-hint
          />
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-btn color="primary" variant="flat" :loading="saving" @click="saveSettings('site')">
            <v-icon start>mdi-content-save</v-icon>
            حفظ إعدادات الموقع
          </v-btn>
        </v-card-actions>
      </v-card>

      <!-- Submission Toggle -->
      <v-card class="mb-6" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="ml-2" color="warning">mdi-file-document-edit</v-icon>
          إدارة التقديمات
        </v-card-title>
        <v-card-text>
          <v-switch
            v-model="settings.submissions_open"
            :label="settings.submissions_open ? 'التقديمات مفتوحة - يمكن للجامعات التقديم' : 'التقديمات مغلقة - لا يمكن للجامعات التقديم'"
            color="success"
            inset
            hide-details
          />
          <v-alert v-if="!settings.submissions_open" type="warning" variant="tonal" class="mt-4" density="compact">
            عند إغلاق التقديمات، لن تتمكن الجامعات من إنشاء تقديمات جديدة أو إرسال المسودات الحالية.
          </v-alert>
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-btn color="warning" variant="flat" :loading="saving" @click="saveSettings('submissions')">
            <v-icon start>mdi-content-save</v-icon>
            حفظ حالة التقديمات
          </v-btn>
        </v-card-actions>
      </v-card>

      <!-- AI Settings -->
      <v-card class="mb-6" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="ml-2" color="deep-purple">mdi-robot</v-icon>
          إعدادات الذكاء الاصطناعي
        </v-card-title>
        <v-card-text>
          <!-- DeepSeek -->
          <div class="text-subtitle-1 font-weight-bold mb-2">
            DeepSeek
            <v-chip v-if="settings.has_deepseek_key" size="x-small" color="success" class="mr-2">مُفعّل</v-chip>
            <v-chip v-else size="x-small" color="grey" class="mr-2">غير مُعدّ</v-chip>
          </div>
          <v-row>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="newDeepSeekKey"
                label="مفتاح API جديد"
                variant="outlined"
                density="comfortable"
                type="password"
                :placeholder="settings.has_deepseek_key ? 'المفتاح محفوظ — اتركه فارغاً للإبقاء عليه' : 'أدخل مفتاح API'"
                persistent-placeholder
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="settings.deepseek_url"
                label="رابط API"
                variant="outlined"
                density="comfortable"
                dir="ltr"
              />
            </v-col>
          </v-row>
          <v-btn
            color="deep-purple"
            variant="tonal"
            size="small"
            class="mb-6"
            :loading="testingDeepSeek"
            :disabled="!newDeepSeekKey && !settings.has_deepseek_key"
            @click="testProvider('deepseek')"
          >
            <v-icon start>mdi-connection</v-icon>
            اختبار اتصال DeepSeek
          </v-btn>
          <v-alert v-if="deepseekResult !== null" :type="deepseekResult.success ? 'success' : 'error'" variant="tonal" density="compact" class="mb-4" closable @click:close="deepseekResult = null">
            {{ deepseekResult.success ? 'الاتصال ناجح: ' + deepseekResult.response : 'فشل الاتصال: ' + deepseekResult.error }}
          </v-alert>

          <v-divider class="mb-4" />

          <!-- Gemini -->
          <div class="text-subtitle-1 font-weight-bold mb-2">
            Gemini
            <v-chip v-if="settings.has_gemini_key" size="x-small" color="success" class="mr-2">مُفعّل</v-chip>
            <v-chip v-else size="x-small" color="grey" class="mr-2">غير مُعدّ</v-chip>
          </div>
          <v-row>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="newGeminiKey"
                label="مفتاح API جديد"
                variant="outlined"
                density="comfortable"
                type="password"
                :placeholder="settings.has_gemini_key ? 'المفتاح محفوظ — اتركه فارغاً للإبقاء عليه' : 'أدخل مفتاح API'"
                persistent-placeholder
              />
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="settings.gemini_url"
                label="رابط API"
                variant="outlined"
                density="comfortable"
                dir="ltr"
              />
            </v-col>
          </v-row>
          <v-btn
            color="deep-purple"
            variant="tonal"
            size="small"
            :loading="testingGemini"
            :disabled="!newGeminiKey && !settings.has_gemini_key"
            @click="testProvider('gemini')"
          >
            <v-icon start>mdi-connection</v-icon>
            اختبار اتصال Gemini
          </v-btn>
          <v-alert v-if="geminiResult !== null" :type="geminiResult.success ? 'success' : 'error'" variant="tonal" density="compact" class="mt-4" closable @click:close="geminiResult = null">
            {{ geminiResult.success ? 'الاتصال ناجح: ' + geminiResult.response : 'فشل الاتصال: ' + geminiResult.error }}
          </v-alert>
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-btn color="deep-purple" variant="flat" :loading="saving" @click="saveSettings('ai')">
            <v-icon start>mdi-content-save</v-icon>
            حفظ إعدادات الذكاء الاصطناعي
          </v-btn>
        </v-card-actions>
      </v-card>
      <!-- Security Settings -->
      <v-card class="mb-6" rounded="lg">
        <v-card-title class="d-flex align-center">
          <v-icon class="ml-2" color="error">mdi-shield-lock</v-icon>
          إعدادات الأمان
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12" md="4">
              <v-text-field
                v-model.number="settings.max_login_attempts"
                label="الحد الأقصى لمحاولات الدخول"
                variant="outlined"
                density="comfortable"
                type="number"
                min="1"
                hint="عدد المحاولات قبل حظر الحساب"
                persistent-hint
              />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field
                v-model.number="settings.block_duration_minutes"
                label="مدة الحظر (بالدقائق)"
                variant="outlined"
                density="comfortable"
                type="number"
                min="1"
                hint="مدة حظر الحساب بعد تجاوز المحاولات"
                persistent-hint
              />
            </v-col>
            <v-col cols="12" md="4">
              <v-text-field
                v-model.number="settings.max_file_size_mb"
                label="الحد الأقصى لحجم الملفات (ميغابايت)"
                variant="outlined"
                density="comfortable"
                type="number"
                min="1"
                hint="أقصى حجم للملفات المرفوعة"
                persistent-hint
              />
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="px-4 pb-4">
          <v-btn color="error" variant="flat" :loading="saving" @click="saveSettings('security')">
            <v-icon start>mdi-content-save</v-icon>
            حفظ إعدادات الأمان
          </v-btn>
        </v-card-actions>
      </v-card>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../services/api'

const loading = ref(true)
const saving = ref(false)
const successMsg = ref('')
const errorMsg = ref('')

const newDeepSeekKey = ref('')
const newGeminiKey = ref('')
const testingDeepSeek = ref(false)
const testingGemini = ref(false)
const deepseekResult = ref(null)
const geminiResult = ref(null)

const settings = ref({
  site_title: '',
  site_description: '',
  submissions_open: true,
  deepseek_url: '',
  gemini_url: '',
  has_deepseek_key: false,
  has_gemini_key: false,
  max_login_attempts: 5,
  block_duration_minutes: 30,
  max_file_size_mb: 10,
})

async function loadSettings() {
  try {
    const { data } = await api.getSettings()
    settings.value = data.settings
    newDeepSeekKey.value = ''
    newGeminiKey.value = ''
  } catch (e) {
    errorMsg.value = 'فشل في تحميل الإعدادات'
  } finally {
    loading.value = false
  }
}

async function saveSettings(section) {
  saving.value = true
  successMsg.value = ''
  errorMsg.value = ''

  try {
    let payload = {}
    if (section === 'site') {
      payload = {
        site_title: settings.value.site_title,
        site_description: settings.value.site_description,
      }
    } else if (section === 'submissions') {
      payload = {
        submissions_open: settings.value.submissions_open,
      }
    } else if (section === 'ai') {
      payload = {
        deepseek_url: settings.value.deepseek_url,
        gemini_url: settings.value.gemini_url,
      }
      // Only send key if user entered a new one
      if (newDeepSeekKey.value) {
        payload.deepseek_api_key = newDeepSeekKey.value
      }
      if (newGeminiKey.value) {
        payload.gemini_api_key = newGeminiKey.value
      }
    } else if (section === 'security') {
      payload = {
        max_login_attempts: String(settings.value.max_login_attempts),
        block_duration_minutes: String(settings.value.block_duration_minutes),
        max_file_size_mb: String(settings.value.max_file_size_mb),
      }
    }

    await api.updateSettings(payload)
    successMsg.value = 'تم حفظ الإعدادات بنجاح'
    // Reload to get updated state
    if (section === 'ai') {
      await loadSettings()
    }
  } catch (e) {
    errorMsg.value = 'فشل في حفظ الإعدادات: ' + (e.response?.data?.error || e.message)
  } finally {
    saving.value = false
  }
}

async function testProvider(provider) {
  const isDeepSeek = provider === 'deepseek'
  if (isDeepSeek) {
    testingDeepSeek.value = true
    deepseekResult.value = null
  } else {
    testingGemini.value = true
    geminiResult.value = null
  }

  try {
    // Use new key if entered, otherwise test with saved key (send empty to use server-side)
    const key = isDeepSeek ? newDeepSeekKey.value : newGeminiKey.value
    const { data } = await api.testAI({
      provider,
      api_key: key || '__use_saved__',
      base_url: isDeepSeek ? settings.value.deepseek_url : settings.value.gemini_url,
    })

    if (isDeepSeek) {
      deepseekResult.value = data
    } else {
      geminiResult.value = data
    }
  } catch (e) {
    const result = { success: false, error: e.response?.data?.error || e.message }
    if (isDeepSeek) {
      deepseekResult.value = result
    } else {
      geminiResult.value = result
    }
  } finally {
    if (isDeepSeek) testingDeepSeek.value = false
    else testingGemini.value = false
  }
}

onMounted(loadSettings)
</script>
