package cpa

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrRuntimeAlreadyRunning = errors.New("cpa runtime already running")

type runtimeConfigStore interface {
	Get(ctx context.Context) (*Config, error)
}

type runtimeSyncer interface {
	RunOnce(ctx context.Context, cfg Config, triggerType string, triggeredBy int64) (SyncResult, error)
}

type runtimeTicker interface {
	C() <-chan time.Time
	Stop()
}

type Runtime struct {
	cfgStore runtimeConfigStore
	syncer   runtimeSyncer

	mu            sync.Mutex
	status        Status
	started       bool
	loopCancel    context.CancelFunc
	loopDone      chan struct{}
	tickerFactory func(time.Duration) runtimeTicker
}

func NewRuntime(cfgStore runtimeConfigStore, syncer runtimeSyncer) *Runtime {
	cfg := normalizeRuntimeConfig(DefaultConfig())
	return &Runtime{
		cfgStore: cfgStore,
		syncer:   syncer,
		status: Status{
			Enabled:         cfg.Enabled,
			IntervalSeconds: cfg.IntervalSeconds,
			SourceDir:       cfg.SourceDir,
			GroupID:         cfg.GroupID,
		},
		tickerFactory: func(interval time.Duration) runtimeTicker {
			return &stdRuntimeTicker{ticker: time.NewTicker(interval)}
		},
	}
}

func (r *Runtime) Start(ctx context.Context) error {
	cfg, err := r.loadConfig(ctx)
	if err != nil {
		return err
	}

	r.mu.Lock()
	if r.started {
		r.applyConfigLocked(cfg)
		r.mu.Unlock()
		return nil
	}
	r.applyConfigLocked(cfg)
	r.started = true
	if !cfg.Enabled {
		r.mu.Unlock()
		return nil
	}

	loopCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	r.loopCancel = cancel
	r.loopDone = done
	ticker := r.tickerFactory(time.Duration(cfg.IntervalSeconds) * time.Second)
	r.mu.Unlock()

	go r.runLoop(loopCtx, done, ticker)
	return nil
}

func (r *Runtime) Stop() {
	r.mu.Lock()
	cancel := r.loopCancel
	done := r.loopDone
	r.started = false
	r.loopCancel = nil
	r.loopDone = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (r *Runtime) RunNow(ctx context.Context, triggeredBy int64) (SyncResult, error) {
	cfg, err := r.loadConfig(ctx)
	if err != nil {
		return SyncResult{}, err
	}
	return r.run(ctx, cfg, TriggerTypeManual, triggeredBy, nil)
}

func (r *Runtime) Status() Status {
	r.mu.Lock()
	defer r.mu.Unlock()

	status := r.status
	status.LastTickAt = cloneRuntimeTimePtr(r.status.LastTickAt)
	status.LastSuccessAt = cloneRuntimeTimePtr(r.status.LastSuccessAt)
	return status
}

func (r *Runtime) runLoop(ctx context.Context, done chan struct{}, ticker runtimeTicker) {
	defer close(done)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tickAt, ok := <-ticker.C():
			if !ok {
				return
			}
			cfg, err := r.loadConfig(ctx)
			if err != nil {
				r.recordRuntimeError(err)
				continue
			}
			if !cfg.Enabled {
				r.mu.Lock()
				r.applyConfigLocked(cfg)
				r.mu.Unlock()
				continue
			}
			if _, err := r.run(ctx, cfg, TriggerTypeAuto, 0, &tickAt); err != nil && !errors.Is(err, ErrRuntimeAlreadyRunning) {
				r.recordRuntimeError(err)
			}
		}
	}
}

func (r *Runtime) run(ctx context.Context, cfg Config, triggerType string, triggeredBy int64, tickAt *time.Time) (SyncResult, error) {
	if !r.beginRun(cfg, tickAt) {
		return SyncResult{}, ErrRuntimeAlreadyRunning
	}

	result, err := r.syncer.RunOnce(ctx, cfg, triggerType, triggeredBy)
	r.finishRun(result, err)
	return result, err
}

func (r *Runtime) beginRun(cfg Config, tickAt *time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.applyConfigLocked(cfg)
	if r.status.Running {
		if tickAt != nil {
			r.status.LastTickAt = cloneRuntimeTimePtr(tickAt)
		}
		return false
	}
	r.status.Running = true
	if tickAt != nil {
		r.status.LastTickAt = cloneRuntimeTimePtr(tickAt)
	}
	r.status.LastError = ""
	return true
}

func (r *Runtime) finishRun(result SyncResult, runErr error) {
	finishedAt := time.Now().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.status.Running = false
	if runErr != nil {
		r.status.LastError = runErr.Error()
		return
	}

	r.status.LastError = ""
	r.status.LastSuccessAt = &finishedAt
	if result.JobID > 0 {
		r.status.CurrentJobID = int(result.JobID)
	}
}

func (r *Runtime) recordRuntimeError(err error) {
	if err == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.status.LastError = err.Error()
}

func (r *Runtime) loadConfig(ctx context.Context) (Config, error) {
	cfg, err := r.cfgStore.Get(ctx)
	if err != nil {
		return Config{}, err
	}
	if cfg == nil {
		defaultCfg := normalizeRuntimeConfig(DefaultConfig())
		return defaultCfg, nil
	}
	return normalizeRuntimeConfig(*cfg), nil
}

func (r *Runtime) applyConfigLocked(cfg Config) {
	cfg = normalizeRuntimeConfig(cfg)
	r.status.Enabled = cfg.Enabled
	r.status.IntervalSeconds = cfg.IntervalSeconds
	r.status.SourceDir = cfg.SourceDir
	r.status.GroupID = cfg.GroupID
}

func normalizeRuntimeConfig(cfg Config) Config {
	defaultCfg := DefaultConfig()
	if cfg.SourceDir == "" {
		cfg.SourceDir = defaultCfg.SourceDir
	}
	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = defaultCfg.IntervalSeconds
	}
	return cfg
}

func cloneRuntimeTimePtr(ts *time.Time) *time.Time {
	if ts == nil {
		return nil
	}
	value := ts.UTC()
	return &value
}

type stdRuntimeTicker struct {
	ticker *time.Ticker
}

func (t *stdRuntimeTicker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *stdRuntimeTicker) Stop() {
	t.ticker.Stop()
}
