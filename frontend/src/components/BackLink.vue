<template>
  <router-link :to="to" class="back-link font-mono">
    ← {{ text }}
  </router-link>
</template>

<script setup>
defineProps({
  to: { type: String, required: true },
  text: { type: String, required: true }
})
</script>
