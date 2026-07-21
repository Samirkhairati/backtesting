import { useQuery } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { fetchExperimentById } from '@/lib/api'
import { formatDuration } from '@/lib/formatters'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

interface EquityPoint {
  Timestamp: string;
  Value: number;
}

export function ExperimentDetails({ experimentId }: { experimentId: string }) {
  const [selectedCurve, setSelectedCurve] = useState<string>('Average')

  const { data: exp, isLoading } = useQuery({
    queryKey: ['experiment', experimentId],
    queryFn: () => fetchExperimentById(experimentId),
    refetchInterval: (query) => {
      // Auto refetch every 2s if it's not in a terminal state
      const expData = query.state.data as { status: string } | undefined;
      if (!expData || expData.status === 'COMPLETED' || expData.status === 'FAILED') return false;
      return 2000;
    }
  });

  const chartData = useMemo(() => {
    if (!exp) return [];
    const equityCurves = exp.equity_curves || {};
    
    if (selectedCurve === 'Average' && Object.keys(equityCurves).length > 0) {
      const allTimestamps = new Set<string>();
      Object.values(equityCurves).forEach((curve: unknown) => {
        (curve as EquityPoint[]).forEach((pt) => allTimestamps.add(pt.Timestamp));
      });
      const sortedTimestamps = Array.from(allTimestamps).sort();
      return sortedTimestamps.map(ts => {
        let sum = 0;
        let count = 0;
        Object.values(equityCurves).forEach((curve: unknown) => {
          const pt = (curve as EquityPoint[]).find((p) => p.Timestamp === ts);
          if (pt) {
            sum += pt.Value;
            count++;
          }
        });
        return {
          Timestamp: new Date(ts).toLocaleDateString(),
          Value: count > 0 ? sum / count : 0
        };
      });
    } else if (equityCurves[selectedCurve]) {
      return (equityCurves[selectedCurve] as EquityPoint[]).map((pt) => ({
        Timestamp: new Date(pt.Timestamp).toLocaleDateString(),
        Value: pt.Value
      }));
    }
    return [];
  }, [exp, selectedCurve]);

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

  return (
    <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex flex-col">
      <CardHeader>
        <CardTitle className="font-mono text-lg text-primary">{exp.name}</CardTitle>
        <CardDescription className="truncate">{exp.description || 'No description provided'}</CardDescription>
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
              <Select value={selectedCurve} onValueChange={setSelectedCurve}>
                <SelectTrigger className="w-[200px] h-8 text-xs font-mono">
                  <SelectValue placeholder="Select curve" />
                </SelectTrigger>
                <SelectContent>
                  {curveKeys.map(key => (
                    <SelectItem key={key} value={key} className="text-xs font-mono">
                      {key}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
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
                  <Line type="monotone" dataKey="Value" stroke="#3b82f6" strokeWidth={2} dot={false} />
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
