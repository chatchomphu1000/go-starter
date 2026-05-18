// Package worker provides a simple goroutine-based background job system.
package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// Job is a unit of work the worker can execute.
type Job interface {
	Name() string
	Execute(ctx context.Context) error
}

// Worker manages a pool of goroutines for executing background jobs.
type Worker struct {
	jobs    chan Job
	wg      sync.WaitGroup
	log     logger.Logger
	workers int
}

// New creates a Worker with the given concurrency level.
func New(concurrency int, queueSize int, log logger.Logger) *Worker {
	if concurrency < 1 {
		concurrency = 2
	}
	return &Worker{
		jobs:    make(chan Job, queueSize),
		log:     log,
		workers: concurrency,
	}
}

// Start launches the worker goroutines and blocks until ctx is cancelled.
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-w.jobs:
					if !ok {
						return
					}
					w.execute(ctx, job)
				}
			}
		}()
	}
	<-ctx.Done()
	close(w.jobs)
	w.wg.Wait()
}

// Enqueue submits a job for async execution (non-blocking; drops if queue full).
func (w *Worker) Enqueue(job Job) {
	select {
	case w.jobs <- job:
	default:
		w.log.Warn("worker queue full, dropping job", zap.String("job", job.Name()))
	}
}

func (w *Worker) execute(ctx context.Context, job Job) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			w.log.Error("job panicked", zap.String("job", job.Name()), zap.Any("panic", r))
		}
	}()

	if err := job.Execute(ctx); err != nil {
		w.log.Error("job failed",
			zap.String("job", job.Name()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return
	}
	w.log.Info("job completed",
		zap.String("job", job.Name()),
		zap.Duration("duration", time.Since(start)),
	)
}
