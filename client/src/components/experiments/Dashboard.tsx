import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchExperiments, subscribeToExperimentEvents } from '@/lib/api'
import { LogOut, Plus, Activity } from 'lucide-react'
import { useState, useEffect } from 'react'
import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { ExperimentList } from './ExperimentList'
import { ExperimentDetails } from './ExperimentDetails'

export function Dashboard({ onLogout }: { onLogout: () => void }) {
  const queryClient = useQueryClient()
  const { data, isLoading, error } = useQuery({
    queryKey: ['experiments'],
    queryFn: fetchExperiments,
    retry: (failureCount, error) => {
      if (error.message === 'Unauthorized') return false; // Do not retry if token expired
      return failureCount < 3;
    }
  })

  const [selectedExperimentId, setSelectedExperimentId] = useState<string | null>(null)
  const [progressMap, setProgressMap] = useState<Record<string, { totalJobs: number; completedJobs: number; stage: string }>>({})

  useEffect(() => {
    if (error && error.message === 'Unauthorized') {
      onLogout();
    }
  }, [error, onLogout]);

  useEffect(() => {
    if (!data?.experiments) return;
    
    const abortControllers: Record<string, AbortController> = {};

    data.experiments.forEach((exp: { id: string; status: string }) => {
      if (exp.status === 'PENDING_INGESTION') {
        const controller = new AbortController();
        abortControllers[exp.id] = controller;
        
        subscribeToExperimentEvents(exp.id, (eventType, eventData) => {
          if (eventType === 'INGESTION_STARTED' || eventType === 'INGESTION_JOB_COMPLETED') {
            setProgressMap(prev => ({
              ...prev,
              [exp.id]: {
                totalJobs: eventData.total_jobs || 0,
                completedJobs: eventData.completed_jobs || 0,
                stage: 'Ingesting'
              }
            }));
          } else if (eventType === 'INGESTION_COMPLETED') {
            queryClient.invalidateQueries({ queryKey: ['experiments'] });
            if (exp.id === selectedExperimentId) {
              queryClient.invalidateQueries({ queryKey: ['experiment', exp.id] });
            }
            setProgressMap(prev => {
              const next = { ...prev };
              delete next[exp.id];
              return next;
            });
          }
        }, controller.signal);
      }
    });

    return () => {
      Object.values(abortControllers).forEach(ctrl => ctrl.abort());
    };
  }, [data?.experiments, queryClient, selectedExperimentId]);

  return (
    <div className="w-full max-w-5xl space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-primary flex items-center gap-2">
            <Activity className="h-8 w-8" />
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
          <Button variant="outline" onClick={onLogout} className="font-mono text-muted-foreground hover:text-foreground">
            <LogOut className="h-4 w-4 mr-2" />
            Logout
          </Button>
        </div>
      </div>

      <div className={`grid gap-6 ${selectedExperimentId ? 'grid-cols-2' : 'grid-cols-1'}`}>
        <ExperimentList
          isLoading={isLoading}
          error={error}
          experiments={data?.experiments}
          selectedExperimentId={selectedExperimentId}
          setSelectedExperimentId={setSelectedExperimentId}
          progressMap={progressMap}
        />

        {selectedExperimentId && (
          <div className="animate-in slide-in-from-right-8 duration-500 h-full">
            <ExperimentDetails experimentId={selectedExperimentId} />
          </div>
        )}
      </div>
    </div>
  )
}
