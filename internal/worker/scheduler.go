package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// ScheduledJob is a periodic job definition.
type ScheduledJob struct {
	Name     string
	Interval time.Duration
	Job      Job
}

// Scheduler runs jobs on a fixed interval.
type Scheduler struct {
	jobs   []ScheduledJob
	worker *Worker
	log    logger.Logger
}

// NewScheduler creates a Scheduler backed by the given Worker.
func NewScheduler(w *Worker, log logger.Logger) *Scheduler {
	return &Scheduler{worker: w, log: log}
}

// Register adds a job to the scheduler.
func (s *Scheduler) Register(name string, interval time.Duration, job Job) {
	s.jobs = append(s.jobs, ScheduledJob{Name: name, Interval: interval, Job: job})
}

// Run starts all scheduled jobs and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	for _, sj := range s.jobs {
		go s.runJob(ctx, sj)
	}
	<-ctx.Done()
}

func (s *Scheduler) runJob(ctx context.Context, sj ScheduledJob) {
	ticker := time.NewTicker(sj.Interval)
	defer ticker.Stop()

	s.log.Info("scheduler: registered job", zap.String("job", sj.Name), zap.Duration("interval", sj.Interval))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.log.Info("scheduler: triggering job", zap.String("job", sj.Name))
			s.worker.Enqueue(sj.Job)
		}
	}
}
