import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Background,
  Handle,
  MarkerType,
  Position,
  ReactFlow,
  type Edge,
  type Node,
  type NodeProps,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { cn } from '@/lib/utils'
import { fetchWorkers } from '@/lib/api'
import { isTerminalStatus, stageForStatus, type StageKey } from '@/lib/stages'
import type { Experiment } from '@/lib/types'

const MAX_WORKERS_SHOWN = 8
const WORKERS_PER_ROW = 4
const WORKER_W = 64
const WORKER_H = 28
const GAP = 8
const GROUP_PAD = 10
const GROUP_TOP = 12
const GROUP_TITLE = 22
const GROUP_W = WORKERS_PER_ROW * WORKER_W + (WORKERS_PER_ROW - 1) * GAP + GROUP_PAD * 2

const ACTIVE_COLOR = '#f97316'
const IDLE_COLOR = '#52525b'

type NodeKind = 'box' | 'bus' | 'store' | 'dot' | 'group' | 'worker'

interface FlowNodeData extends Record<string, unknown> {
  label: string
  sub?: string
  kind: NodeKind
  active: boolean
  w: number
  h: number
  /** Extra connection points along the top/bottom edge, at `ratio` (0..1) of the width. */
  taps?: { id: string; side: 't' | 'b'; ratio: number }[]
}

/** Edges and nodes light up for the stages listed here. */
const ACTIVE_IN: Record<string, StageKey[]> = {
  client_api: ['PENDING', 'COMPLETED'],
  api_bus: ['PENDING', 'CREATING_INGESTION_JOBS'],
  coord_bus: ['CREATING_INGESTION_JOBS', 'CREATING_BACKTEST_JOBS', 'AGGREGATING'],
  ingestion: ['PENDING_INGESTION'],
  backtest: ['RUNNING'],
  coord_pg: ['AGGREGATING'],
}

function FlowNode({ data }: NodeProps<Node<FlowNodeData>>) {
  const { label, sub, kind, active, w, h } = data
  const side = (id: string, position: Position) => (
    <>
      <Handle id={`${id}-in`} type="target" position={position} className="!opacity-0 !h-1 !w-1" />
      <Handle id={`${id}-out`} type="source" position={position} className="!opacity-0 !h-1 !w-1" />
    </>
  )

  if (kind === 'dot') {
    return (
      <div
        style={{ width: w, height: h }}
        className={cn('rounded-full', active ? 'bg-orange-500 shadow-[0_0_10px_rgba(249,115,22,0.9)]' : 'bg-zinc-600')}
      >
        {side('t', Position.Top)}
        {side('b', Position.Bottom)}
      </div>
    )
  }

  return (
    <div
      style={{ width: w, height: h }}
      className={cn(
        'flex flex-col items-center font-mono transition-colors duration-300',
        kind === 'group'
          ? 'justify-end rounded-md border border-dashed bg-white/[0.03] pb-1'
          : 'justify-center rounded border bg-black/60',
        kind === 'bus' && 'rounded-full',
        kind === 'store' && 'rounded-b-xl',
        active
          ? 'border-orange-500 text-orange-400 shadow-[0_0_12px_rgba(249,115,22,0.35)]'
          : 'border-zinc-600 text-zinc-400',
      )}
    >
      <span className={cn('leading-tight', kind === 'worker' ? 'text-[10px]' : 'text-[11px]')}>{label}</span>
      {sub && <span className="text-[9px] leading-tight text-zinc-500">{sub}</span>}
      {side('t', Position.Top)}
      {side('b', Position.Bottom)}
      {side('l', Position.Left)}
      {side('r', Position.Right)}
      {data.taps?.map((tap) => (
        <span key={tap.id}>
          <Handle
            id={`${tap.id}-in`}
            type="target"
            position={tap.side === 't' ? Position.Top : Position.Bottom}
            style={{ left: `${tap.ratio * 100}%` }}
            className="!opacity-0 !h-1 !w-1"
          />
          <Handle
            id={`${tap.id}-out`}
            type="source"
            position={tap.side === 't' ? Position.Top : Position.Bottom}
            style={{ left: `${tap.ratio * 100}%` }}
            className="!opacity-0 !h-1 !w-1"
          />
        </span>
      ))}
    </div>
  )
}

const nodeTypes = { flow: FlowNode }

function workerCount(n: number | undefined): number {
  return Math.min(MAX_WORKERS_SHOWN, Math.max(0, n ?? 0))
}

function buildGraph(
  stage: StageKey | undefined,
  counts: { ingestion: number; backtesting: number },
  loaded: boolean,
) {
  const on = (key: keyof typeof ACTIVE_IN) => !!stage && ACTIVE_IN[key].includes(stage)
  const ingestionOn = on('ingestion')
  const backtestOn = on('backtest')

  const nodes: Node<FlowNodeData>[] = []
  const edges: Edge[] = []

  const node = (
    id: string,
    x: number,
    y: number,
    data: FlowNodeData,
    extra: Partial<Node<FlowNodeData>> = {},
  ) =>
    nodes.push({
      id,
      type: 'flow',
      position: { x, y },
      data,
      draggable: false,
      selectable: false,
      connectable: false,
      ...extra,
    })

  const edge = (
    id: string,
    source: string,
    target: string,
    active: boolean,
    opts: {
      sourceHandle?: string
      targetHandle?: string
      label?: string
      bidirectional?: boolean
      type?: string
    } = {},
  ) => {
    const color = active ? ACTIVE_COLOR : IDLE_COLOR
    edges.push({
      id,
      source,
      target,
      sourceHandle: opts.sourceHandle,
      targetHandle: opts.targetHandle,
      type: opts.type ?? 'smoothstep',
      animated: active,
      zIndex: 1,
      label: opts.label,
      labelStyle: { fill: active ? '#fdba74' : '#a1a1aa', fontSize: 10, fontFamily: 'monospace' },
      labelBgStyle: { fill: '#09090b', fillOpacity: 0.9 },
      labelBgPadding: [4, 2],
      style: { stroke: color, strokeWidth: active ? 2 : 1.25, opacity: active ? 1 : 0.55 },
      markerEnd: { type: MarkerType.ArrowClosed, color },
      ...(opts.bidirectional ? { markerStart: { type: MarkerType.ArrowClosed, color } } : {}),
    })
  }

  const ingN = workerCount(counts.ingestion)
  const btN = workerCount(counts.backtesting)
  const rows = Math.max(1, Math.ceil(Math.max(ingN, btN, 1) / WORKERS_PER_ROW))
  const groupH = GROUP_TOP + rows * (WORKER_H + GAP) - GAP + GROUP_TITLE

  const totalW = GROUP_W * 2 + 40
  const ingX = 0
  const btX = GROUP_W + 40
  const ingCx = ingX + GROUP_W / 2
  const btCx = btX + GROUP_W / 2
  const busW = 380
  const busX = (totalW - busW) / 2
  const groupY = 215
  const pgY = groupY + groupH + 45

  node('client', 20, 0, { label: 'Client', sub: 'React', kind: 'box', active: on('client_api'), w: 110, h: 40 })
  node('api', 210, 0, {
    label: 'API', sub: 'Go · Gin', kind: 'box', w: 110, h: 40,
    active: on('client_api') || on('api_bus'),
  })
  node('coord', totalW - 150, 0, { label: 'Coordinator', sub: 'Go', kind: 'box', active: on('coord_bus') || on('coord_pg'), w: 130, h: 40 })
  node('bus', busX, 88, {
    label: 'Redpanda / Kafka', kind: 'bus', w: busW, h: 34,
    active: on('api_bus') || on('coord_bus') || ingestionOn || backtestOn,
    taps: [
      { id: 'api', side: 't', ratio: (210 + 55 - busX) / busW },
      { id: 'coord', side: 't', ratio: (totalW - 85 - busX) / busW },
      { id: 'ing', side: 'b', ratio: (ingCx - busX) / busW },
      { id: 'bt', side: 'b', ratio: (btCx - busX) / busW },
    ],
  })
  node('ing-dot', ingCx - 7, 160, { label: '', kind: 'dot', active: ingestionOn, w: 14, h: 14 })
  node('bt-dot', btCx - 7, 160, { label: '', kind: 'dot', active: backtestOn, w: 14, h: 14 })

  const group = (
    id: string, x: number, title: string, active: boolean, n: number, accent: boolean,
  ) => {
    node(id, x, groupY, { label: title, kind: 'group', active: accent, w: GROUP_W, h: groupH }, { zIndex: -1 })
    const shown = Math.max(1, n)
    for (let i = 0; i < shown; i++) {
      const col = i % WORKERS_PER_ROW
      const row = Math.floor(i / WORKERS_PER_ROW)
      // Center a partial last row.
      const inRow = Math.min(WORKERS_PER_ROW, shown - row * WORKERS_PER_ROW)
      const rowW = inRow * WORKER_W + (inRow - 1) * GAP
      const startX = x + (GROUP_W - rowW) / 2
      const placeholder = n === 0
      node(
        `${id}-w${i}`,
        startX + col * (WORKER_W + GAP),
        groupY + GROUP_TOP + row * (WORKER_H + GAP),
        {
          label: placeholder ? (loaded ? 'none' : '…') : `#${i + 1}`,
          kind: 'worker',
          active: active && !placeholder,
          w: WORKER_W,
          h: WORKER_H,
        },
        { zIndex: 2 },
      )
    }
    return shown
  }

  const ingShown = group('ing', ingX, 'Ingestion workers · Python', ingestionOn, ingN, ingestionOn)
  const btShown = group('bt', btX, 'Backtesting workers · Go', backtestOn, btN, backtestOn)

  node('pg', (totalW - 220) / 2, pgY, { label: 'PostgreSQL', kind: 'store', active: ingestionOn || backtestOn || on('coord_pg'), w: 220, h: 38 })

  edge('e-client-api', 'client', 'api', on('client_api'), {
    sourceHandle: 'r-out', targetHandle: 'l-in', label: 'REST + SSE', bidirectional: true, type: 'straight',
  })
  edge('e-api-bus', 'api', 'bus', on('api_bus'), { sourceHandle: 'b-out', targetHandle: 'api-in' })
  edge('e-coord-bus', 'coord', 'bus', on('coord_bus'), {
    sourceHandle: 'b-out', targetHandle: 'coord-in', bidirectional: true,
  })
  edge('e-bus-ing', 'bus', 'ing-dot', ingestionOn, { sourceHandle: 'ing-out', targetHandle: 't-in', label: 'ingestion jobs' })
  edge('e-bus-bt', 'bus', 'bt-dot', backtestOn, { sourceHandle: 'bt-out', targetHandle: 't-in', label: 'backtesting jobs' })

  for (let i = 0; i < ingShown; i++) {
    edge(`e-ing-w${i}`, 'ing-dot', `ing-w${i}`, ingestionOn && ingN > 0, { sourceHandle: 'b-out', targetHandle: 't-in' })
  }
  for (let i = 0; i < btShown; i++) {
    edge(`e-bt-w${i}`, 'bt-dot', `bt-w${i}`, backtestOn && btN > 0, { sourceHandle: 'b-out', targetHandle: 't-in' })
  }

  edge('e-ing-pg', 'ing', 'pg', ingestionOn, { sourceHandle: 'b-out', targetHandle: 'l-in' })
  edge('e-bt-pg', 'bt', 'pg', backtestOn, { sourceHandle: 'b-out', targetHandle: 'r-in' })
  edge('e-coord-pg', 'coord', 'pg', on('coord_pg'), { sourceHandle: 'r-out', targetHandle: 'r-in', bidirectional: true })

  return { nodes, edges }
}

const CAPTION: Partial<Record<StageKey, string>> = {
  PENDING: 'Request received by the API',
  CREATING_INGESTION_JOBS: 'Coordinator is queueing ingestion jobs',
  PENDING_INGESTION: 'Ingestion workers are downloading market data',
  CREATING_BACKTEST_JOBS: 'Coordinator is queueing backtesting jobs',
  RUNNING: 'Backtesting workers are running strategies',
  AGGREGATING: 'Coordinator is aggregating results',
  COMPLETED: 'Done. Results are served by the API',
}

export function PipelineDiagram({ experiment }: { experiment: Pick<Experiment, 'status'> }) {
  const stage = stageForStatus(experiment.status)?.key
  const terminal = isTerminalStatus(experiment.status)

  const { data, isSuccess } = useQuery({
    queryKey: ['workers'],
    queryFn: fetchWorkers,
    refetchInterval: 5000,
  })

  const counts = useMemo(
    () => ({ ingestion: data?.ingestion ?? 0, backtesting: data?.backtesting ?? 0 }),
    [data],
  )
  const { nodes, edges } = useMemo(
    () => buildGraph(terminal && stage !== 'COMPLETED' ? undefined : stage, counts, isSuccess),
    [stage, terminal, counts, isSuccess],
  )

  const caption =
    experiment.status === 'FAILED'
      ? 'Experiment failed'
      : experiment.status === 'CANCELLED'
        ? 'Experiment cancelled'
        : stage
          ? CAPTION[stage]
          : undefined

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Pipeline</h3>
        <div className="font-mono text-xs text-muted-foreground">
          {counts.ingestion} ingestion · {counts.backtesting} backtesting workers online
        </div>
      </div>
      <div className="h-[380px] w-full overflow-hidden rounded-md border border-border/50 bg-black/20">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          colorMode="dark"
          fitView
          fitViewOptions={{ padding: 0.08 }}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable={false}
          panOnDrag={false}
          zoomOnScroll={false}
          zoomOnPinch={false}
          zoomOnDoubleClick={false}
          preventScrolling={false}
          proOptions={{ hideAttribution: true }}
        >
          <Background gap={20} size={1} color="#27272a" />
        </ReactFlow>
      </div>
      {caption && <p className="font-mono text-[11px] text-muted-foreground">{caption}</p>}
    </div>
  )
}
