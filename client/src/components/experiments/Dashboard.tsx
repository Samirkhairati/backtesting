import { useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query'
import { fetchExperiments, subscribeToExperimentEvents } from '@/lib/api'
import { Plus, Activity } from 'lucide-react'
import { useState, useEffect, useMemo, useRef } from 'react'
import { Link, useNavigate } from '@tanstack/react-router'
import { useAuth } from '@/lib/auth'
import { isTerminalStatus } from '@/lib/stages'
import type { ExperimentsResponse } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { ExperimentList } from './ExperimentList'
import { ExperimentDetails } from './ExperimentDetails'

// Events that always trigger a refetch (throttled to once per 250 ms per experiment).
const MAJOR_EVENTS = new Set([
  'STATUS_CHANGED',
  'EXPERIMENT_CANCELLED',
  'INGESTION_STARTED',
  'INGESTION_COMPLETED',
  'BACKTESTING_STARTED',
  'BACKTESTING_COMPLETED',
  'AGGREGATION_STARTED',
  'AGGREGATION_COMPLETED',
])
// High-frequency progress events (throttled to once per 500 ms per experiment).
const PROGRESS_EVENTS = new Set(['INGESTION_JOB_COMPLETED', 'BACKTESTING_JOB_COMPLETED', 'AGGREGATION_PROGRESS'])

const MAJOR_THROTTLE_MS = 250
const PROGRESS_THROTTLE_MS = 500

interface ThrottleState {
  last: number
  timer: ReturnType<typeof setTimeout> | null
}

interface Stream {
  controller: AbortController
  throttle: ThrottleState
}

/** Trailing throttle: at most one invalidation per `intervalMs`, and a pending one is never dropped. */
function scheduleInvalidate(queryClient: QueryClient, id: string, state: ThrottleState, intervalMs: number) {
  if (state.timer !== null) return
  const wait = Math.max(0, state.last + intervalMs - Date.now())
  state.timer = setTimeout(() => {
    state.timer = null
    state.last = Date.now()
    void queryClient.invalidateQueries({ queryKey: ['experiments'] })
    void queryClient.invalidateQueries({ queryKey: ['experiment', id] })
  }, wait)
}

function closeStream(stream: Stream) {
  stream.controller.abort()
  if (stream.throttle.timer !== null) clearTimeout(stream.throttle.timer)
  stream.throttle.timer = null
}

export function Dashboard() {
  const logout = useAuth((s) => s.logout)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data, isLoading, error } = useQuery<ExperimentsResponse, Error>({
    queryKey: ['experiments'],
    queryFn: fetchExperiments,
    refetchOnWindowFocus: true,
    refetchInterval: (query) => {
      const list = query.state.data?.experiments
      return list?.some((e) => !isTerminalStatus(e.status)) ? 2000 : false
    },
    retry: (failureCount, error) => {
      if (error.message === 'Unauthorized') return false; // Do not retry if token expired
      return failureCount < 3;
    }
  })

  const [selectedExperimentId, setSelectedExperimentId] = useState<string | null>(null)
  const streamsRef = useRef<Map<string, Stream>>(new Map())

  useEffect(() => {
    if (error && error.message === 'Unauthorized') {
      logout();
      navigate({ to: '/' });
    }
  }, [error, logout, navigate]);

  // Stable key of the non-terminal experiment ids, so the SSE effect only re-runs when that set changes
  // (not on every poll and not when the selected row changes).
  const activeIdsKey = useMemo(
    () =>
      (data?.experiments ?? [])
        .filter((e) => !isTerminalStatus(e.status))
        .map((e) => e.id)
        .sort()
        .join(','),
    [data?.experiments],
  )

  useEffect(() => {
    const wanted = new Set(activeIdsKey ? activeIdsKey.split(',') : [])
    const streams = streamsRef.current

    // Close streams for experiments that became terminal or disappeared.
    for (const [id, stream] of streams) {
      if (!wanted.has(id)) {
        closeStream(stream)
        streams.delete(id)
      }
    }

    // Open streams only for new experiments.
    for (const id of wanted) {
      if (streams.has(id)) continue
      const controller = new AbortController()
      const throttle: ThrottleState = { last: 0, timer: null }
      streams.set(id, { controller, throttle })

      subscribeToExperimentEvents(id, (eventType) => {
        if (MAJOR_EVENTS.has(eventType)) {
          scheduleInvalidate(queryClient, id, throttle, MAJOR_THROTTLE_MS)
        } else if (PROGRESS_EVENTS.has(eventType)) {
          scheduleInvalidate(queryClient, id, throttle, PROGRESS_THROTTLE_MS)
        }
      }, controller.signal)
    }
  }, [activeIdsKey, queryClient])

  // Close everything on unmount.
  useEffect(() => {
    const streams = streamsRef.current
    return () => {
      for (const stream of streams.values()) closeStream(stream)
      streams.clear()
    }
  }, [])

  return (
    <div className="w-full max-w-7xl mx-auto space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-primary flex items-center gap-2">
            <Activity className="h-6 w-6 sm:h-8 sm:w-8" />
            Dashboard
          </h1>
          <p className="text-muted-foreground font-mono text-sm">Manage your backtesting experiments.</p>
        </div>
        <div className="flex items-center gap-4">
          <Link to="/create">
            <Button variant="default" className="font-mono flex items-center gap-2">
              <Plus className="h-4 w-4" />
              New Experiment
            </Button>
          </Link>
        </div>
      </div>

      <div className={`grid gap-6 grid-cols-1 ${selectedExperimentId ? 'lg:grid-cols-2' : ''}`}>
        <div className="min-w-0">
          <ExperimentList
            isLoading={isLoading}
            error={error}
            experiments={data?.experiments}
            selectedExperimentId={selectedExperimentId}
            setSelectedExperimentId={setSelectedExperimentId}
          />
        </div>

        {selectedExperimentId && (
          <div className="animate-in slide-in-from-right-8 duration-500 h-full min-w-0">
            <ExperimentDetails experimentId={selectedExperimentId} />
          </div>
        )}
      </div>
    </div>
  )
}
