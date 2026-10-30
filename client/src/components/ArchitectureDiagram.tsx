import { useState } from 'react'
import { cn } from '@/lib/utils'

interface Node {
  id: string
  label: string
  sub: string
  x: number
  y: number
  w: number
  h: number
  kind: 'app' | 'bus' | 'store'
  title: string
  body: string
}

const NODES: Node[] = [
  {
    id: 'client', label: 'Client', sub: 'React + Vite', x: 20, y: 10, w: 150, h: 56, kind: 'app',
    title: 'Client (React)',
    body: 'The web UI. Creates experiments, lists them and shows live progress, stage timings, equity curves and per-symbol results. Talks only to the API over REST and a Server-Sent Events stream.',
  },
  {
    id: 'api', label: 'API', sub: 'Go · Gin', x: 250, y: 10, w: 150, h: 56, kind: 'app',
    title: 'API server (Go, Gin)',
    body: 'Authenticates users with JWT, validates and stores new experiments, publishes the first job to the event bus, and serves experiment status, results and the live event stream.',
  },
  {
    id: 'bus', label: 'Redpanda (Kafka)', sub: 'event bus & job queue', x: 20, y: 120, w: 760, h: 50, kind: 'bus',
    title: 'Redpanda (Kafka-compatible)',
    body: 'Asynchronous backbone. Carries ingestion jobs, backtesting jobs, progress events and worker heartbeats, so every service scales and fails independently and workers pull jobs at their own pace.',
  },
  {
    id: 'coordinator', label: 'Coordinator', sub: 'Go · lifecycle', x: 40, y: 230, w: 180, h: 56, kind: 'app',
    title: 'Coordinator (Go)',
    body: 'Drives the experiment lifecycle: splits an experiment into ingestion and backtesting jobs, advances the stage as jobs finish, monitors worker heartbeats, handles cancellation, and aggregates final metrics and equity curves.',
  },
  {
    id: 'ingestion', label: 'Ingestion Workers', sub: 'Python · scalable', x: 300, y: 230, w: 200, h: 56, kind: 'app',
    title: 'Ingestion workers (Python)',
    body: 'Download and normalise historical market data for each symbol, then store it in PostgreSQL. A Redis cache avoids re-downloading data that has already been ingested.',
  },
  {
    id: 'backtest', label: 'Backtesting Workers', sub: 'Go · scalable', x: 580, y: 230, w: 200, h: 56, kind: 'app',
    title: 'Backtesting workers (Go)',
    body: 'Run a strategy (Buy and Hold, SMA Cross, MACD, RSI Mean Reversion, Bollinger Bands) over one symbol at a time through a simulated broker and portfolio, then report metrics and the equity curve. Add more workers to run more symbols in parallel.',
  },
  {
    id: 'postgres', label: 'PostgreSQL', sub: 'source of truth', x: 20, y: 340, w: 400, h: 50, kind: 'store',
    title: 'PostgreSQL',
    body: 'Durable storage for users, experiments, status history, job state, market data and results. The source of truth that every service reads from and writes to.',
  },
  {
    id: 'redis', label: 'Redis', sub: 'market-data cache', x: 440, y: 340, w: 340, h: 50, kind: 'store',
    title: 'Redis',
    body: 'In-memory cache for market data, so repeated experiments over the same symbols and date ranges skip slow database reads and re-downloads.',
  },
]

interface Edge {
  id: string
  d: string
  ends: [string, string]
  /** Draw behind the nodes (used for the API -> PostgreSQL line that passes behind the bus). */
  both?: boolean
}

const EDGES: Edge[] = [
  { id: 'client-api', d: 'M170 38 H250', ends: ['client', 'api'], both: true },
  { id: 'api-bus', d: 'M330 66 V120', ends: ['api', 'bus'] },
  { id: 'api-pg', d: 'M262 66 V340', ends: ['api', 'postgres'] },
  { id: 'bus-coord', d: 'M130 170 V230', ends: ['bus', 'coordinator'], both: true },
  { id: 'bus-ing', d: 'M400 170 V230', ends: ['bus', 'ingestion'], both: true },
  { id: 'bus-bt', d: 'M680 170 V230', ends: ['bus', 'backtest'], both: true },
  { id: 'coord-pg', d: 'M110 286 V340', ends: ['coordinator', 'postgres'] },
  { id: 'ing-pg', d: 'M350 286 V340', ends: ['ingestion', 'postgres'] },
  { id: 'ing-redis', d: 'M470 286 V340', ends: ['ingestion', 'redis'] },
  { id: 'bt-redis', d: 'M650 286 V340', ends: ['backtest', 'redis'] },
  { id: 'bt-pg', d: 'M600 286 L400 340', ends: ['backtest', 'postgres'] },
]

const KIND_STYLE: Record<Node['kind'], string> = {
  app: 'fill-black/60 stroke-primary/40',
  bus: 'fill-primary/10 stroke-primary/60',
  store: 'fill-black/60 stroke-border',
}

export function ArchitectureDiagram() {
  const [activeId, setActiveId] = useState<string | null>(null)
  const active = NODES.find((n) => n.id === activeId) ?? null

  return (
    <div className="space-y-4">
      <div className="overflow-x-auto">
        <svg
          viewBox="0 0 800 400"
          role="group"
          aria-label="System architecture diagram. Focus or hover a component to read what it does."
          className="mx-auto h-auto w-full min-w-[560px] max-w-4xl"
          onMouseLeave={() => setActiveId(null)}
        >
          <defs>
            <marker id="arch-arrow" viewBox="0 0 8 8" refX="6" refY="4" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M0 0 L8 4 L0 8 z" fill="currentColor" />
            </marker>
          </defs>

          {EDGES.map((e) => {
            const on = activeId !== null && e.ends.includes(activeId)
            return (
              <path
                key={e.id}
                d={e.d}
                fill="none"
                strokeWidth={on ? 2 : 1.25}
                strokeDasharray={on ? undefined : '4 4'}
                markerEnd="url(#arch-arrow)"
                markerStart={e.both ? 'url(#arch-arrow)' : undefined}
                className={cn(
                  'transition-colors',
                  on ? 'stroke-primary text-primary' : 'stroke-muted-foreground/40 text-muted-foreground/40',
                  activeId !== null && !on && 'opacity-40',
                )}
              />
            )
          })}

          {NODES.map((n) => {
            const on = activeId === n.id
            return (
              <g
                key={n.id}
                tabIndex={0}
                role="button"
                aria-label={`${n.title}: ${n.body}`}
                onMouseEnter={() => setActiveId(n.id)}
                onFocus={() => setActiveId(n.id)}
                onBlur={() => setActiveId(null)}
                onClick={() => setActiveId(n.id)}
                className="cursor-pointer outline-none"
              >
                <rect
                  x={n.x}
                  y={n.y}
                  width={n.w}
                  height={n.h}
                  rx={4}
                  strokeWidth={on ? 2 : 1}
                  className={cn(KIND_STYLE[n.kind], 'transition-colors', on && 'stroke-primary')}
                  style={on ? { filter: 'drop-shadow(0 0 6px hsl(var(--primary) / 0.6))' } : undefined}
                />
                <text
                  x={n.x + n.w / 2}
                  y={n.y + n.h / 2 - 3}
                  textAnchor="middle"
                  className="fill-foreground font-mono text-[13px] font-semibold"
                >
                  {n.label}
                </text>
                <text
                  x={n.x + n.w / 2}
                  y={n.y + n.h / 2 + 14}
                  textAnchor="middle"
                  className="fill-muted-foreground font-mono text-[10px]"
                >
                  {n.sub}
                </text>
              </g>
            )
          })}
        </svg>
      </div>

      <div className="min-h-[5.5rem] border border-primary/20 bg-black/40 p-4 backdrop-blur-md" aria-live="polite">
        {active ? (
          <>
            <div className="mb-1 font-mono text-sm font-semibold text-primary">{active.title}</div>
            <p className="font-mono text-xs leading-relaxed text-muted-foreground">{active.body}</p>
          </>
        ) : (
          <p className="font-mono text-xs leading-relaxed text-muted-foreground">
            Hover, focus or tap a component to see what it does. Create an experiment and the API stores it and
            publishes a job; the coordinator fans it out to ingestion workers, then backtesting workers, and finally
            aggregates the results, with every step streamed live to your browser.
          </p>
        )}
      </div>
    </div>
  )
}
