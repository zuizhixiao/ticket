<script setup lang="ts">
import { toastState, dismiss } from '@/composables/toast'

const icons: Record<string, string> = {
  success: '✔',
  error: '✕',
  info: '✦'
}
</script>

<template>
  <Teleport to="body">
    <div class="toasts">
      <TransitionGroup name="toast">
        <div v-for="t in toastState.list" :key="t.id" class="toast" :class="t.type">
          <span class="toast-ico">{{ icons[t.type] || '✦' }}</span>
          <span>{{ t.message }}</span>
          <button
            class="dropdown-item"
            style="margin-left:auto;width:auto;padding:2px 6px;color:var(--text-muted)"
            @click="dismiss(t.id)"
          >
            ✕
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s cubic-bezier(0.22, 1, 0.36, 1);
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(24px);
}
</style>
