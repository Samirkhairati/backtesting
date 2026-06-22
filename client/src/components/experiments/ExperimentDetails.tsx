import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { fetchExperimentById } from '@/lib/api'
import { formatDuration } from '@/lib/formatters'

export function ExperimentDetails({ experimentId }: { experimentId: string }) {
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

  if (isLoading || !exp) {
    return (
      <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex items-center justify-center min-h-[400px]">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </Card>
    );
  }

  const logs = exp.status_logs || [];

  return (
    <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-full flex flex-col">
      <CardHeader>
        <CardTitle className="font-mono text-lg text-primary">{exp.name}</CardTitle>
        <CardDescription className="truncate">{exp.description || 'No description provided'}</CardDescription>
      </CardHeader>
      
      <CardContent className="flex-1 space-y-6">
        {/* Equity Curve Placeholder */}
        <div className="space-y-2">
          <h3 className="font-mono uppercase tracking-widest text-xs text-muted-foreground">Equity Curve</h3>
          <div className="w-full h-48 border border-dashed border-border/50 rounded flex items-center justify-center bg-black/20">
            <span className="font-mono text-sm text-muted-foreground/50">Chart not yet implemented</span>
          </div>
        </div>

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
