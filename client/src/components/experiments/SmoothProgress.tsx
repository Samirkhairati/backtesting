import { cn } from '@/lib/utils'

interface SmoothProgressProps {
  /** Left-hand label (e.g. the stage name). Omit for a bare bar. */
  stage?: string
  /** 0-100. Ignored when `indeterminate`. */
  percentage: number
  /** Right-hand text, defaults to the percentage. */
  detail?: string
  /** Show a pulsing full-width bar (unknown total / no measurable progress). */
  indeterminate?: boolean
  /** Muted/red/etc. colour override for the fill. */
  barClassName?: string
  className?: string
}

export function SmoothProgress({
  stage,
  percentage,
  detail,
  indeterminate = false,
  barClassName,
  className,
}: SmoothProgressProps) {
  const pct = Math.min(100, Math.max(0, Number.isFinite(percentage) ? percentage : 0))
  const showHeader = stage !== undefined || detail !== undefined || !indeterminate

  return (
    <div className={cn('w-full text-xs font-mono', className)}>
      {showHeader && (
        <div className="flex justify-between gap-2 mb-1 text-muted-foreground">
          <span className="truncate">{stage}</span>
          <span className="tabular-nums shrink-0">{detail ?? (indeterminate ? '' : `${Math.round(pct)}%`)}</span>
        </div>
      )}
      <div
        className="h-1.5 w-full bg-secondary rounded-full overflow-hidden"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={indeterminate ? undefined : Math.round(pct)}
      >
        <div
          className={cn(
            'h-full bg-primary transition-[width] duration-500 ease-out',
            indeterminate && 'animate-pulse opacity-60',
            barClassName,
          )}
          style={{ width: indeterminate ? '100%' : `${pct}%` }}
        />
      </div>
    </div>
  )
}
