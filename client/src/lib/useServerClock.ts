// Tracks the offset between the server clock and the browser clock so that live
// durations computed against server-provided timestamps are not skewed.

let offsetMs = 0

/** Record the server's clock from a response's `server_time` (latest wins). */
export function recordServerTime(serverTime: string | undefined | null): void {
  if (!serverTime) return
  const parsed = Date.parse(serverTime)
  if (Number.isNaN(parsed)) return
  offsetMs = parsed - Date.now()
}

/** Current time on the server's clock, in epoch milliseconds. */
export function serverNow(): number {
  return Date.now() + offsetMs
}

/** Hook-style accessor, handy for components that prefer a hook. */
export function useServerClock(): { serverNow: () => number } {
  return { serverNow }
}
