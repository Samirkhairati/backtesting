package coordinator

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"go.uber.org/zap"
)

type ExperimentService struct {
	Repo     *ExperimentRepository
	Producer kafka.Publisher
}

func NewExperimentService(repo *ExperimentRepository, producer kafka.Publisher) *ExperimentService {
	return &ExperimentService{
		Repo:     repo,
		Producer: producer,
	}
}

// maxConcurrentPublishes bounds parallel Kafka produce calls when fanning out jobs.
const maxConcurrentPublishes = 32

// maxSymbolLength mirrors the VARCHAR(15) symbol columns.
const maxSymbolLength = 15

// CreateExperiment stores the experiment and returns immediately. Everything slow (coverage check,
// creating thousands of jobs, publishing them) runs in the background so the caller is never
// blocked and the experiment's progress is visible from the first moment.
func (s *ExperimentService) CreateExperiment(ctx context.Context, userID, name, description, configJSON string) (string, string, error) {
	experimentID, err := s.Repo.InsertExperiment(ctx, userID, name, description, configJSON)
	if err != nil {
		logger.Log.Error("Failed to insert experiment", zap.Error(err))
		return "", "", err
	}

	var config struct {
		Symbols    []string `json:"symbols"`
		Timeframes []struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"timeframes"`
	}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return experimentID, string(models.ExperimentStatusPending), err
	}

	if len(config.Symbols) == 0 || len(config.Timeframes) == 0 {
		return experimentID, string(models.ExperimentStatusReady), nil
	}

	// The data we need spans all requested timeframes.
	reqStart, reqEnd := "", ""
	for _, tf := range config.Timeframes {
		st, err1 := time.Parse(time.RFC3339, tf.Start)
		en, err2 := time.Parse(time.RFC3339, tf.End)
		if err1 != nil || err2 != nil {
			continue
		}
		sd, ed := st.Format("2006-01-02"), en.Format("2006-01-02")
		if reqStart == "" || sd < reqStart {
			reqStart = sd
		}
		if ed > reqEnd {
			reqEnd = ed
		}
	}
	if reqStart == "" || reqEnd == "" {
		return experimentID, string(models.ExperimentStatusPending), nil
	}

	go s.prepareExperiment(experimentID, uniqueStrings(config.Symbols), reqStart, reqEnd)

	logger.Log.Info("Experiment created", zap.String("id", experimentID))
	return experimentID, string(models.ExperimentStatusPending), nil
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// publishStatusChanged tells connected clients to refresh immediately.
func (s *ExperimentService) publishStatusChanged(ctx context.Context, experimentID string, status models.ExperimentStatus) {
	if s.Producer == nil {
		return
	}
	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.Producer.Publish(pubCtx, "event-stream", experimentID, models.ExperimentEvent{
		ExperimentID: experimentID,
		EventType:    models.EventTypeStatusChanged,
		Status:       string(status),
	}); err != nil {
		logger.Log.Warn("Failed to publish status change", zap.Error(err), zap.String("id", experimentID))
	}
}

// Transition moves an experiment between statuses only if it is still in `from`. Because the
// update is conditional, a cancelled experiment can never be resurrected by a late worker.
func (s *ExperimentService) Transition(ctx context.Context, experimentID string, from, to models.ExperimentStatus) bool {
	ok, err := s.Repo.TryTransitionExperimentStatus(ctx, experimentID, from, to)
	if err != nil {
		logger.Log.Error("Failed to transition experiment", zap.Error(err), zap.String("id", experimentID), zap.String("to", string(to)))
		return false
	}
	if ok {
		s.publishStatusChanged(ctx, experimentID, to)
	}
	return ok
}

// publishAll publishes messages concurrently (bounded) and waits for all of them.
func (s *ExperimentService) publishAll(ctx context.Context, topic string, msgs []keyedMessage) {
	if s.Producer == nil {
		return
	}
	sem := make(chan struct{}, maxConcurrentPublishes)
	var wg sync.WaitGroup
	for _, m := range msgs {
		sem <- struct{}{}
		wg.Add(1)
		go func(m keyedMessage) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := s.Producer.Publish(ctx, topic, m.key, m.value); err != nil {
				logger.Log.Error("Failed to publish message", zap.Error(err), zap.String("topic", topic), zap.String("key", m.key))
			}
		}(m)
	}
	wg.Wait()
}

type keyedMessage struct {
	key   string
	value any
}

// prepareExperiment decides which symbols still need data, creates ingestion jobs for them, or goes
// straight to backtesting when everything is already stored.
func (s *ExperimentService) prepareExperiment(experimentID string, symbols []string, reqStart, reqEnd string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	coverage, err := s.Repo.GetMarketDataCoverageBatch(ctx, symbols)
	if err != nil {
		// Safe fallback: treat everything as missing. The ingestion worker re-checks coverage
		// before downloading, so this costs a no-op job per symbol, not a download.
		logger.Log.Error("Failed to load market data coverage", zap.Error(err), zap.String("id", experimentID))
		coverage = map[string][2]string{}
	}

	var missing []string
	for _, sym := range symbols {
		// market data keys are varchar(15); anything longer cannot exist or be ingested.
		if len(sym) > maxSymbolLength {
			logger.Log.Warn("Skipping invalid symbol", zap.String("symbol", sym), zap.String("id", experimentID))
			continue
		}
		cov, ok := coverage[sym]
		if !ok || reqStart < cov[0] || reqEnd > cov[1] {
			missing = append(missing, sym)
		}
	}

	if len(missing) == 0 {
		if s.Transition(ctx, experimentID, models.ExperimentStatusPending, models.ExperimentStatusCreatingBacktestJobs) {
			if err := s.GenerateBacktestJobs(ctx, experimentID); err != nil {
				logger.Log.Error("Failed to generate backtest jobs", zap.Error(err), zap.String("id", experimentID))
				s.Transition(ctx, experimentID, models.ExperimentStatusCreatingBacktestJobs, models.ExperimentStatusFailed)
			}
		}
		return
	}

	if !s.Transition(ctx, experimentID, models.ExperimentStatusPending, models.ExperimentStatusCreatingIngestionJobs) {
		return // cancelled before we started
	}

	jobs, err := s.Repo.InsertIngestionJobs(ctx, experimentID, missing, reqStart, reqEnd)
	if err != nil {
		logger.Log.Error("Failed to insert ingestion jobs", zap.Error(err), zap.String("id", experimentID))
		s.Transition(ctx, experimentID, models.ExperimentStatusCreatingIngestionJobs, models.ExperimentStatusFailed)
		return
	}

	// Switch to "ingesting" BEFORE publishing so a worker finishing instantly cannot race the
	// status change. If it fails, the experiment was cancelled while we were creating jobs.
	if !s.Transition(ctx, experimentID, models.ExperimentStatusCreatingIngestionJobs, models.ExperimentStatusPendingIngestion) {
		if err := s.Repo.CancelOutstandingJobs(ctx, experimentID); err != nil {
			logger.Log.Error("Failed to clean up jobs of cancelled experiment", zap.Error(err), zap.String("id", experimentID))
		}
		return
	}

	if s.Producer != nil {
		s.Producer.Publish(ctx, "event-stream", experimentID, models.ExperimentEvent{
			ExperimentID: experimentID,
			EventType:    models.EventTypeIngestionStarted,
			TotalJobs:    len(jobs),
		})
	}

	msgs := make([]keyedMessage, 0, len(jobs))
	for _, j := range jobs {
		msgs = append(msgs, keyedMessage{key: j.ID, value: models.IngestionJobMessage{
			JobID:        j.ID,
			ExperimentID: experimentID,
			Symbol:       j.Symbol,
			StartDate:    reqStart,
			EndDate:      reqEnd,
		}})
	}
	s.publishAll(ctx, "ingestion-jobs", msgs)

	logger.Log.Info("Ingestion jobs created", zap.String("id", experimentID), zap.Int("jobs", len(jobs)), zap.Int("already_covered", len(symbols)-len(missing)))
}

// GenerateBacktestJobs expects the experiment to be in CREATING_BACKTEST_JOBS. It inserts all jobs,
// moves the experiment to RUNNING, then publishes them. Every transition is conditional, so a
// cancel at any point wins and leaves no live jobs behind.
func (s *ExperimentService) GenerateBacktestJobs(ctx context.Context, experimentID string) error {
	configJSON, err := s.Repo.GetExperimentConfig(ctx, experimentID)
	if err != nil {
		return err
	}

	var config models.CreateExperimentRequest
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return err
	}

	var msgs []keyedMessage
	inserted := 0

	for _, sym := range config.Symbols {
		for _, tf := range config.Timeframes {
			for _, strat := range config.Strategies {
				combinations := generateCombinations(strat.Parameters)
				if len(combinations) == 0 {
					combinations = append(combinations, make(map[string]interface{}))
				}

				for _, params := range combinations {
					paramsJSON, _ := json.Marshal(params)
					jobID, err := s.Repo.InsertBacktestJob(ctx, experimentID, sym, strat.Name, paramsJSON, tf.Start, tf.End)
					if err != nil {
						logger.Log.Error("Failed to insert backtest job", zap.Error(err))
						continue
					}
					if jobID == "" {
						// Job already exists (idempotency conflict) - skip publishing
						continue
					}
					inserted++
					msgs = append(msgs, keyedMessage{key: jobID, value: models.BacktestJobMessage{
						JobID:        jobID,
						ExperimentID: experimentID,
						Symbol:       sym,
						StrategyName: strat.Name,
						Parameters:   params,
						StartDate:    tf.Start,
						EndDate:      tf.End,
					}})

					// Notice a cancel quickly instead of inserting thousands of orphan jobs.
					if inserted%200 == 0 {
						if st, err := s.Repo.GetExperimentStatus(ctx, experimentID); err == nil && st == string(models.ExperimentStatusCancelled) {
							s.Repo.CancelOutstandingJobs(ctx, experimentID)
							return nil
						}
					}
				}
			}
		}
	}

	if inserted == 0 {
		s.Transition(ctx, experimentID, models.ExperimentStatusCreatingBacktestJobs, models.ExperimentStatusCompleted)
		return nil
	}

	// Go to RUNNING before publishing so a job finishing instantly cannot race the status change.
	if !s.Transition(ctx, experimentID, models.ExperimentStatusCreatingBacktestJobs, models.ExperimentStatusRunning) {
		s.Repo.CancelOutstandingJobs(ctx, experimentID)
		return nil
	}

	if s.Producer != nil {
		s.Producer.Publish(ctx, "event-stream", experimentID, models.ExperimentEvent{
			ExperimentID: experimentID,
			EventType:    models.EventTypeBacktestingStarted,
			TotalJobs:    inserted,
		})
	}
	s.publishAll(ctx, "backtesting-jobs", msgs)
	return nil
}

func generateCombinations(params map[string][]interface{}) []map[string]interface{} {
	var keys []string
	var values [][]interface{}
	for k, v := range params {
		keys = append(keys, k)
		values = append(values, v)
	}

	var results []map[string]interface{}
	var helper func(int, map[string]interface{})
	helper = func(depth int, current map[string]interface{}) {
		if depth == len(keys) {
			clone := make(map[string]interface{})
			for k, v := range current {
				clone[k] = v
			}
			results = append(results, clone)
			return
		}
		
		for _, v := range values[depth] {
			current[keys[depth]] = v
			helper(depth+1, current)
		}
	}
	
	if len(keys) > 0 {
		helper(0, make(map[string]interface{}))
	}
	return results
}

func (s *ExperimentService) CancelExperiment(ctx context.Context, experimentID string) error {
	logger.Log.Info("Cancelling experiment", zap.String("id", experimentID))

	// Use our own deadline: a caller that disconnects must not abort a half-finished cancel.
	opCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	changed, err := s.Repo.CancelExperimentJobs(opCtx, experimentID)
	if err != nil {
		logger.Log.Error("Failed to cancel experiment jobs", zap.Error(err), zap.String("id", experimentID))
		return err
	}

	if changed && s.Producer != nil {
		pubCtx, pubCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer pubCancel()
		s.Producer.Publish(pubCtx, "event-stream", experimentID, models.ExperimentEvent{
			ExperimentID: experimentID,
			EventType:    models.EventTypeExperimentCancelled,
		})
		s.publishStatusChanged(pubCtx, experimentID, models.ExperimentStatusCancelled)
	}

	return nil
}
