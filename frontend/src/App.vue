<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'

const darkMode = ref(false)

function applyTheme(dark: boolean) {
  darkMode.value = dark
  document.documentElement.classList.toggle('dark', dark)
  localStorage.setItem('puppy-food-dark', String(dark))
}

onMounted(() => {
  applyTheme(localStorage.getItem('puppy-food-dark') === 'true')
})
</script>

<template>
  <div class="min-h-dvh bg-slate-50 text-slate-900 transition-colors dark:bg-slate-950 dark:text-slate-100">
    <header class="sticky top-0 z-20 border-b border-slate-200 bg-slate-50/95 backdrop-blur dark:border-slate-800 dark:bg-slate-950/95">
      <div class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4">
        <nav class="flex items-center gap-5 text-sm font-medium">
          <RouterLink to="/" class="text-slate-900 dark:text-slate-100">Schedule</RouterLink>
          <RouterLink to="/configure" class="text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100">Configure</RouterLink>
        </nav>
        <button
          type="button"
          class="grid size-9 place-items-center rounded-full text-lg text-slate-600 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-800"
          :aria-label="darkMode ? 'Switch to light mode' : 'Switch to dark mode'"
          @click="applyTheme(!darkMode)"
        >
          <span aria-hidden="true">{{ darkMode ? '☼' : '☾' }}</span>
        </button>
      </div>
    </header>
    <main class="mx-auto max-w-3xl px-4 pb-12 pt-5">
      <RouterView />
    </main>
  </div>
</template>
