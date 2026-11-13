import type { FeedingFeeding } from '../api/generated/types.gen'
import { localDateKey } from './dates.ts'

function scheduledTime(feeding: FeedingFeeding): number {
  return feeding.scheduledAt ? new Date(feeding.scheduledAt).getTime() : Number.NaN
}

function validFeedings(feedings: FeedingFeeding[]): FeedingFeeding[] {
  return feedings.filter((feeding) => Number.isFinite(scheduledTime(feeding)))
}

function nextUpcoming(feedings: FeedingFeeding[], now: number): FeedingFeeding | undefined {
  return validFeedings(feedings)
    .filter((feeding) => !feeding.completed && scheduledTime(feeding) >= now)
    .sort((a, b) => scheduledTime(a) - scheduledTime(b))[0]
}

function closestToNow(feedings: FeedingFeeding[], now: number): FeedingFeeding | undefined {
  return validFeedings(feedings)
    .sort((a, b) => Math.abs(scheduledTime(a) - now) - Math.abs(scheduledTime(b) - now))[0]
}

export function chooseFeedingForDay(feedings: FeedingFeeding[], now: number): FeedingFeeding | undefined {
  return nextUpcoming(feedings, now) ?? closestToNow(feedings, now)
}

export function chooseInitialFeeding(
  feedings: FeedingFeeding[],
  today: string,
  now: number,
): FeedingFeeding | undefined {
  const valid = validFeedings(feedings)
  const todayFeedings = valid.filter((feeding) => localDateKey(new Date(scheduledTime(feeding))) === today)

  const upcomingToday = nextUpcoming(todayFeedings, now)
  if (upcomingToday) return upcomingToday

  if (todayFeedings.length > 0 && todayFeedings.every((feeding) => feeding.completed)) {
    return closestToNow(todayFeedings, now)
  }

  const upcoming = nextUpcoming(valid, now)
  return upcoming ?? closestToNow(todayFeedings, now)
}
