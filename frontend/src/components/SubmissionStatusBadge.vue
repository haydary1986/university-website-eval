<template>
  <v-chip :color="color" size="small" variant="tonal">{{ label }}</v-chip>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({ status: String })

const statusMap = {
  draft: { color: 'grey', label: 'مسودة' },
  submitted: { color: 'blue', label: 'مقدم' },
  under_review: { color: 'orange', label: 'قيد المراجعة' },
  approved: { color: 'green', label: 'معتمد' },
  rejected: { color: 'red', label: 'مرفوض' },
}

const color = computed(() => statusMap[props.status]?.color || 'grey')
const label = computed(() => statusMap[props.status]?.label || props.status)
</script>
