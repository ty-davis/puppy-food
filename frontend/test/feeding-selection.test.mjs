import assert from 'node:assert/strict'
import test from 'node:test'
import { localDateKey } from '../src/lib/dates.ts'
import { chooseFeedingForDay, chooseInitialFeeding } from '../src/lib/feeding-selection.ts'

const todayDate = new Date(2026, 3, 15, 12)
const today = localDateKey(todayDate)

function feeding(id, date, completed = false) {
  return { id, scheduledAt: date.toISOString(), completed }
}

function at(dayOffset, hour, minute = 0) {
  const date = new Date(todayDate)
  date.setDate(date.getDate() + dayOffset)
  date.setHours(hour, minute, 0, 0)
  return date
}

test('initial selection picks the next upcoming incomplete feeding today', () => {
  const now = at(0, 10)
  const feedings = [
    feeding('past', at(0, 9)),
    feeding('completed-upcoming', at(0, 11), true),
    feeding('next-today', at(0, 12)),
    feeding('tomorrow', at(1, 8)),
  ]

  assert.equal(chooseInitialFeeding(feedings, today, now.getTime())?.id, 'next-today')
})

test('initial selection falls forward to the next day when no feeding remains today', () => {
  const now = at(0, 10)
  const feedings = [
    feeding('past', at(0, 9)),
    feeding('tomorrow', at(1, 8)),
  ]

  assert.equal(chooseInitialFeeding(feedings, today, now.getTime())?.id, 'tomorrow')
})

test('when all of today’s feedings are complete, initial selection opens the one closest to now', () => {
  const now = at(0, 10)
  const feedings = [
    feeding('closest-today', at(0, 9, 45), true),
    feeding('farther-today', at(0, 7), true),
    feeding('tomorrow', at(1, 8)),
  ]

  assert.equal(chooseInitialFeeding(feedings, today, now.getTime())?.id, 'closest-today')
})

test('day selection chooses the next upcoming incomplete feeding for that day', () => {
  const now = at(0, 10)
  const feedings = [
    feeding('past-incomplete', at(0, 9)),
    feeding('completed-upcoming', at(0, 11), true),
    feeding('next-incomplete', at(0, 12)),
  ]

  assert.equal(chooseFeedingForDay(feedings, now.getTime())?.id, 'next-incomplete')
})

test('day selection falls back to the feeding closest to now when none remain upcoming', () => {
  const now = at(0, 10)
  const feedings = [
    feeding('past', at(0, 9)),
    feeding('closest', at(0, 10, 15), true),
    feeding('farther', at(0, 12), true),
  ]

  assert.equal(chooseFeedingForDay(feedings, now.getTime())?.id, 'closest')
})
