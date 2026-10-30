import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { MultiSelect } from "@/components/ui/multi-select"
import { fetchExperimentById, cancelExperiment } from '@/lib/api'
import { isTerminalStatus } from '@/lib/stages'
import { serverNow } from '@/lib/useServerClock'
import type { Experiment, ExperimentsResponse, JobMetric } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { StopCircle } from 'lucide-react'
import { StageTimeline } from './StageTimeline'

interface CancelContext {
  previousDetail: Experiment | undefined
  previousList: ExperimentsResponse | undefined
}

export function ExperimentDetails({ experimentId }: { experimentId: string }) {
  const queryClient = useQueryClient()
  const [selectedCurves, setSelectedCurves] = useState<string[]>(['Average'])
  const [page, setPage] = useState(0)
  const pageSize = 10

  const { data: exp, isLoading, isError, error, refetch, isFetching } = useQuery<Experiment, Error>({
    queryKey: ['experiment', experimentId],
    queryFn: () => fetchExperimentById(experimentId),
    // Poll every second while the experiment is still running; SSE only accelerates this.
    refetchInterval: (query) => {
      const expData = query.state.data
      if (expData && isTerminalStatus(expData.status)) return false
      return 1000
    },
  });

  const detailKey = ['experiment', experimentId] as const
  const listKey = ['experiments'] as const

  const cancelMutation = useMutation<unknown, Error, void, CancelContext>({
    mutationFn: () => cancelExperiment(experimentId),
    onMutate: async () => {
      await Promise.all([
        queryClient.cancelQueries({ queryKey: detailKey }),
        queryClient.cancelQueries({ queryKey: listKey }),
      ])
      const previousDetail = queryClient.getQueryData<Experiment>(detailKey)
      const previousList = queryClient.getQueryData<ExperimentsResponse>(listKey)

      queryClient.setQueryData<Experiment>(detailKey, (old) =>
        old
          ? {
              ...old,
              status: 'CANCELLED',
              status_logs: [
                ...(old.status_logs ?? []),
                { status: 'CANCELLED', created_at: new Date(serverNow()).toISOString() },
              ],
            }
          : old,
      )
      queryClient.setQueryData<ExperimentsResponse>(listKey, (old) =>
        old
          ? {
              ...old,
              experiments: old.experiments.map((e) =>
                e.id === experimentId ? { ...e, status: 'CANCELLED' } : e,
              ),
            }
          : old,
      )
      return { previousDetail, previousList }
    },
    onError: (_err, _vars, context) => {
      if (!context) return
      queryClient.setQueryData(detailKey, context.previousDetail)
      queryClient.setQueryData(listKey, context.previousList)
    },
    onSettled: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: detailKey }),
        queryClient.invalidateQueries({ queryKey: listKey }),
      ])
    },
  });

  const chartData = useMemo(() => {
    if (!exp || selectedCurves.length === 0) return [];
    const equityCurves = exp.equity_curves || {};
    
    // Map of timestamp -> data point
    const dataByTs: Record<string, Record<string, string | number>> = {};

    selectedCurves.forEach(curveName => {
      if (curveName === 'Average') {
         // calculate average curve once
         const allPts: Record<string, { sum: number, count: number }> = {};
         Object.values(equityCurves).forEach((curve) => {
           curve.forEach(pt => {
              if (!allPts[pt.Timestamp]) allPts[pt.Timestamp] = { sum: 0, count: 0 };
              allPts[pt.Timestamp].sum += pt.Value;
              allPts[pt.Timestamp].count += 1;
           });
         });
         Object.entries(allPts).forEach(([ts, { sum, count }]) => {
           if (!dataByTs[ts]) dataByTs[ts] = { Timestamp: ts };
           dataByTs[ts]['Average'] = sum / count;
         });
      } else if (equityCurves[curveName]) {
         equityCurves[curveName].forEach((pt) => {
           if (!dataByTs[pt.Timestamp]) dataByTs[pt.Timestamp] = { Timestamp: pt.Timestamp };
           dataByTs[pt.Timestamp][curveName] = pt.Value;
         });
      }
    });

    const sortedTimestamps = Object.keys(dataByTs).sort();
    return sortedTimestamps.map(ts => {
      return {
        ...dataByTs[ts],
        Timestamp: new Date(ts).toLocaleDateString(),
      };
    });
  }, [exp, selectedCurves]);

  if (!exp && isError) {
    return (
      <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex flex-col items-center justify-center gap-4 min-h-[400px] p-6 text-center">
        <p className="font-mono text-sm text-red-500">
          {error?.message || 'Failed to load experiment'}
        </p>
        <Button variant="outline" size="sm" className="font-mono" onClick={() => refetch()} disabled={isFetching}>
          {isFetching ? 'Retrying...' : 'Retry'}
        </Button>
      </Card>
    );
  }

  if (isLoading || !exp) {
    return (
      <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex items-center justify-center min-h-[400px]">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </Card>
    );
  }

  const resultSummary = exp.result_summary;
  const equityCurves = exp.equity_curves || {};
  const curveKeys = ['Average', ...Object.keys(equityCurves)];
  const COLORS = [
    '#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6',
    '#ec4899', '#06b6d4', '#14b8a6', '#f97316', '#6366f1',
  ];
  
  const isTerminal = isTerminalStatus(exp.status);

  return (
    <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex flex-col">
      <CardHeader className="flex flex-row items-start justify-between gap-3">
        <div className="min-w-0">
          <CardTitle className="font-mono text-lg text-primary break-words">{exp.name}</CardTitle>
          <CardDescription className="truncate">{exp.description || 'No description provided'}</CardDescription>
        </div>
        {!isTerminal && (
          <div className="flex flex-col items-end gap-1 shrink-0">
            <Button 
              variant="destructive" 
              size="sm" 
              className="font-mono text-xs gap-1"
              onClick={() => {
                if(confirm('Are you sure you want to cancel this experiment?')) {
                  cancelMutation.mutate()
                }
              }}
              disabled={cancelMutation.isPending}
            >
              <StopCircle className="h-4 w-4" />
              {cancelMutation.isPending ? 'Cancelling...' : 'Cancel'}
            </Button>
            {cancelMutation.isError && (
              <p className="max-w-[220px] text-right font-mono text-xs text-red-500">
                {cancelMutation.error.message || 'Failed to cancel experiment'}
              </p>
            )}
          </div>
        )}
      </CardHeader>
      
      <CardContent className="flex-1 min-w-0 space-y-6">
        {exp.status === 'COMPLETED' && resultSummary && (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4">
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Total Return</div>
                <div className="text-lg font-bold sm:text-xl">{resultSummary.TotalReturn?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Ann. Return</div>
                <div className="text-lg font-bold sm:text-xl">{resultSummary.AnnualizedReturn?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Max Drawdown</div>
                <div className="text-lg font-bold sm:text-xl text-red-500">{resultSummary.MaxDrawdown?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Win Rate</div>
                <div className="text-lg font-bold sm:text-xl">{resultSummary.WinRate?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Sharpe Ratio</div>
                <div className="text-lg font-bold sm:text-xl">{resultSummary.SharpeRatio?.toFixed(2)}</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Total Trades</div>
                <div className="text-lg font-bold sm:text-xl">{resultSummary.TotalTrades}</div>
              </CardContent>
            </Card>
          </div>
        )}

        {exp.status === 'COMPLETED' && Object.keys(equityCurves).length > 0 ? (
          <div className="space-y-2">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Equity Curve</h3>
              <div className="w-full sm:w-[300px] sm:max-w-[60%]">
                <MultiSelect 
                  options={curveKeys.map(k => ({ label: k, value: k }))} 
                  selected={selectedCurves} 
                  onChange={setSelectedCurves}
                  placeholder="Select curves" 
                />
              </div>
            </div>
            <div className="w-full h-64 border border-border/50 rounded bg-black/20 p-2 sm:p-4">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="Timestamp" stroke="#888" fontSize={12} tickMargin={10} />
                  <YAxis stroke="#888" fontSize={12} tickFormatter={(val) => `$${val.toLocaleString()}`} domain={['auto', 'auto']} />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#000', border: '1px solid #333' }}
                    itemStyle={{ color: '#fff' }}
                    labelStyle={{ color: '#888' }}
                  />
                  <Legend wrapperStyle={{ fontSize: '12px', color: '#888', paddingTop: '10px' }} />
                  {selectedCurves.map((curve, idx) => (
                    <Line 
                      key={curve} 
                      type="monotone" 
                      dataKey={curve} 
                      stroke={COLORS[idx % COLORS.length]} 
                      strokeWidth={2} 
                      dot={false} 
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        ) : (
          <div className="space-y-2">
            <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Equity Curve</h3>
            <div className="w-full h-48 border border-dashed border-border/50 rounded flex items-center justify-center bg-black/20">
              <span className="font-mono text-sm text-muted-foreground/50">
                {exp.status === 'COMPLETED' ? 'No curve data' : 'Chart will be available when completed'}
              </span>
            </div>
          </div>
        )}

        {exp.status === 'COMPLETED' && resultSummary?.JobMetrics && (
          <div className="space-y-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Results By Symbol</h3>
              <div className="flex gap-2">
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={() => setPage(p => Math.max(0, p - 1))}
                  disabled={page === 0}
                  className="font-mono text-xs h-7"
                >
                  Prev
                </Button>
                <span className="font-mono text-xs flex items-center text-muted-foreground">Page {page + 1}</span>
                <Button 
                  variant="outline" 
                  size="sm" 
                  onClick={() => setPage(p => p + 1)}
                  disabled={(page + 1) * pageSize >= resultSummary.JobMetrics.length}
                  className="font-mono text-xs h-7"
                >
                  Next
                </Button>
              </div>
            </div>
            <div className="rounded-md border border-border/50 overflow-x-auto">
              <Table className="w-full min-w-[480px]">
                <TableHeader className="bg-muted/50">
                  <TableRow>
                    <TableHead className="font-mono uppercase text-xs w-[20%]">Symbol</TableHead>
                    <TableHead className="font-mono uppercase text-xs w-[40%]">Strategy</TableHead>
                    <TableHead className="font-mono uppercase text-xs text-right">Return %</TableHead>
                    <TableHead className="font-mono uppercase text-xs text-right">Max Drawdown</TableHead>
                    <TableHead className="font-mono uppercase text-xs text-right">Win Rate</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {([...resultSummary.JobMetrics] as JobMetric[])
                    .sort((a, b) => a.Symbol.localeCompare(b.Symbol))
                    .slice(page * pageSize, (page + 1) * pageSize)
                    .map((job, idx) => (
                    <TableRow key={idx} className="hover:bg-primary/5 transition-colors">
                      <TableCell className="font-medium">{job.Symbol}</TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">{job.Strategy}</TableCell>
                      <TableCell className="text-right font-mono text-xs tabular-nums">{job.TotalReturn?.toFixed(2)}%</TableCell>
                      <TableCell className="text-right font-mono text-xs tabular-nums text-red-500">{job.MaxDrawdown?.toFixed(2)}%</TableCell>
                      <TableCell className="text-right font-mono text-xs tabular-nums">{job.WinRate?.toFixed(2)}%</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        )}

        <StageTimeline experiment={exp} />
      </CardContent>
    </Card>
  );
}
