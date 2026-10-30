// Shared API types for experiments.

export type KnownExperimentStatus =
  | 'PENDING'
  | 'CREATING_INGESTION_JOBS'
  | 'PENDING_INGESTION'
  | 'CREATING_BACKTEST_JOBS'
  | 'RUNNING'
  | 'AGGREGATING'
  | 'COMPLETED'
  | 'FAILED'
  | 'CANCELLED'
  // legacy values that may still exist on old rows
  | 'READY'
  | 'INGESTED'

// The backend may send values this build does not know about, so keep it open.
export type ExperimentStatus = KnownExperimentStatus | (string & {})

export interface StatusLog {
  status: ExperimentStatus
  created_at: string
}

export interface StageProgress {
  total: number
  completed: number
}

export interface ExperimentProgress {
  ingestion: StageProgress
  backtesting: StageProgress
}

export interface EquityPoint {
  Timestamp: string
  Value: number
}

export interface JobMetric {
  Symbol: string
  Strategy: string
  TotalReturn?: number
  MaxDrawdown?: number
  WinRate?: number
}

export interface ResultSummary {
  TotalReturn?: number
  AnnualizedReturn?: number
  MaxDrawdown?: number
  WinRate?: number
  SharpeRatio?: number
  TotalTrades?: number
  JobMetrics?: JobMetric[]
}

export interface Experiment {
  id: string
  name: string
  description: string
  config?: Record<string, unknown>
  status: ExperimentStatus
  created_at: string
  status_logs?: StatusLog[]
  progress?: ExperimentProgress
  server_time?: string
  result_summary?: ResultSummary
  equity_curves?: Record<string, EquityPoint[]>
}

export interface ExperimentsResponse {
  experiments: Experiment[]
  server_time?: string
}

export interface ExperimentEvent {
  experiment_id: string
  event_type: string
  symbol?: string
  total_jobs?: number
  completed_jobs?: number
  status?: ExperimentStatus
}

export interface SymbolInfo {
  symbol: string
  name: string
}
