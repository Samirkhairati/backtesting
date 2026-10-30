import { useState } from 'react'
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { SmoothProgress } from '@/components/experiments/SmoothProgress'
import { StageTimeline } from '@/components/experiments/StageTimeline'
import { statusBadgeClass, statusLabel } from '@/lib/stages'
import { cn } from '@/lib/utils'

// Everything below is static, illustrative sample data: no API calls are made.

const STRATEGIES = [
  { name: 'Buy and Hold', color: '#3b82f6', drift: 0.011, vol: 0.035, seed: 3 },
  { name: 'SMA Cross', color: '#10b981', drift: 0.0125, vol: 0.028, seed: 11 },
  { name: 'MACD', color: '#f59e0b', drift: 0.009, vol: 0.03, seed: 7 },
  { name: 'Bollinger Bands', color: '#ec4899', drift: 0.007, vol: 0.022, seed: 19 },
]

// Small seeded PRNG so the sample curves are identical on every render.
function mulberry32(seed: number) {
  let a = seed
  return () => {
    a = (a + 0x6d2b79f5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

function buildCurves() {
  const months = 60
  return Array.from({ length: months }, (_, i) => {
    const d = new Date(Date.UTC(2019, i, 1))
    return { label: d.toLocaleDateString('en-US', { month: 'short', year: '2-digit', timeZone: 'UTC' }) } as Record<
      string,
      number | string
    >
  }).map((point, i, arr) => {
    STRATEGIES.forEach((s) => {
      const rnd = mulberry32(s.seed + i * 31)
      const prev = i === 0 ? 100000 : (arr[i - 1][s.name] as number)
      point[s.name] = Math.round(prev * (1 + s.drift + (rnd() - 0.5) * 2 * s.vol))
    })
    return point
  })
}

const DATA = buildCurves()

const EXPERIMENTS = [
  { name: 'Tech momentum vs. hold', desc: 'NASDAQ 100 · 4 strategies', status: 'COMPLETED', date: 'Today, 09:41' },
  { name: 'S&P 500 mean reversion', desc: 'S&P 500 · RSI Mean Reversion', status: 'RUNNING', date: 'Today, 09:52', progress: 64, detail: '322/503 · 64%' },
  { name: 'MACD sweep 2015-2020', desc: 'Custom · MACD', status: 'PENDING_INGESTION', date: 'Today, 09:55', progress: 28, detail: '14/50 · 28%' },
]

const SUMMARY = [
  { label: 'Total Return', value: '+84.2%' },
  { label: 'Ann. Return', value: '+13.0%' },
  { label: 'Max Drawdown', value: '-18.7%', negative: true },
  { label: 'Win Rate', value: '58.4%' },
  { label: 'Sharpe Ratio', value: '1.12' },
  { label: 'Total Trades', value: '1,284' },
]

const BY_SYMBOL = [
  { symbol: 'AAPL', strategy: 'SMA Cross', ret: 112.4, dd: -14.2, win: 61.0 },
  { symbol: 'AMZN', strategy: 'Buy and Hold', ret: 96.8, dd: -24.1, win: 100 },
  { symbol: 'GOOGL', strategy: 'MACD', ret: 71.3, dd: -17.5, win: 55.2 },
  { symbol: 'MSFT', strategy: 'SMA Cross', ret: 138.9, dd: -12.8, win: 63.7 },
  { symbol: 'NVDA', strategy: 'Bollinger Bands', ret: 54.6, dd: -21.3, win: 52.1 },
]

// A finished run: timestamps are fixed offsets so the timeline shows real-looking stage durations.
const T0 = Date.UTC(2024, 0, 15, 9, 41, 0)
const at = (secs: number) => new Date(T0 + secs * 1000).toISOString()
const TIMELINE_EXPERIMENT = {
  status: 'COMPLETED',
  created_at: at(0),
  status_logs: [
    { status: 'PENDING', created_at: at(0) },
    { status: 'CREATING_INGESTION_JOBS', created_at: at(1) },
    { status: 'PENDING_INGESTION', created_at: at(3) },
    { status: 'CREATING_BACKTEST_JOBS', created_at: at(48) },
    { status: 'RUNNING', created_at: at(50) },
    { status: 'AGGREGATING', created_at: at(131) },
    { status: 'COMPLETED', created_at: at(134) },
  ],
  progress: { ingestion: { total: 100, completed: 100 }, backtesting: { total: 400, completed: 400 } },
}

export function SampleDashboard() {
  const [enabled, setEnabled] = useState<string[]>(STRATEGIES.map((s) => s.name))
  const toggle = (name: string) =>
    setEnabled((cur) => (cur.includes(name) ? cur.filter((n) => n !== name) : [...cur, name]))

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <Card className="h-fit min-w-0 border-primary/20 bg-black/40 backdrop-blur-md">
        <CardHeader>
          <CardTitle className="font-mono text-sm uppercase tracking-widest text-primary">Your Experiments</CardTitle>
          <CardDescription>Every run, with live status and progress.</CardDescription>
        </CardHeader>
        <CardContent>
          <ul className="space-y-3">
            {EXPERIMENTS.map((e, i) => (
              <li
                key={e.name}
                className={cn('space-y-2 rounded-md border border-border/50 p-3', i === 0 && 'bg-muted/50')}
              >
                <div className="flex items-start justify-between gap-2">
                  <span className="min-w-0 break-words font-medium">{e.name}</span>
                  <span className={cn('inline-flex shrink-0 rounded-full px-2.5 py-0.5 text-xs font-semibold', statusBadgeClass(e.status))}>
                    {statusLabel(e.status)}
                  </span>
                </div>
                <div className="truncate font-mono text-xs text-muted-foreground">{e.desc}</div>
                {e.progress !== undefined && (
                  <SmoothProgress
                    stage={e.status === 'RUNNING' ? 'Backtesting' : 'Ingesting'}
                    percentage={e.progress}
                    detail={e.detail}
                  />
                )}
                <div className="font-mono text-xs text-muted-foreground">{e.date}</div>
              </li>
            ))}
          </ul>
          <div className="mt-6">
            <StageTimeline experiment={TIMELINE_EXPERIMENT} />
          </div>
        </CardContent>
      </Card>

      <Card className="min-w-0 border-primary/20 bg-black/40 backdrop-blur-md">
        <CardHeader>
          <CardTitle className="font-mono text-lg text-primary">Tech momentum vs. hold</CardTitle>
          <CardDescription>NASDAQ 100 · 2019-2023 · 4 strategies</CardDescription>
        </CardHeader>
        <CardContent className="min-w-0 space-y-6">
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {SUMMARY.map((s) => (
              <Card key={s.label} className="border-border/50 bg-black/20">
                <CardContent className="p-4">
                  <div className="mb-1 font-mono text-xs uppercase text-muted-foreground">{s.label}</div>
                  <div className={cn('text-lg font-bold sm:text-xl', s.negative && 'text-red-500')}>{s.value}</div>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="space-y-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Equity Curve</h3>
              <div className="flex flex-wrap gap-1.5">
                {STRATEGIES.map((s) => {
                  const on = enabled.includes(s.name)
                  return (
                    <button
                      key={s.name}
                      type="button"
                      aria-pressed={on}
                      onClick={() => toggle(s.name)}
                      className={cn(
                        'flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 font-mono text-[11px] transition-colors',
                        on ? 'border-border text-foreground' : 'border-border/40 text-muted-foreground/60',
                      )}
                    >
                      <span className="h-2 w-2 rounded-full" style={{ background: on ? s.color : 'transparent', border: `1px solid ${s.color}` }} />
                      {s.name}
                    </button>
                  )
                })}
              </div>
            </div>
            <div className="h-64 w-full rounded border border-border/50 bg-black/20 p-2 sm:p-4">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={DATA}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="label" stroke="#888" fontSize={11} tickMargin={8} interval={11} />
                  <YAxis stroke="#888" fontSize={11} width={52} tickFormatter={(v) => `$${Math.round(v / 1000)}k`} domain={['auto', 'auto']} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#000', border: '1px solid #333' }}
                    itemStyle={{ color: '#fff' }}
                    labelStyle={{ color: '#888' }}
                    formatter={(v) => `$${Number(v).toLocaleString()}`}
                  />
                  {STRATEGIES.filter((s) => enabled.includes(s.name)).map((s) => (
                    <Line key={s.name} type="monotone" dataKey={s.name} stroke={s.color} strokeWidth={2} dot={false} isAnimationActive={false} />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="space-y-2">
            <h3 className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Results By Symbol</h3>
            <div className="overflow-x-auto rounded-md border border-border/50">
              <Table className="min-w-[440px] w-full">
                <TableHeader className="bg-muted/50">
                  <TableRow>
                    <TableHead className="font-mono text-xs uppercase">Symbol</TableHead>
                    <TableHead className="font-mono text-xs uppercase">Strategy</TableHead>
                    <TableHead className="text-right font-mono text-xs uppercase">Return %</TableHead>
                    <TableHead className="text-right font-mono text-xs uppercase">Max DD</TableHead>
                    <TableHead className="text-right font-mono text-xs uppercase">Win Rate</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {BY_SYMBOL.map((r) => (
                    <TableRow key={r.symbol}>
                      <TableCell className="font-medium">{r.symbol}</TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">{r.strategy}</TableCell>
                      <TableCell className="text-right font-mono text-xs tabular-nums">{r.ret.toFixed(2)}%</TableCell>
                      <TableCell className="text-right font-mono text-xs tabular-nums text-red-500">{r.dd.toFixed(2)}%</TableCell>
                      <TableCell className="text-right font-mono text-xs tabular-nums">{r.win.toFixed(2)}%</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
