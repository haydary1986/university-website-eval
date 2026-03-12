<template>
  <div>
    <div class="d-flex align-center mb-6">
      <v-btn icon variant="text" @click="$router.back()" class="ml-2">
        <v-icon>mdi-arrow-right</v-icon>
      </v-btn>
      <h1 class="text-h4 font-weight-bold text-primary">
        <v-icon icon="mdi-compare" class="ml-2" />
        مقارنة النسخ
      </h1>
    </div>

    <v-card class="mb-4 pa-4" rounded="xl">
      <v-row>
        <v-col cols="12" md="4">
          <v-select v-model="selectedVersion" :items="versions" item-title="label" item-value="version" label="مقارنة مع النسخة" @update:model-value="loadDiff" />
        </v-col>
        <v-col cols="12" md="4">
          <div class="text-caption text-medium-emphasis">الجامعة</div>
          <div class="font-weight-bold">{{ submission?.university_name }}</div>
        </v-col>
      </v-row>
    </v-card>

    <v-skeleton-loader v-if="loading" type="card, card" />

    <template v-if="diffData && !loading">
      <v-card v-for="cat in diffData" :key="cat.category_id" class="mb-4" rounded="xl">
        <v-card-title class="bg-primary text-white pa-3">{{ cat.category_name }}</v-card-title>
        <v-card-text class="pa-0">
          <v-table density="comfortable">
            <thead>
              <tr>
                <th style="width:25%">المعيار</th>
                <th style="width:30%">النسخة السابقة</th>
                <th style="width:30%">النسخة الحالية</th>
                <th style="width:15%">التغيير</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in cat.items" :key="item.criteria_id" :class="diffRowClass(item)">
                <td class="font-weight-medium">{{ item.criteria_name }}</td>
                <td>
                  <div class="text-body-2">{{ item.old_evidence || '-' }}</div>
                  <v-chip v-if="item.old_score !== undefined" size="x-small" class="mt-1">{{ item.old_score }}</v-chip>
                </td>
                <td>
                  <div class="text-body-2">{{ item.new_evidence || '-' }}</div>
                  <v-chip v-if="item.new_score !== undefined" size="x-small" color="primary" class="mt-1">{{ item.new_score }}</v-chip>
                </td>
                <td>
                  <v-icon v-if="item.change === 'added'" color="success">mdi-plus-circle</v-icon>
                  <v-icon v-else-if="item.change === 'removed'" color="error">mdi-minus-circle</v-icon>
                  <v-icon v-else-if="item.change === 'modified'" color="warning">mdi-pencil-circle</v-icon>
                  <v-icon v-else color="grey">mdi-minus</v-icon>
                  <span class="text-caption mr-1">{{ changeLabel(item.change) }}</span>
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-card-text>
      </v-card>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../services/api'

const route = useRoute()
const loading = ref(false)
const submission = ref(null)
const diffData = ref(null)
const versions = ref([])
const selectedVersion = ref(null)

function diffRowClass(item) {
  const map = { added: 'bg-green-lighten-5', removed: 'bg-red-lighten-5', modified: 'bg-yellow-lighten-5' }
  return map[item.change] || ''
}

function changeLabel(change) {
  const map = { added: 'جديد', removed: 'محذوف', modified: 'معدّل', unchanged: 'بدون تغيير' }
  return map[change] || ''
}

async function loadDiff() {
  if (!selectedVersion.value) return
  loading.value = true
  try {
    const res = await api.getSubmissionDiff(route.params.id, selectedVersion.value)
    diffData.value = res.data.categories || res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const res = await api.getSubmission(route.params.id)
    submission.value = res.data
    // Build version list
    for (let i = 1; i < submission.value.version; i++) {
      versions.value.push({ label: `النسخة ${i}`, version: i })
    }
    if (versions.value.length > 0) {
      selectedVersion.value = versions.value[versions.value.length - 1].version
      loadDiff()
    }
  } catch (e) {
    console.error(e)
  }
})
</script>
