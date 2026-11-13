<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { FeedingPublishInput } from '../api/generated/types.gen'
import { publish } from '../lib/api'
import { browserTimeZone, localDateKey, localMidnightUtc } from '../lib/dates'

const router = useRouter()
const today = localDateKey(new Date())
const startDate = ref(today)
const endDate = ref('')
const feedingCount = ref(3)
const firstMeal = ref('07:00')
const lastMeal = ref('19:00')
const automaticTimes = ref(true)
const feedingTimes = ref<string[]>([])
const mode = ref<'standard' | 'weaning'>('standard')
const amountMode = ref<'daily' | 'perMeal'>('daily')
const standardQuantity = ref<number | null>(null)
const standardUnit = ref('')
const oldFood = ref('')
const newFood = ref('')
const oldQuantity = ref<number | null>(null)
const oldUnit = ref('')
const newQuantity = ref<number | null>(null)
const newUnit = ref('')
const saving = ref(false)
const errorMessage = ref('')

const startLabel = computed(() => oldFood.value.trim() || 'Old food')
const newLabel = computed(() => newFood.value.trim() || 'New food')

function calculatedTimes(): string[] {
  const count = Math.min(12, Math.max(1, Number(feedingCount.value) || 1))
  const first = toMinutes(firstMeal.value)
  const last = toMinutes(lastMeal.value)
  if (count === 1) return [firstMeal.value]
  return Array.from({ length: count }, (_, index) => {
    const minute = Math.round(first + ((last - first) * index) / (count - 1))
    return `${String(Math.floor(minute / 60)).padStart(2, '0')}:${String(minute % 60).padStart(2, '0')}`
  })
}

function toMinutes(clock: string): number {
  const [hour, minute] = clock.split(':').map(Number)
  return hour * 60 + minute
}

watch([feedingCount, firstMeal, lastMeal, automaticTimes], () => {
  const defaults = calculatedTimes()
  if (automaticTimes.value) {
    feedingTimes.value = defaults
    return
  }
  const next = feedingTimes.value.slice(0, feedingCount.value)
  while (next.length < feedingCount.value) next.push(defaults[next.length] ?? '12:00')
  feedingTimes.value = next
}, { immediate: true })

function setMode(next: 'standard' | 'weaning') {
  mode.value = next
  errorMessage.value = ''
}

function validate(): string {
  if (!startDate.value || startDate.value < today) return 'Choose today or a later start date.'
  if (endDate.value && endDate.value < startDate.value) return 'End date must be on or after the start date.'
  if (mode.value === 'weaning' && !endDate.value) return 'Choose an end date for weaning.'
  if (feedingTimes.value.length !== feedingCount.value || feedingTimes.value.some((time) => !time)) {
    return 'Set a time for every feeding.'
  }
  const minutes = feedingTimes.value.map(toMinutes)
  if (minutes.some((minute, index) => index > 0 && minute <= minutes[index - 1])) {
    return 'Feeding times must be in increasing order.'
  }
  if (mode.value === 'standard') {
    if (!standardQuantity.value || standardQuantity.value <= 0 || !standardUnit.value.trim()) {
      return 'Enter a quantity and unit.'
    }
  } else if (
    !oldFood.value.trim() || !newFood.value.trim() ||
    !oldQuantity.value || oldQuantity.value <= 0 || !oldUnit.value.trim() ||
    !newQuantity.value || newQuantity.value <= 0 || !newUnit.value.trim()
  ) {
    return 'Enter both food names, daily amounts, and units.'
  }
  return ''
}

async function save() {
  errorMessage.value = validate()
  if (errorMessage.value) return

  const body: FeedingPublishInput = {
    startAt: localMidnightUtc(startDate.value),
    endAt: endDate.value ? localMidnightUtc(endDate.value) : undefined,
    timeZone: browserTimeZone(),
    times: [...feedingTimes.value],
    mode: mode.value,
  }
  if (mode.value === 'standard') {
    body.standard = {
      amountMode: amountMode.value,
      quantity: standardQuantity.value ?? 0,
      unit: standardUnit.value.trim(),
    }
  } else {
    body.weaning = {
      oldFood: oldFood.value.trim(),
      newFood: newFood.value.trim(),
      oldQuantity: oldQuantity.value ?? 0,
      oldUnit: oldUnit.value.trim(),
      newQuantity: newQuantity.value ?? 0,
      newUnit: newUnit.value.trim(),
    }
  }

  saving.value = true
  try {
    await publish(body)
    await router.push('/')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Could not publish schedule.'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <form class="space-y-7" @submit.prevent="save">
    <h1 class="text-xl font-semibold">Configure schedule</h1>

    <section class="space-y-4">
      <div class="grid grid-cols-2 gap-3">
        <label class="space-y-1.5 text-sm font-medium">
          <span>Start date</span>
          <input v-model="startDate" type="date" :min="today" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
        </label>
        <label class="space-y-1.5 text-sm font-medium">
          <span>End date</span>
          <input v-model="endDate" type="date" :min="startDate || today" :required="mode === 'weaning'" class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
        </label>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <label class="space-y-1.5 text-sm font-medium">
          <span>Feedings per day</span>
          <input v-model.number="feedingCount" type="number" min="1" max="12" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
        </label>
        <div class="flex items-end">
          <label class="flex h-[46px] w-full items-center justify-between rounded-xl border border-slate-300 bg-white px-3 dark:border-slate-700 dark:bg-slate-900">
            <span class="text-sm font-medium">Automatic times</span>
            <input v-model="automaticTimes" type="checkbox" class="size-5 accent-emerald-600" />
          </label>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <label class="space-y-1.5 text-sm font-medium">
          <span>First meal</span>
          <input v-model="firstMeal" type="time" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
        </label>
        <label class="space-y-1.5 text-sm font-medium">
          <span>Last meal</span>
          <input v-model="lastMeal" type="time" :disabled="feedingCount === 1" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base disabled:opacity-50 dark:border-slate-700 dark:bg-slate-900" />
        </label>
      </div>

      <div class="space-y-2">
        <span class="text-sm font-medium">Feeding times</span>
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
          <label v-for="(_, index) in feedingTimes" :key="index" class="space-y-1 text-xs text-slate-500 dark:text-slate-400">
            <span>Meal {{ index + 1 }}</span>
            <input v-model="feedingTimes[index]" type="time" :readonly="automaticTimes" required class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-base text-slate-900 read-only:bg-slate-100 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:read-only:bg-slate-800" />
          </label>
        </div>
      </div>
    </section>

    <section class="space-y-4">
      <div class="grid grid-cols-2 rounded-xl bg-slate-200 p-1 dark:bg-slate-800">
        <button type="button" class="rounded-lg px-3 py-2 text-sm font-semibold" :class="mode === 'standard' ? 'bg-white shadow-sm dark:bg-slate-700' : 'text-slate-600 dark:text-slate-300'" :aria-pressed="mode === 'standard'" @click="setMode('standard')">Standard</button>
        <button type="button" class="rounded-lg px-3 py-2 text-sm font-semibold" :class="mode === 'weaning' ? 'bg-white shadow-sm dark:bg-slate-700' : 'text-slate-600 dark:text-slate-300'" :aria-pressed="mode === 'weaning'" @click="setMode('weaning')">Weaning</button>
      </div>

      <div v-if="mode === 'standard'" class="space-y-4 rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
        <div class="grid grid-cols-2 rounded-xl bg-slate-100 p-1 dark:bg-slate-900">
          <button type="button" class="rounded-lg px-2 py-2 text-sm font-medium" :class="amountMode === 'daily' ? 'bg-white shadow-sm dark:bg-slate-700' : 'text-slate-600 dark:text-slate-300'" :aria-pressed="amountMode === 'daily'" @click="amountMode = 'daily'">Per day</button>
          <button type="button" class="rounded-lg px-2 py-2 text-sm font-medium" :class="amountMode === 'perMeal' ? 'bg-white shadow-sm dark:bg-slate-700' : 'text-slate-600 dark:text-slate-300'" :aria-pressed="amountMode === 'perMeal'" @click="amountMode = 'perMeal'">Per meal</button>
        </div>
        <div class="grid grid-cols-[1fr_6rem] gap-3">
          <label class="space-y-1.5 text-sm font-medium">
            <span>{{ amountMode === 'daily' ? 'Daily amount' : 'Amount per meal' }}</span>
            <input v-model.number="standardQuantity" type="number" min="0.001" step="any" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
          </label>
          <label class="space-y-1.5 text-sm font-medium">
            <span>Unit</span>
            <input v-model="standardUnit" type="text" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
          </label>
        </div>
      </div>

      <div v-else class="space-y-4 rounded-2xl border border-slate-200 p-4 dark:border-slate-800">
        <label class="block space-y-1.5 text-sm font-medium">
          <span>Old food</span>
          <input v-model="oldFood" type="text" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
        </label>
        <label class="block space-y-1.5 text-sm font-medium">
          <span>New food</span>
          <input v-model="newFood" type="text" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
        </label>
        <div class="grid grid-cols-[1fr_6rem] gap-3">
          <label class="space-y-1.5 text-sm font-medium">
            <span>{{ startLabel }} daily recommended amount</span>
            <input v-model.number="oldQuantity" type="number" min="0.001" step="any" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
          </label>
          <label class="space-y-1.5 text-sm font-medium">
            <span>Unit</span>
            <input v-model="oldUnit" type="text" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
          </label>
        </div>
        <div class="grid grid-cols-[1fr_6rem] gap-3">
          <label class="space-y-1.5 text-sm font-medium">
            <span>{{ newLabel }} daily recommended amount</span>
            <input v-model.number="newQuantity" type="number" min="0.001" step="any" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
          </label>
          <label class="space-y-1.5 text-sm font-medium">
            <span>Unit</span>
            <input v-model="newUnit" type="text" required class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-base dark:border-slate-700 dark:bg-slate-900" />
          </label>
        </div>
      </div>
    </section>

    <p v-if="errorMessage" role="alert" class="text-sm font-medium text-red-600 dark:text-red-400">{{ errorMessage }}</p>
    <button type="submit" :disabled="saving" class="w-full rounded-xl bg-emerald-600 px-4 py-3 font-semibold text-white hover:bg-emerald-700 disabled:opacity-60">
      {{ saving ? 'Publishing…' : 'Publish schedule' }}
    </button>
  </form>
</template>
