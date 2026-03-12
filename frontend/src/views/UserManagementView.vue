<template>
  <div>
    <div class="d-flex align-center mb-6">
      <h1 class="text-h4 font-weight-bold text-primary">
        <v-icon icon="mdi-account-group" class="ml-2" />
        إدارة المستخدمين
      </h1>
      <v-spacer />
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openDialog()">إضافة مستخدم</v-btn>
    </div>

    <v-card rounded="xl">
      <v-data-table :headers="headers" :items="users" :loading="loading" hover density="comfortable">
        <template v-slot:item.role="{ item }">
          <v-chip :color="roleColor(item.role)" size="small" variant="tonal">{{ roleLabel(item.role) }}</v-chip>
        </template>
        <template v-slot:item.assigned_categories="{ item }">
          <span v-if="item.role === 'admin'">{{ item.assigned_categories?.length || 0 }} فئة</span>
          <span v-else class="text-grey">-</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon variant="text" size="small" @click="openDialog(item)">
            <v-icon>mdi-pencil</v-icon>
          </v-btn>
          <v-btn icon variant="text" size="small" color="error" @click="confirmDelete(item)">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </template>
      </v-data-table>
    </v-card>

    <!-- User Dialog -->
    <v-dialog v-model="dialog" max-width="600" persistent>
      <v-card rounded="xl">
        <v-card-title class="bg-primary text-white pa-4">
          {{ editMode ? 'تعديل مستخدم' : 'إضافة مستخدم جديد' }}
        </v-card-title>
        <v-card-text class="pa-4">
          <v-form ref="userForm">
            <v-text-field v-model="formData.username" label="اسم المستخدم" :rules="[v => !!v || 'مطلوب']" class="mb-2" />
            <v-text-field v-if="!editMode" v-model="formData.password" label="كلمة المرور" type="password" :rules="[v => !!v || 'مطلوب']" class="mb-2" />
            <v-text-field v-model="formData.full_name" label="الاسم الكامل" :rules="[v => !!v || 'مطلوب']" class="mb-2" />
            <v-text-field v-model="formData.email" label="البريد الالكتروني" type="email" class="mb-2" />
            <v-text-field v-model="formData.phone" label="رقم الهاتف" class="mb-2" />
            <v-select v-model="formData.role" :items="roleOptions" label="الدور" :rules="[v => !!v || 'مطلوب']" class="mb-2" />

            <v-autocomplete v-if="formData.role === 'university'" v-model="formData.university_id" :items="universities" item-title="name" item-value="id" label="الجامعة" class="mb-2" />

            <v-select v-if="formData.role === 'admin'" v-model="formData.assigned_categories" :items="categories" item-title="name_ar" item-value="id" label="الفئات المخصصة للمراجعة" multiple chips class="mb-2" />
          </v-form>
        </v-card-text>
        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn @click="dialog = false">إلغاء</v-btn>
          <v-btn color="primary" @click="saveUser" :loading="saving">حفظ</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card rounded="xl">
        <v-card-title class="text-error">تأكيد الحذف</v-card-title>
        <v-card-text>هل أنت متأكد من حذف المستخدم "{{ deleteTarget?.full_name }}"؟</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">إلغاء</v-btn>
          <v-btn color="error" @click="deleteUser" :loading="deleting">حذف</v-btn>
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
const deleting = ref(false)
const dialog = ref(false)
const deleteDialog = ref(false)
const editMode = ref(false)
const editId = ref(null)
const deleteTarget = ref(null)
const users = ref([])
const universities = ref([])
const categories = ref([])
const userForm = ref(null)

const formData = reactive({
  username: '', password: '', full_name: '', email: '', phone: '',
  role: '', university_id: null, assigned_categories: [],
})

const roleOptions = [
  { title: 'مدير عام', value: 'super_admin' },
  { title: 'مراجع', value: 'admin' },
  { title: 'جامعة', value: 'university' },
]

const headers = [
  { title: 'اسم المستخدم', key: 'username', sortable: true },
  { title: 'الاسم الكامل', key: 'full_name', sortable: true },
  { title: 'الدور', key: 'role', sortable: true },
  { title: 'البريد', key: 'email' },
  { title: 'الفئات', key: 'assigned_categories' },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

function roleColor(role) {
  return { super_admin: 'error', admin: 'warning', university: 'info' }[role] || 'grey'
}

function roleLabel(role) {
  return { super_admin: 'مدير عام', admin: 'مراجع', university: 'جامعة' }[role] || role
}

function openDialog(user) {
  if (user) {
    editMode.value = true
    editId.value = user.id
    Object.assign(formData, {
      username: user.username, full_name: user.full_name, email: user.email,
      phone: user.phone, role: user.role, university_id: user.university_id,
      assigned_categories: user.assigned_categories || [], password: '',
    })
  } else {
    editMode.value = false
    editId.value = null
    Object.assign(formData, {
      username: '', password: '', full_name: '', email: '', phone: '',
      role: '', university_id: null, assigned_categories: [],
    })
  }
  dialog.value = true
}

function confirmDelete(user) {
  deleteTarget.value = user
  deleteDialog.value = true
}

async function saveUser() {
  const { valid } = await userForm.value.validate()
  if (!valid) return
  saving.value = true
  try {
    if (editMode.value) {
      await api.updateUser(editId.value, formData)
      if (formData.role === 'admin') {
        await api.assignCategories(editId.value, { category_ids: formData.assigned_categories })
      }
    } else {
      await api.createUser(formData)
    }
    dialog.value = false
    loadData()
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function deleteUser() {
  deleting.value = true
  try {
    await api.deleteUser(deleteTarget.value.id)
    deleteDialog.value = false
    loadData()
  } catch (e) {
    console.error(e)
  } finally {
    deleting.value = false
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await api.getUsers()
    users.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const [uniRes, catRes] = await Promise.all([api.getUniversities(), api.getCategories()])
  universities.value = uniRes.data || []
  categories.value = catRes.data || []
  loadData()
})
</script>
