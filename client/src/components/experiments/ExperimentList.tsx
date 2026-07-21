import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { SmoothProgress } from './SmoothProgress'

interface Experiment {
  id: string;
  name: string;
  description: string;
  status: string;
  created_at: string;
}

interface ExperimentListProps {
  isLoading: boolean;
  error: Error | null;
  experiments: Experiment[] | undefined;
  selectedExperimentId: string | null;
  setSelectedExperimentId: (id: string) => void;
  progressMap: Record<string, { totalJobs: number; completedJobs: number; stage: string }>;
}

export function ExperimentList({
  isLoading,
  error,
  experiments,
  selectedExperimentId,
  setSelectedExperimentId,
  progressMap
}: ExperimentListProps) {
  return (
    <Card className="border-primary/20 bg-black/40 backdrop-blur-md h-fit">
      <CardHeader>
        <CardTitle className="font-mono uppercase tracking-widest text-sm text-primary">Your Experiments</CardTitle>
        <CardDescription>A list of all backtesting runs you have initiated.</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center p-8">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        ) : error ? (
          <div className="text-red-500 font-mono text-sm p-4 border border-red-500/20 rounded bg-red-500/10">
            Error loading experiments. Please try logging in again.
          </div>
        ) : !experiments || experiments.length === 0 ? (
          <div className="text-center p-12 border border-dashed border-border/50 rounded-lg">
            <p className="text-muted-foreground font-mono mb-4">No experiments found.</p>
            <Link to="/create">
              <Button variant="outline" className="font-mono border-primary/50 text-primary">
                Create your first experiment
              </Button>
            </Link>
          </div>
        ) : (
          <div className="rounded-md border border-border/50 overflow-hidden">
            <Table className="table-fixed w-full">
              <TableHeader className="bg-muted/50">
                <TableRow>
                  <TableHead className="font-mono uppercase text-xs w-[25%]">Name</TableHead>
                  <TableHead className="font-mono uppercase text-xs w-[55%]">Status</TableHead>
                  <TableHead className="font-mono uppercase text-xs w-[20%] text-right">Created At</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {experiments.map((exp) => (
                  <TableRow 
                    key={exp.id} 
                    className={`cursor-pointer transition-colors ${selectedExperimentId === exp.id ? 'bg-muted/50 hover:bg-muted/70' : 'hover:bg-primary/5'}`}
                    onClick={() => setSelectedExperimentId(exp.id)}
                  >
                    <TableCell className="font-medium">
                      <div>{exp.name}</div>
                      <div className="text-xs text-muted-foreground font-mono mt-1 truncate">{exp.description}</div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col items-start gap-2">
                        <span className={`inline-flex w-fit items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                          exp.status === 'PENDING' ? 'bg-yellow-500/20 text-yellow-500' :
                          exp.status === 'PENDING_INGESTION' ? 'bg-orange-500/20 text-orange-500' :
                          exp.status === 'INGESTED' ? 'bg-purple-500/20 text-purple-500' :
                          exp.status === 'READY' ? 'bg-indigo-500/20 text-indigo-500' :
                          exp.status === 'RUNNING' ? 'bg-blue-500/20 text-blue-500' :
                          exp.status === 'AGGREGATING' ? 'bg-cyan-500/20 text-cyan-500' :
                          exp.status === 'COMPLETED' ? 'bg-green-500/20 text-green-500' :
                          'bg-red-500/20 text-red-500'
                        }`}>
                          {exp.status.replace('_', ' ')}
                        </span>
                        {progressMap[exp.id] && (
                          <div className="w-full mt-1">
                            <SmoothProgress 
                              stage={progressMap[exp.id].stage} 
                              percentage={Math.round((progressMap[exp.id].completedJobs / Math.max(1, progressMap[exp.id].totalJobs)) * 100)} 
                            />
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground text-right leading-tight">
                      {new Date(exp.created_at).toLocaleDateString()}<br/>
                      {new Date(exp.created_at).toLocaleTimeString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
