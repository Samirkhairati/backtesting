import { serverNow } from './useServerClock'

const pad = (n: number, width = 2) => n.toString().padStart(width, '0')

/**
 * Formats a duration in milliseconds with millisecond precision:
 *   845 ms, 12.345 s, 1m 02.345s, 1h 02m 02.345s
 * Negative or invalid values are clamped to 0.
 */
export function formatDurationMs(input: number): string {
  const ms = Number.isFinite(input) ? Math.max(0, Math.floor(input)) : 0
  if (ms < 1000) return `${ms} ms`
  if (ms < 60_000) return `${Math.floor(ms / 1000)}.${pad(ms % 1000, 3)} s`
  const totalSeconds = Math.floor(ms / 1000)
  const millis = ms % 1000
  const seconds = totalSeconds % 60
  const minutes = Math.floor(totalSeconds / 60) % 60
  const hours = Math.floor(totalSeconds / 3600)
  const secs = `${pad(seconds)}.${pad(millis, 3)}s`
  if (hours > 0) return `${hours}h ${pad(minutes)}m ${secs}`
  return `${minutes}m ${secs}`
}

/** Duration between two ISO timestamps; when `end` is omitted the (skew-corrected) server time is used. */
export function formatDuration(start: string, end?: string): string {
  const startMs = Date.parse(start)
  const endMs = end ? Date.parse(end) : serverNow()
  return formatDurationMs(endMs - startMs)
}

/** HH:MM:SS.mmm in local time. */
export function formatClock(ms: number): string {
  const d = new Date(ms)
  if (Number.isNaN(d.getTime())) return '-'
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}.${pad(d.getMilliseconds(), 3)}`
}
