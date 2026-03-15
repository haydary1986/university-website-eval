<template>
  <div>
    <div class="d-flex align-center mb-6">
      <h2 class="text-h5 font-weight-bold">إدارة معايير التقييم</h2>
      <v-spacer />
      <v-chip color="primary" variant="tonal" class="ml-4">
        الوزن الكلي: {{ totalWeight }} نقطة
      </v-chip>
      <v-btn color="primary" variant="flat" class="mr-4" @click="openCategoryDialog()">
        <v-icon start>mdi-plus</v-icon>
        إضافة فئة
      </v-btn>
    </div>

    <v-alert v-if="successMsg" type="success" closable class="mb-4" @click:close="successMsg = ''">{{ successMsg }}</v-alert>
    <v-alert v-if="errorMsg" type="error" closable class="mb-4" @click:close="errorMsg = ''">{{ errorMsg }}</v-alert>

    <v-skeleton-loader v-if="loading" type="card, card, card" />

    <template v-else>
      <v-expansion-panels v-model="openPanels" multiple>
        <v-expansion-panel v-for="cat in categories" :key="cat.id" rounded="lg" class="mb-3">
          <v-expansion-panel-title>
            <div class="d-flex align-center flex-grow-1">
              <v-avatar color="primary" size="36" class="ml-3">
                <span class="text-white font-weight-bold">{{ cat.number }}</span>
              </v-avatar>
              <div>
                <div class="font-weight-bold">{{ cat.name_ar }}</div>
                <div class="text-caption text-grey">
                  {{ cat.criteria?.length || 0 }} فقرة
                  <v-chip v-if="cat.is_bonus" size="x-small" color="amber" class="mr-1">إضافي</v-chip>
                </div>
              </div>
              <v-spacer />
              <v-chip color="primary" variant="tonal" size="small" class="ml-2">
                {{ cat.weight }} نقطة
              </v-chip>
              <v-btn icon size="small" variant="text" color="primary" @click.stop="openCategoryDialog(cat)" class="ml-1">
                <v-icon size="18">mdi-pencil</v-icon>
                <v-tooltip activator="parent">تعديل الفئة</v-tooltip>
              </v-btn>
              <v-btn icon size="small" variant="text" color="error" @click.stop="confirmDeleteCategory(cat)" class="ml-1">
                <v-icon size="18">mdi-delete</v-icon>
                <v-tooltip activator="parent">حذف الفئة</v-tooltip>
              </v-btn>
            </div>
          </v-expansion-panel-title>

          <v-expansion-panel-text>
            <v-table density="comfortable">
              <thead>
                <tr>
                  <th>#</th>
                  <th>اسم الفقرة</th>
                  <th>الوصف</th>
                  <th>الدرجة القصوى</th>
                  <th style="width: 100px">إجراءات</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(cr, idx) in cat.criteria" :key="cr.id">
                  <td>{{ idx + 1 }}</td>
                  <td>{{ cr.name_ar }}</td>
                  <td class="text-caption text-grey">{{ cr.description || '—' }}</td>
                  <td>
                    <v-chip size="small" color="success" variant="tonal">{{ cr.max_score }}</v-chip>
                  </td>
                  <td>
                    <v-btn icon size="x-small" variant="text" color="primary" @click="openCriteriaDialog(cat, cr)">
                      <v-icon size="16">mdi-pencil</v-icon>
                    </v-btn>
                    <v-btn icon size="x-small" variant="text" color="error" @click="confirmDeleteCriteria(cat, cr)">
                      <v-icon size="16">mdi-delete</v-icon>
                    </v-btn>
                  </td>
                </tr>
                <tr v-if="!cat.criteria?.length">
                  <td colspan="5" class="text-center text-grey py-4">لا توجد فقرات</td>
                </tr>
              </tbody>
            </v-table>
            <v-btn color="success" variant="tonal" size="small" class="mt-3" @click="openCriteriaDialog(cat)">
              <v-icon start>mdi-plus</v-icon>
              إضافة فقرة
            </v-btn>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>

      <v-alert v-if="!categories.length" type="info" variant="tonal" class="mt-4">
        لا توجد فئات. أضف فئة جديدة للبدء.
      </v-alert>
    </template>

    <!-- Category Dialog -->
    <v-dialog v-model="catDialog" max-width="500" persistent>
      <v-card>
        <v-card-title>{{ editingCategory ? 'تعديل الفئة' : 'إضافة فئة جديدة' }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="catForm.name_ar"
            label="اسم الفئة"
            variant="outlined"
            density="comfortable"
            class="mb-3"
            :rules="[v => !!v || 'مطلوب']"
          />
          <v-text-field
            v-model.number="catForm.number"
            label="رقم الفئة"
            variant="outlined"
            density="comfortable"
            type="number"
            class="mb-3"
            hint="اتركه فارغاً للترقيم التلقائي"
            persistent-hint
          />
          <v-switch
            v-model="catForm.is_bonus"
            label="فئة إضافية (Bonus)"
            color="amber"
            inset
            hide-details
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="catDialog = false">إلغاء</v-btn>
          <v-btn color="primary" variant="flat" :loading="saving" @click="saveCategory">حفظ</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Criteria Dialog -->
    <v-dialog v-model="crDialog" max-width="500" persistent>
      <v-card>
        <v-card-title>{{ editingCriteria ? 'تعديل الفقرة' : 'إضافة فقرة جديدة' }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="crForm.name_ar"
            label="اسم الفقرة"
            variant="outlined"
            density="comfortable"
            class="mb-3"
            :rules="[v => !!v || 'مطلوب']"
          />
          <v-textarea
            v-model="crForm.description"
            label="الوصف (اختياري)"
            variant="outlined"
            density="comfortable"
            rows="2"
            class="mb-3"
          />
          <v-text-field
            v-model.number="crForm.max_score"
            label="الدرجة القصوى"
            variant="outlined"
            density="comfortable"
            type="number"
            min="0"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="crDialog = false">إلغاء</v-btn>
          <v-btn color="primary" variant="flat" :loading="saving" @click="saveCriteria">حفظ</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="text-error">تأكيد الحذف</v-card-title>
        <v-card-text>{{ deleteMsg }}</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">إلغاء</v-btn>
          <v-btn color="error" variant="flat" :loading="saving" @click="executeDelete">حذف</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../services/api'

const loading = ref(true)
const saving = ref(false)
const successMsg = ref('')
const errorMsg = ref('')
const categories = ref([])
const openPanels = ref([])

// Category dialog
const catDialog = ref(false)
const editingCategory = ref(null)
const catForm = ref({ name_ar: '', number: null, is_bonus: false })

// Criteria dialog
const crDialog = ref(false)
const editingCriteria = ref(null)
const currentCategoryId = ref(null)
const crForm = ref({ name_ar: '', description: '', max_score: 0 })

// Delete dialog
const deleteDialog = ref(false)
const deleteMsg = ref('')
const deleteAction = ref(null)

const totalWeight = computed(() => {
  return categories.value.filter(c => !c.is_bonus).reduce((sum, c) => sum + (c.weight || 0), 0)
})

async function loadCategories() {
  try {
    const { data } = await api.getCategories()
    categories.value = data
  } catch (e) {
    errorMsg.value = 'فشل في تحميل الفئات'
  } finally {
    loading.value = false
  }
}

// Category CRUD
function openCategoryDialog(cat = null) {
  editingCategory.value = cat
  if (cat) {
    catForm.value = { name_ar: cat.name_ar, number: cat.number, is_bonus: cat.is_bonus }
  } else {
    catForm.value = { name_ar: '', number: null, is_bonus: false }
  }
  catDialog.value = true
}

async function saveCategory() {
  if (!catForm.value.name_ar) return
  saving.value = true
  errorMsg.value = ''
  try {
    if (editingCategory.value) {
      await api.updateCategory(editingCategory.value.id, catForm.value)
      successMsg.value = 'تم تحديث الفئة بنجاح'
    } else {
      await api.createCategory(catForm.value)
      successMsg.value = 'تم إنشاء الفئة بنجاح'
    }
    catDialog.value = false
    await loadCategories()
  } catch (e) {
    errorMsg.value = e.response?.data?.error || 'فشل في حفظ الفئة'
  } finally {
    saving.value = false
  }
}

function confirmDeleteCategory(cat) {
  deleteMsg.value = `هل أنت متأكد من حذف الفئة "${cat.name_ar}" وجميع فقراتها؟`
  deleteAction.value = async () => {
    try {
      await api.deleteCategory(cat.id)
      successMsg.value = 'تم حذف الفئة بنجاح'
      await loadCategories()
    } catch (e) {
      errorMsg.value = e.response?.data?.error || 'فشل في حذف الفئة'
    }
  }
  deleteDialog.value = true
}

// Criteria CRUD
function openCriteriaDialog(cat, cr = null) {
  currentCategoryId.value = cat.id
  editingCriteria.value = cr
  if (cr) {
    crForm.value = { name_ar: cr.name_ar, description: cr.description || '', max_score: cr.max_score }
  } else {
    crForm.value = { name_ar: '', description: '', max_score: 0 }
  }
  crDialog.value = true
}

async function saveCriteria() {
  if (!crForm.value.name_ar) return
  saving.value = true
  errorMsg.value = ''
  try {
    if (editingCriteria.value) {
      await api.updateCriteria(editingCriteria.value.id, crForm.value)
      successMsg.value = 'تم تحديث الفقرة بنجاح'
    } else {
      await api.createCriteria(currentCategoryId.value, crForm.value)
      successMsg.value = 'تم إنشاء الفقرة بنجاح'
    }
    crDialog.value = false
    await loadCategories()
  } catch (e) {
    errorMsg.value = e.response?.data?.error || 'فشل في حفظ الفقرة'
  } finally {
    saving.value = false
  }
}

function confirmDeleteCriteria(cat, cr) {
  deleteMsg.value = `هل أنت متأكد من حذف الفقرة "${cr.name_ar}" من فئة "${cat.name_ar}"؟`
  deleteAction.value = async () => {
    try {
      await api.deleteCriteria(cr.id)
      successMsg.value = 'تم حذف الفقرة بنجاح'
      await loadCategories()
    } catch (e) {
      errorMsg.value = e.response?.data?.error || 'فشل في حذف الفقرة'
    }
  }
  deleteDialog.value = true
}

async function executeDelete() {
  saving.value = true
  deleteDialog.value = false
  if (deleteAction.value) await deleteAction.value()
  saving.value = false
}

onMounted(loadCategories)
</script>
