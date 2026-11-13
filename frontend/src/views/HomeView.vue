<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import type { FeedingFeeding } from '../api/generated/types.gen'
import { loadFeedings, updateCompletion } from '../lib/api'
import { addDays, dateFromKey, formatClock, formatDay, formatQuantity, localDateKey, localMidnightUtc } from '../lib/dates'
import { chooseFeedingForDay, chooseInitialFeeding } from '../lib/feeding-selection'

const today = localDateKey(new Date())
const dates = ref<string[]>([])
const selectedDate = ref(today)
const feedings = ref<FeedingFeeding[]>([])
const openId = ref('')
const loading = ref(false)
const changingId = ref('')
const errorMessage = ref('')
const calendar = ref<HTMLElement>()
let extending = false

const selectedFeedings = computed(() => feedingsFor(selectedDate.value))

function feedingsFor(key: string): FeedingFeeding[] {
  return feedings.value
    .filter((feeding) => feeding.scheduledAt && localDateKey(new Date(feeding.scheduledAt)) === key)
    .sort((a, b) => new Date(a.scheduledAt ?? 0).getTime() - new Date(b.scheduledAt ?? 0).getTime())
}

function dotsFor(key: string): FeedingFeeding[] {
  return feedingsFor(key)
}

async function fetchRange(start: string, endExclusive: string) {
  loading.value = true
  errorMessage.value = ''
  try {
    const received = await loadFeedings(localMidnightUtc(start), localMidnightUtc(endExclusive))
    const byId = new Map(feedings.value.map((feeding) => [feeding.id, feeding]))
    for (const feeding of received) {
      if (feeding.id) byId.set(feeding.id, feeding)
    }
    feedings.value = [...byId.values()]
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Could not load the schedule.'
  } finally {
    loading.value = false
  }
}

function fillDates(start: string, count: number): string[] {
  return Array.from({ length: count }, (_, index) => addDays(start, index))
}

function selectFeedingDay(feeding: FeedingFeeding | undefined) {
  if (!feeding?.scheduledAt) return
  selectedDate.value = localDateKey(new Date(feeding.scheduledAt))
  openId.value = feeding.id ?? ''
}

function selectDate(key: string) {
  selectedDate.value = key
  const dayFeedings = feedingsFor(key)
  const selected = chooseFeedingForDay(dayFeedings, Date.now())
  openId.value = selected?.id ?? ''
}

async function jumpToToday() {
  selectDate(today)
  await nextTick()
  scrollDateIntoView(today)
}

function scrollDateIntoView(key: string) {
  const target = calendar.value?.querySelector<HTMLElement>(`[data-day="${key}"]`)
  target?.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' })
}

async function extendCalendar(direction: 'before' | 'after') {
  if (extending || dates.value.length === 0) return
  extending = true
  const node = calendar.value
  const previousWidth = node?.scrollWidth ?? 0
  const previousScroll = node?.scrollLeft ?? 0
  const first = dates.value[0]
  const last = dates.value[dates.value.length - 1]
  const start = direction === 'before' ? addDays(first, -21) : addDays(last, 1)
  const newDates = fillDates(start, 21)

  if (direction === 'before') dates.value = [...newDates, ...dates.value]
  else dates.value = [...dates.value, ...newDates]
  await nextTick()
  if (direction === 'before' && node) node.scrollLeft = previousScroll + node.scrollWidth - previousWidth
  await fetchRange(start, addDays(start, 21))
  extending = false
}

function handleCalendarScroll() {
  const node = calendar.value
  if (!node || extending) return
  if (node.scrollLeft < 80) void extendCalendar('before')
  else if (node.scrollLeft + node.clientWidth > node.scrollWidth - 80) void extendCalendar('after')
}

async function toggleCompletion(feeding: FeedingFeeding) {
  if (!feeding.id) return
  changingId.value = feeding.id
  errorMessage.value = ''
  try {
    await updateCompletion(feeding.id, !feeding.completed)
    const current = feedings.value.find((item) => item.id === feeding.id)
    if (current) current.completed = !feeding.completed
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Could not update feeding.'
  } finally {
    changingId.value = ''
  }
}

function formatDateHeading(key: string): string {
  return formatDay(key, { weekday: 'long', month: 'short', day: 'numeric' })
}

function isSelectedToday(key: string): boolean {
  return key === today
}

let midnightRefresh: number | undefined

onMounted(async () => {
  const start = addDays(today, -7)
  dates.value = fillDates(start, 30)
  await fetchRange(start, addDays(start, 30))
  selectFeedingDay(chooseInitialFeeding(feedings.value, today, Date.now()))
  await nextTick()
  scrollDateIntoView(selectedDate.value)

  midnightRefresh = window.setInterval(() => {
    if (localDateKey(new Date()) !== today) window.location.reload()
  }, 60_000)
})

onBeforeUnmount(() => {
  if (midnightRefresh) window.clearInterval(midnightRefresh)
})
</script>

<template>
  <section class="space-y-5">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">{{ formatDateHeading(selectedDate) }}</h1>
      <button type="button" class="rounded-lg px-3 py-2 text-sm font-semibold text-emerald-700 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-slate-900" @click="jumpToToday">Today</button>
    </div>

    <div
      ref="calendar"
      class="flex snap-x snap-mandatory gap-2 overflow-x-auto overscroll-x-contain pb-2"
      @scroll.passive="handleCalendarScroll"
    >
      <button
        v-for="day in dates"
        :key="day"
        :data-day="day"
        type="button"
        class="flex min-w-12 snap-center flex-col items-center gap-1 rounded-xl px-2 py-2.5 transition-colors"
        :class="selectedDate === day ? 'bg-emerald-600 text-white' : 'text-slate-600 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-800'"
        :aria-label="`${formatDateHeading(day)}${isSelectedToday(day) ? ', today' : ''}`"
        :aria-pressed="selectedDate === day"
        @click="selectDate(day)"
      >
        <span class="text-[11px] font-medium uppercase">{{ formatDay(day, { weekday: 'short' }) }}</span>
        <span class="text-base font-semibold" :class="isSelectedToday(day) && selectedDate !== day ? 'text-emerald-700 dark:text-emerald-400' : ''">{{ formatDay(day, { day: 'numeric' }) }}</span>
        <span class="flex h-2 items-center gap-0.5">
          <span
            v-for="feeding in dotsFor(day)"
            :key="feeding.id"
            class="size-1.5 rounded-full"
            :class="feeding.completed ? 'bg-current' : 'border border-current'"
          />
        </span>
      </button>
    </div>

    <p v-if="errorMessage" role="alert" class="rounded-xl bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">{{ errorMessage }}</p>
    <p v-if="loading && feedings.length === 0" class="py-8 text-center text-sm text-slate-500">Loading…</p>

    <div v-else-if="selectedFeedings.length === 0" class="rounded-2xl border border-dashed border-slate-300 px-4 py-12 text-center text-sm text-slate-500 dark:border-slate-700 dark:text-slate-400">
      No feedings
    </div>

    <div v-else class="space-y-3">
      <article v-for="feeding in selectedFeedings" :key="feeding.id" class="overflow-hidden rounded-2xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
        <button
          type="button"
          class="flex w-full items-center justify-between gap-4 px-4 py-4 text-left"
          :aria-expanded="openId === feeding.id"
          @click="openId = openId === feeding.id ? '' : (feeding.id ?? '')"
        >
          <span class="flex items-center gap-3">
            <span class="size-3 rounded-full" :class="feeding.completed ? 'bg-emerald-600' : 'border-2 border-slate-400 dark:border-slate-500'" />
            <span class="font-semibold">{{ feeding.scheduledAt ? formatClock(feeding.scheduledAt) : '' }}</span>
            <span v-if="feeding.phase" class="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600 dark:bg-slate-800 dark:text-slate-300">{{ feeding.phase }}</span>
          </span>
          <span class="text-xl leading-none text-slate-400 transition-transform duration-300" :class="openId === feeding.id ? 'rotate-180' : ''" aria-hidden="true">⌄</span>
        </button>

        <div class="grid transition-[grid-template-rows] duration-300 ease-in-out" :class="openId === feeding.id ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'">
          <div class="min-h-0 overflow-hidden">
            <div class="space-y-4 border-t border-slate-100 px-4 pb-4 pt-3 dark:border-slate-800">
              <div class="space-y-2">
                <div v-for="(item, index) in feeding.items ?? []" :key="index" class="flex items-baseline justify-between gap-4">
                  <span v-if="item.name" class="truncate text-sm text-slate-600 dark:text-slate-300">{{ item.name }}</span>
                  <span v-else class="text-sm text-slate-600 dark:text-slate-300">{{ index === 0 && feeding.items?.length === 1 ? 'Food' : '' }}</span>
                  <span class="shrink-0 text-lg font-semibold tabular-nums">{{ formatQuantity(item.quantity ?? 0, item.unit) }} {{ item.unit }}</span>
                </div>
              </div>
              <button
                type="button"
                :disabled="changingId === feeding.id"
                class="w-full rounded-xl px-4 py-3 text-sm font-semibold transition-colors disabled:opacity-50"
                :class="feeding.completed ? 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700' : 'bg-emerald-600 text-white hover:bg-emerald-700'"
                @click="toggleCompletion(feeding)"
              >
                {{ feeding.completed ? 'Remove' : 'Complete' }}
              </button>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>
