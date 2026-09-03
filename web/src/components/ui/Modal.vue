<script setup lang="ts">
defineProps<{ show: boolean; title?: string; width?: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

function onMask(e: MouseEvent) {
  if (e.target === e.currentTarget) emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-mask" @click.self="onMask">
      <div class="modal-card" :style="{ maxWidth: width || '640px' }">
        <div class="row-between mb-2">
          <h3 class="panel-title" style="margin:0">
            {{ title || '提示' }}
          </h3>
          <button class="btn btn-ghost btn-sm" @click="emit('close')">✕</button>
        </div>
        <slot />
      </div>
    </div>
  </Teleport>
</template>
