import type { Experiment, ExperimentProgress, StageProgress, StatusLog } from './types'

export type StageKey =
  | 'PENDING'
  | 'CREATING_INGESTION_JOBS'
  | 'PENDING_INGESTION'
  | 'CREATING_BACKTEST_JOBS'
  | 'RUNNING'
  | 'AGGREGATING'
  | 'COMPLETED'

export interface StageMeta {
  key: StageKey
  label: string
  /** Legacy status values that are treated as this stage. */
  aliases: string[]
  /** Tailwind classes for the status badge. */
  badge: string
}

export const STAGES: readonly StageMeta[] = [
  { key: 'PENDING', label: 'Pending', aliases: ['READY'], badge: 'bg-yellow-500/20 text-yellow-500' },
  { key: 'CREATING_INGESTION_JOBS', label: 'Creating ingestion jobs', aliases: [], badge: 'bg-indigo-500/20 text-indigo-500' },
  { key: 'PENDING_INGESTION', label: 'Ingesting', aliases: [], badge: 'bg-orange-500/20 text-orange-500' },
  { key: 'CREATING_BACKTEST_JOBS', label: 'Creating backtesting jobs', aliases: ['INGESTED'], badge: 'bg-purple-500/20 text-purple-500' },
  { key: 'RUNNING', label: 'Backtesting', aliases: [], badge: 'bg-blue-500/20 text-blue-500' },
  { key: 'AGGREGATING', label: 'Aggregating', aliases: [], badge: 'bg-cyan-500/20 text-cyan-500' },
  { key: 'COMPLETED', label: 'Completed', aliases: [], badge: 'bg-green-500/20 text-green-500' },
]

export const TERMINAL_STATUSES = ['COMPLETED', 'FAILED', 'CANCELLED'] as const

export function isTerminalStatus(status: string | undefined | null): boolean {
  return !!status && (TERMINAL_STATUSES as readonly string[]).includes(status)
}

const STAGE_INDEX = new Map<string, number>()
STAGES.forEach((s, i) => {
  STAGE_INDEX.set(s.key, i)
  s.aliases.forEach((a) => STAGE_INDEX.set(a, i))
})

/** Index of the stage a status belongs to (legacy aliases resolved), or -1 for non-stage statuses. */
export function stageIndexOf(status: string | undefined | null): number {
  if (!status) return -1
  return STAGE_INDEX.get(status) ?? -1
}

export function stageForStatus(status: string | undefined | null): StageMeta | undefined {
  const i = stageIndexOf(status)
  return i >= 0 ? STAGES[i] : undefined
}

const FAILED_BADGE = 'bg-red-500/20 text-red-500'
const CANCELLED_BADGE = 'bg-amber-500/20 text-amber-500'
const UNKNOWN_BADGE = 'bg-muted text-muted-foreground'

export function statusLabel(status: string | undefined | null): string {
  if (!status) return 'Unknown'
  if (status === 'FAILED') return 'Failed'
  if (status === 'CANCELLED') return 'Cancelled'
  const stage = stageForStatus(status)
  if (stage) return stage.label
  const text = status.replace(/_/g, ' ').toLowerCase()
  return text.charAt(0).toUpperCase() + text.slice(1)
}

export function statusBadgeClass(status: string | undefined | null): string {
  if (status === 'FAILED') return FAILED_BADGE
  if (status === 'CANCELLED') return CANCELLED_BADGE
  return stageForStatus(status)?.badge ?? UNKNOWN_BADGE
}

const EMPTY_PROGRESS: StageProgress = { total: 0, completed: 0 }

export function getStageProgress(
  progress: ExperimentProgress | undefined,
  kind: 'ingestion' | 'backtesting',
): StageProgress {
  const p = progress?.[kind]
  if (!p) return EMPTY_PROGRESS
  const total = Number.isFinite(p.total) && p.total > 0 ? p.total : 0
  const completed = Number.isFinite(p.completed) && p.completed > 0 ? p.completed : 0
  return { total, completed: total > 0 ? Math.min(completed, total) : completed }
}

/** Which progress counter (if any) belongs to a stage. */
export function progressKindForStage(key: StageKey): 'ingestion' | 'backtesting' | null {
  if (key === 'PENDING_INGESTION') return 'ingestion'
  if (key === 'RUNNING') return 'backtesting'
  return null
}

// ---------------------------------------------------------------------------
// Timeline

export type RowState = 'done' | 'active' | 'upcoming' | 'skipped' | 'cancelled' | 'failed'

export interface TimelineRow {
  key: StageKey
  label: string
  state: RowState
  /** When the stage began (epoch ms, server clock), if known. */
  startMs?: number
  /** Final duration for finished/frozen rows. Undefined for active/upcoming/skipped. */
  durationMs?: number
  /** True for the Completed row, whose duration is the total run time. */
  isTotal?: boolean
}

export interface Timeline {
  rows: TimelineRow[]
  /** When the experiment started (epoch ms), if known. */
  startMs?: number
  /** When it stopped (epoch ms). Undefined while still running (total is live). */
  endMs?: number
  /** True when the total / active stage should tick live. */
  live: boolean
}

interface Entry {
  status: string
  /** Stage index, or -1 for terminal markers (FAILED / CANCELLED) and unknown statuses. */
  stage: number
  ms: number
  /** Original position in the logs, used to keep the sort stable on equal timestamps. */
  order: number
}

function normalizeLogs(logs: StatusLog[] | undefined): Entry[] {
  const entries: Entry[] = []
  ;(logs ?? []).forEach((log, order) => {
    if (!log || typeof log.status !== 'string') return
    const ms = Date.parse(log.created_at)
    if (Number.isNaN(ms)) return
    entries.push({ status: log.status, stage: stageIndexOf(log.status), ms, order })
  })
  // Stable sort by time; ties keep the original order.
  entries.sort((a, b) => a.ms - b.ms || a.order - b.order)
  // Collapse consecutive entries that map to the same stage (e.g. READY then PENDING).
  const out: Entry[] = []
  for (const e of entries) {
    const prev = out[out.length - 1]
    if (prev && e.stage >= 0 && prev.stage === e.stage) continue
    out.push(e)
  }
  return out
}

/**
 * Builds the per-stage rows for an experiment from its status logs.
 * Pure: durations of active rows are not computed here (they are live), see `LiveDuration`.
 */
export function buildTimeline(
  exp: Pick<Experiment, 'status' | 'status_logs' | 'created_at'>,
): Timeline {
  const entries = normalizeLogs(exp.status_logs)
  const status = exp.status
  const terminal = isTerminalStatus(status)
  const currentStage = stageIndexOf(status)
  const createdMs = Date.parse(exp.created_at)

  // If the current status is a stage that has no log yet (no logs, or logs lagging behind),
  // add a synthetic entry so the right row is active.
  if (currentStage >= 0 && !entries.some((e) => e.stage === currentStage)) {
    const last = entries[entries.length - 1]
    const ms = last ? last.ms : Number.isNaN(createdMs) ? NaN : createdMs
    if (!Number.isNaN(ms)) entries.push({ status, stage: currentStage, ms, order: entries.length })
  }

  const lastEntry = entries[entries.length - 1]
  const startMs = entries.length > 0 ? entries[0].ms : Number.isNaN(createdMs) ? undefined : createdMs

  const rows: TimelineRow[] = STAGES.map((meta, i) => {
    const base = { key: meta.key, label: meta.label }
    const idx = entries.findIndex((e) => e.stage === i)
    if (idx < 0) {
      const laterHasEntry = entries.some((e) => e.stage > i)
      return { ...base, state: laterHasEntry ? 'skipped' : 'upcoming' } satisfies TimelineRow
    }
    const entry = entries[idx]
    const next = entries[idx + 1]

    if (meta.key === 'COMPLETED') {
      const total = startMs !== undefined ? Math.max(0, entry.ms - startMs) : undefined
      return { ...base, state: 'done', startMs: entry.ms, durationMs: total, isTotal: true } satisfies TimelineRow
    }

    if (next) {
      const durationMs = Math.max(0, next.ms - entry.ms)
      if (next.stage < 0 && next.status === 'CANCELLED') {
        return { ...base, state: 'cancelled', startMs: entry.ms, durationMs } satisfies TimelineRow
      }
      if (next.stage < 0 && next.status === 'FAILED') {
        return { ...base, state: 'failed', startMs: entry.ms, durationMs } satisfies TimelineRow
      }
      return { ...base, state: 'done', startMs: entry.ms, durationMs } satisfies TimelineRow
    }

    // Last entry for the experiment.
    if (!terminal) {
      return { ...base, state: 'active', startMs: entry.ms } satisfies TimelineRow
    }
    if (status === 'CANCELLED' || status === 'FAILED') {
      // Terminal status without a terminal log entry: freeze at the last known moment.
      return {
        ...base,
        state: status === 'CANCELLED' ? 'cancelled' : 'failed',
        startMs: entry.ms,
        durationMs: 0,
      } satisfies TimelineRow
    }
    return { ...base, state: 'done', startMs: entry.ms, durationMs: 0 } satisfies TimelineRow
  })

  let endMs: number | undefined
  if (terminal && lastEntry) endMs = lastEntry.ms

  return { rows, startMs, endMs, live: !terminal && startMs !== undefined }
}
