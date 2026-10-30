import { useEffect, useMemo, useState } from 'react'
import { Check, Minus, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { formatClock, formatDurationMs } from '@/lib/formatters'
import { serverNow } from '@/lib/useServerClock'
import {
  buildTimeline,
  getStageProgress,
  progressKindForStage,
  type RowState,
  type TimelineRow,
} from '@/lib/stages'
import type { Experiment } from '@/lib/types'
import { SmoothProgress } from './SmoothProgress'

/**
 * Self-ticking duration. While `endMs` is undefined it re-renders on every animation
 * frame (only this tiny component), measuring against the skew-corrected server clock.
 */
export function LiveDuration({
  startMs,
  endMs,
  className,
}: {
  startMs: number | undefined
  endMs?: number
  className?: string
}) {
  const live = startMs !== undefined && endMs === undefined
  const [now, setNow] = useState(() => serverNow())

  useEffect(() => {
    if (!live) return
    let raf = 0
    const tick = () => {
      setNow(serverNow())
      raf = requestAnimationFrame(tick)
    }
    raf = requestAnimationFrame(tick)
    // rAF is paused in background tabs; a slow interval keeps the value fresh there.
    const interval = window.setInterval(() => setNow(serverNow()), 250)
    return () => {
      cancelAnimationFrame(raf)
      window.clearInterval(interval)
    }
  }, [live])

  const text =
    startMs === undefined ? '-' : formatDurationMs((endMs ?? now) - startMs)

  return (
    <span className={cn('inline-block min-w-[11ch] text-right tabular-nums whitespace-nowrap', className)}>
      {text}
    </span>
  )
}

function StateIcon({ state }: { state: RowState }) {
  const base = 'flex h-5 w-5 shrink-0 items-center justify-center rounded-full'
  switch (state) {
    case 'done':
      return (
        <span className={cn(base, 'bg-green-500/20 text-green-500')}>
          <Check className="h-3 w-3" />
        </span>
      )
    case 'active':
      return (
        <span className={cn(base)}>
          <span className="h-2.5 w-2.5 rounded-full bg-orange-500 shadow-[0_0_8px_rgba(249,115,22,0.8)] animate-pulse" />
        </span>
      )
    case 'cancelled':
      return (
        <span className={cn(base, 'bg-amber-500/20 text-amber-500')}>
          <X className="h-3 w-3" />
        </span>
      )
    case 'failed':
      return (
        <span className={cn(base, 'bg-red-500/20 text-red-500')}>
          <X className="h-3 w-3" />
        </span>
      )
    case 'skipped':
      return (
        <span className={cn(base, 'text-muted-foreground/50')}>
          <Minus className="h-3 w-3" />
        </span>
      )
    default:
      return (
        <span className={cn(base)}>
          <span className="h-2 w-2 rounded-full border border-muted-foreground/40" />
        </span>
      )
  }
}

function RowProgress({ row, progress }: { row: TimelineRow; progress: Experiment['progress'] }) {
  const kind = progressKindForStage(row.key)
  if (!kind) return null
  const { total, completed } = getStageProgress(progress, kind)
  const unit = kind === 'ingestion' ? 'symbols' : 'jobs'
  const detail = `${completed} / ${total} ${unit}`

  if (row.state === 'done') {
    return total > 0 ? <div className="mt-1 text-[11px] text-muted-foreground tabular-nums">{detail}</div> : null
  }
  if (row.state === 'upcoming' || row.state === 'skipped') return null

  const frozen = row.state === 'cancelled' || row.state === 'failed'
  const indeterminate = total === 0 && !frozen
  return (
    <div className="mt-2 max-w-xs">
      <SmoothProgress
        percentage={total > 0 ? (completed / total) * 100 : 0}
        detail={total > 0 ? detail : 'preparing...'}
        indeterminate={indeterminate}
        barClassName={cn(
          row.state === 'cancelled' && 'bg-amber-500',
          row.state === 'failed' && 'bg-red-500',
        )}
      />
    </div>
  )
}

function RowDuration({ row }: { row: TimelineRow }) {
  switch (row.state) {
    case 'active':
      return <LiveDuration startMs={row.startMs} className="text-orange-500" />
    case 'done':
    case 'cancelled':
    case 'failed': {
      const color = row.state === 'cancelled' ? 'text-amber-500' : row.state === 'failed' ? 'text-red-500' : 'text-foreground'
      return (
        <span className={cn('inline-block min-w-[11ch] text-right tabular-nums whitespace-nowrap', color)}>
          {row.durationMs === undefined ? '-' : formatDurationMs(row.durationMs)}
        </span>
      )
    }
    case 'skipped':
      return <span className="text-muted-foreground/60">skipped</span>
    default:
      return <span className="text-muted-foreground/60">-</span>
  }
}

function StageRow({ row, progress }: { row: TimelineRow; progress: Experiment['progress'] }) {
  const muted = row.state === 'upcoming' || row.state === 'skipped'
  const note =
    row.state === 'cancelled' ? 'cancelled' : row.state === 'failed' ? 'failed' : row.isTotal ? 'total run time' : null

  return (
    <li
      className={cn(
        'flex items-start gap-3 px-3 py-2.5 border-b border-border/40 last:border-b-0',
        row.state === 'active' && 'bg-orange-500/5',
      )}
    >
      <div className="pt-0.5">
        <StateIcon state={row.state} />
      </div>
      <div className="min-w-0 flex-1">
        <div
          className={cn(
            'text-sm font-mono leading-5',
            muted && 'text-muted-foreground/60',
            row.state === 'active' && 'text-orange-500',
            row.state === 'cancelled' && 'text-amber-500',
            row.state === 'failed' && 'text-red-500',
          )}
        >
          {row.label}
        </div>
        {(row.startMs !== undefined || note) && (
          <div className="text-[11px] font-mono text-muted-foreground tabular-nums">
            {row.startMs !== undefined && formatClock(row.startMs)}
            {row.startMs !== undefined && note && ' · '}
            {note}
          </div>
        )}
        <RowProgress row={row} progress={progress} />
      </div>
      <div className="shrink-0 font-mono text-xs leading-5 text-right pt-0.5">
        <RowDuration row={row} />
      </div>
    </li>
  )
}

export function StageTimeline({
  experiment,
}: {
  experiment: Pick<Experiment, 'status' | 'status_logs' | 'created_at' | 'progress'>
}) {
  const { status, status_logs, created_at, progress } = experiment
  const timeline = useMemo(
    () => buildTimeline({ status, status_logs, created_at }),
    [status, status_logs, created_at],
  )

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-2">
        <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Status Timeline</h3>
        <div className="font-mono text-xs text-muted-foreground flex items-center gap-2">
          <span>Total</span>
          <LiveDuration
            startMs={timeline.startMs}
            endMs={timeline.live ? undefined : (timeline.endMs ?? timeline.startMs)}
            className={cn(timeline.live ? 'text-orange-500' : 'text-foreground')}
          />
        </div>
      </div>
      <ol className="rounded-md border border-border/50 overflow-hidden bg-black/10">
        {timeline.rows.map((row) => (
          <StageRow key={row.key} row={row} progress={progress} />
        ))}
      </ol>
    </div>
  )
}
