<template>
  <div>
    <h1 class="text-h4 font-weight-bold text-primary mb-6">
      <v-icon icon="mdi-school" class="ml-2" />
      الجامعات
    </h1>

    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row dense>
        <v-col cols="12" md="4">
          <v-text-field v-model="search" label="بحث بالاسم" prepend-inner-icon="mdi-magnify" clearable hide-details />
        </v-col>
        <v-col cols="12" md="3">
          <v-btn-toggle v-model="typeFilter" mandatory color="primary" variant="outlined">
            <v-btn value="all">الكل</v-btn>
            <v-btn value="government">حكومية</v-btn>
            <v-btn value="private">أهلية</v-btn>
          </v-btn-toggle>
        </v-col>
      </v-row>
    </v-card>

    <v-card rounded="xl">
      <v-data-table :headers="headers" :items="filteredUniversities" :search="search" :loading="loading" hover density="comfortable" :items-per-page="20">
        <template v-slot:item.type="{ item }">
          <v-chip :color="item.type === 'government' ? 'primary' : 'secondary'" size="small" variant="tonal">
            {{ item.type === 'government' ? 'حكومية' : 'أهلية' }}
          </v-chip>
        </template>
        <template v-slot:item.website="{ item }">
          <a v-if="item.website" :href="item.website.startsWith('http') ? item.website : 'https://' + item.website" target="_blank" class="text-primary">
            {{ item.website }}
          </a>
          <span v-else class="text-grey">-</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon variant="text" size="small" @click="viewUniversity(item)">
            <v-icon>mdi-eye</v-icon>
          </v-btn>
        </template>
      </v-data-table>
    </v-card>

    <!-- University Detail Dialog -->
    <v-dialog v-model="detailDialog" max-width="600">
      <v-card v-if="selectedUni" rounded="xl">
        <v-card-title class="bg-primary text-white pa-4">
          {{ selectedUni.name }}
        </v-card-title>
        <v-card-text class="pa-4">
          <v-list>
            <v-list-item>
              <template v-slot:prepend><v-icon>mdi-web</v-icon></template>
              <v-list-item-title>الموقع</v-list-item-title>
              <v-list-item-subtitle>{{ selectedUni.website || '-' }}</v-list-item-subtitle>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend><v-icon>mdi-tag</v-icon></template>
              <v-list-item-title>النوع</v-list-item-title>
              <v-list-item-subtitle>{{ selectedUni.type === 'government' ? 'حكومية' : 'أهلية' }}</v-list-item-subtitle>
            </v-list-item>
            <v-list-item>
              <template v-slot:prepend><v-icon>mdi-account</v-icon></template>
              <v-list-item-title>جهة الاتصال</v-list-item-title>
              <v-list-item-subtitle>{{ selectedUni.contact_person || '-' }}</v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="detailDialog = false">إغلاق</v-btn>
          <v-btn color="primary" :to="`/submissions?university_id=${selectedUni.id}`">عرض التقديمات</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../services/api'

const loading = ref(false)
const universities = ref([])
const search = ref('')
const typeFilter = ref('all')
const detailDialog = ref(false)
const selectedUni = ref(null)

const headers = [
  { title: 'الاسم', key: 'name', sortable: true },
  { title: 'النوع', key: 'type', sortable: true },
  { title: 'الموقع الالكتروني', key: 'website' },
  { title: 'جهة الاتصال', key: 'contact_person' },
  { title: 'الإجراءات', key: 'actions', sortable: false },
]

const filteredUniversities = computed(() => {
  if (typeFilter.value === 'all') return universities.value
  return universities.value.filter(u => u.type === typeFilter.value)
})

function viewUniversity(uni) {
  selectedUni.value = uni
  detailDialog.value = true
}

onMounted(async () => {
  loading.value = true
  try {
    const res = await api.getUniversities()
    universities.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
})
</script>
