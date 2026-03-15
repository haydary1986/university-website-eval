<template>
  <div>
    <div class="d-flex align-center mb-6">
      <h1 class="text-h5 text-md-h4 font-weight-bold text-primary">
        <v-icon icon="mdi-calendar-range" class="ml-2" />
        السنوات الدراسية
      </h1>
      <v-spacer />
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openDialog()">إضافة سنة</v-btn>
    </div>

    <v-card rounded="xl">
      <v-data-table :headers="headers" :items="years" :loading="loading" hover density="comfortable">
        <template v-slot:item.is_active="{ item }">
          <v-chip :color="item.is_active ? 'success' : 'grey'" size="small">
            {{ item.is_active ? 'فعالة' : 'غير فعالة' }}
          </v-chip>
        </template>
        <template v-slot:item.submission_deadline="{ item }">
          <template v-if="item.submission_deadline">
            <v-chip :color="isDeadlinePassed(item.submission_deadline) ? 'error' : 'info'" size="small">
              {{ formatDate(item.submission_deadline) }}
            </v-chip>
          </template>
          <span v-else class="text-grey">—</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon variant="text" size="small" @click="openDialog(item)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="500" persistent>
      <v-card rounded="xl">
        <v-card-title class="bg-primary text-white pa-4">
          {{ editMode ? 'تعديل سنة دراسية' : 'إضافة سنة دراسية' }}
        </v-card-title>
        <v-card-text class="pa-4">
          <v-form ref="yearForm">
            <v-text-field v-model="formData.name" label="الاسم (مثال: 2025-2026)" :rules="[v => !!v || 'مطلوب']" class="mb-3" />
            <v-text-field v-model="formData.start_date" label="تاريخ البدء" type="date" class="mb-3" />
            <v-text-field v-model="formData.end_date" label="تاريخ الانتهاء" type="date" class="mb-3" />
            <v-text-field v-model="formData.submission_deadline" label="آخر موعد للتقديم" type="date" class="mb-3" hint="بعد هذا التاريخ لن تتمكن الجامعات من التقديم لهذه السنة" persistent-hint />
            <v-switch v-model="formData.is_active" label="فعالة" color="success" />
          </v-form>
        </v-card-text>
        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn @click="dialog = false">إلغاء</v-btn>
          <v-btn color="primary" @click="save" :loading="saving">حفظ</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import api from '../services/api'

const loading = ref(false)
const saving = ref(false)
const dialog = ref(false)
const editMode = ref(false)
const editId = ref(null)
const years = ref([])
const yearForm = ref(null)

const formData = reactive({ name: '', start_date: '', end_date: '', submission_deadline: '', is_active: false })

const headers = [
  { title: 'الاسم', key: 'name', sortable: true },
  { title: 'تاريخ البدء', key: 'start_date' },
  { title: 'تاريخ الانتهاء', key: 'end_date' },
  { title: 'آخر موعد للتقديم', key: 'submission_deadline' },
  { title: 'الحالة', key: 'is_active', sortable: true },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('ar-IQ')
}

function isDeadlinePassed(d) {
  return new Date(d) < new Date()
}

function openDialog(year) {
  if (year) {
    editMode.value = true
    editId.value = year.id
    Object.assign(formData, {
      name: year.name,
      start_date: year.start_date?.substring(0, 10),
      end_date: year.end_date?.substring(0, 10),
      submission_deadline: year.submission_deadline?.substring(0, 10) || '',
      is_active: year.is_active,
    })
  } else {
    editMode.value = false
    Object.assign(formData, { name: '', start_date: '', end_date: '', submission_deadline: '', is_active: false })
  }
  dialog.value = true
}

async function save() {
  const { valid } = await yearForm.value.validate()
  if (!valid) return
  saving.value = true
  try {
    if (editMode.value) {
      await api.updateAcademicYear(editId.value, formData)
    } else {
      await api.createAcademicYear(formData)
    }
    dialog.value = false
    loadData()
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await api.getAcademicYears()
    years.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
