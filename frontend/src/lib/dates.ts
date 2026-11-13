export function localDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function dateFromKey(key: string): Date {
  const [year, month, day] = key.split('-').map(Number)
  return new Date(year, month - 1, day)
}

export function addDays(key: string, amount: number): string {
  const date = dateFromKey(key)
  date.setDate(date.getDate() + amount)
  return localDateKey(date)
}

export function localMidnightUtc(key: string): string {
  return `${dateFromKey(key).toISOString()}`
}

export function utcForLocalTime(key: string, clock: string): string {
  const [hours, minutes] = clock.split(':').map(Number)
  const date = dateFromKey(key)
  date.setHours(hours, minutes, 0, 0)
  return date.toISOString()
}

export function formatDay(key: string, options: Intl.DateTimeFormatOptions): string {
  return new Intl.DateTimeFormat(undefined, options).format(dateFromKey(key))
}

export function formatClock(iso: string): string {
  return new Intl.DateTimeFormat(undefined, { hour: 'numeric', minute: '2-digit' }).format(new Date(iso))
}

export function browserTimeZone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
}

export function formatQuantity(quantity: number): string {
	return quantity.toFixed(3).replace(/(\.\d*?[1-9])0+$|\.0+$/, '$1')
}
