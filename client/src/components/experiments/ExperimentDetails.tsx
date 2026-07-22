import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { MultiSelect } from "@/components/ui/multi-select"
import { fetchExperimentById, cancelExperiment } from '@/lib/api'
import { formatDuration } from '@/lib/formatters'
import { Button } from '@/components/ui/button'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { StopCircle } from 'lucide-react'

interface EquityPoint {
  Timestamp: string;
  Value: number;
}

export function ExperimentDetails({ experimentId }: { experimentId: string }) {
  const queryClient = useQueryClient()
  const [selectedCurves, setSelectedCurves] = useState<string[]>(['Average'])
  const [page, setPage] = useState(0)
  const pageSize = 10

  const { data: exp, isLoading } = useQuery({
    queryKey: ['experiment', experimentId],
    queryFn: () => fetchExperimentById(experimentId),
    refetchInterval: (query) => {
      // Auto refetch every 2s if it's not in a terminal state
      const expData = query.state.data as { status: string } | undefined;
      if (!expData || ['COMPLETED', 'FAILED', 'CANCELLED'].includes(expData.status)) return false;
      return 2000;
    }
  });

  const cancelMutation = useMutation({
    mutationFn: () => cancelExperiment(experimentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experiment', experimentId] })
      queryClient.invalidateQueries({ queryKey: ['experiments'] })
    }
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
         Object.values(equityCurves).forEach((curve: unknown) => {
           (curve as EquityPoint[]).forEach(pt => {
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
         (equityCurves[curveName] as EquityPoint[]).forEach((pt: EquityPoint) => {
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

  if (isLoading || !exp) {
    return (
      <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex items-center justify-center min-h-[400px]">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </Card>
    );
  }

  const logs = exp.status_logs || [];
  const resultSummary = exp.result_summary;
  const equityCurves = exp.equity_curves || {};
  const curveKeys = ['Average', ...Object.keys(equityCurves)];
  const COLORS = [
    '#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6',
    '#ec4899', '#06b6d4', '#14b8a6', '#f97316', '#6366f1',
  ];
  
  const isTerminal = ['COMPLETED', 'FAILED', 'CANCELLED'].includes(exp.status);

  return (
    <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex flex-col">
      <CardHeader className="flex flex-row items-start justify-between">
        <div>
          <CardTitle className="font-mono text-lg text-primary">{exp.name}</CardTitle>
          <CardDescription className="truncate">{exp.description || 'No description provided'}</CardDescription>
        </div>
        {!isTerminal && (
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
        )}
      </CardHeader>
      
      <CardContent className="flex-1 space-y-6">
        {exp.status === 'COMPLETED' && resultSummary && (
          <div className="grid grid-cols-3 gap-4">
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Total Return</div>
                <div className="text-xl font-bold">{resultSummary.TotalReturn?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Ann. Return</div>
                <div className="text-xl font-bold">{resultSummary.AnnualizedReturn?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Max Drawdown</div>
                <div className="text-xl font-bold text-red-500">{resultSummary.MaxDrawdown?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Win Rate</div>
                <div className="text-xl font-bold">{resultSummary.WinRate?.toFixed(2)}%</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Sharpe Ratio</div>
                <div className="text-xl font-bold">{resultSummary.SharpeRatio?.toFixed(2)}</div>
              </CardContent>
            </Card>
            <Card className="bg-black/20 border-border/50">
              <CardContent className="p-4 flex flex-col justify-center">
                <div className="text-xs text-muted-foreground font-mono uppercase mb-1">Total Trades</div>
                <div className="text-xl font-bold">{resultSummary.TotalTrades}</div>
              </CardContent>
            </Card>
          </div>
        )}

        {exp.status === 'COMPLETED' && Object.keys(equityCurves).length > 0 ? (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Equity Curve</h3>
              <div className="w-[300px]">
                <MultiSelect 
                  options={curveKeys.map(k => ({ label: k, value: k }))} 
                  selected={selectedCurves} 
                  onChange={setSelectedCurves}
                  placeholder="Select curves" 
                />
              </div>
            </div>
            <div className="w-full h-64 border border-border/50 rounded bg-black/20 p-4">
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
            <div className="flex items-center justify-between">
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
            <div className="rounded-md border border-border/50 overflow-hidden">
              <Table className="table-fixed w-full">
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
                  {([...resultSummary.JobMetrics] as { Symbol: string, Strategy: string, TotalReturn?: number, MaxDrawdown?: number, WinRate?: number }[])
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

        {/* Status Timeline */}
        <div className="space-y-2">
          <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Status Timeline</h3>
          <div className="rounded-md border border-border/50 overflow-hidden">
            <Table className="table-fixed w-full">
              <TableHeader className="bg-muted/50">
                <TableRow>
                  <TableHead className="font-mono uppercase text-xs">Step</TableHead>
                  <TableHead className="font-mono uppercase text-xs">Time Started</TableHead>
                  <TableHead className="font-mono uppercase text-xs text-right">Duration</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={3} className="text-center font-mono text-xs text-muted-foreground py-4">No logs available</TableCell>
                  </TableRow>
                ) : logs.map((log: { status: string; created_at: string }, index: number) => {
                  const nextLog = logs[index + 1];
                  const isTerminal = ['COMPLETED', 'FAILED'].includes(log.status);
                  const isLast = index === logs.length - 1;
                  
                  let durationDisplay = '-';
                  if (!isLast) {
                    durationDisplay = formatDuration(log.created_at, nextLog.created_at);
                  } else if (!isTerminal) {
                    durationDisplay = formatDuration(log.created_at); // from log start to now
                  }

                  return (
                    <TableRow key={index} className="hover:bg-primary/5 transition-colors">
                      <TableCell>
                        <span className={`inline-flex w-fit items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                          log.status === 'PENDING' ? 'bg-yellow-500/20 text-yellow-500' :
                          log.status === 'PENDING_INGESTION' ? 'bg-orange-500/20 text-orange-500' :
                          log.status === 'INGESTED' ? 'bg-purple-500/20 text-purple-500' :
                          log.status === 'READY' ? 'bg-indigo-500/20 text-indigo-500' :
                          log.status === 'RUNNING' ? 'bg-blue-500/20 text-blue-500' :
                          log.status === 'AGGREGATING' ? 'bg-cyan-500/20 text-cyan-500' :
                          log.status === 'COMPLETED' ? 'bg-green-500/20 text-green-500' :
                          'bg-red-500/20 text-red-500'
                        }`}>
                          {log.status.replace('_', ' ')}
                        </span>
                      </TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {(() => {
                          const d = new Date(log.created_at);
                          return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}:${d.getMilliseconds().toString().padStart(3, '0')}`;
                        })()}
                      </TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground text-right tabular-nums">
                        {durationDisplay}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
