package scheduler

import (
	"log/slog"
	"skillsRockGRPC/internal/config"
	"sync"
	"time"
)

type Scheduler struct {
	lg     *slog.Logger
	cfg    *config.Scheduler
	wg     *sync.WaitGroup
	chStop chan struct{}
}

func New(lg *slog.Logger, cfg *config.Scheduler) *Scheduler {
	return &Scheduler{
		lg:     lg,
		cfg:    cfg,
		wg:     &sync.WaitGroup{},
		chStop: make(chan struct{}, 1),
	}
}

func (s *Scheduler) RemoveRefreshTokens(fn func(time.Time) (int64, error)) {
	s.wg.Add(1)
	go func() {
		const op = "scheduler.RemoveRefreshTokens"
		s.lg.Info("Task scheduler 'RemoveRefreshTokens' is strarted", slog.String("op", op), slog.Any("interval", s.cfg.TimeoutRemoveRefreshTokens))
		for {
			select {
			case <-s.chStop:
				s.lg.Info("Task scheduler 'RemoveRefreshTokens' is stopped", slog.String("op", op))
				s.wg.Done()
				return
			case <-time.After(s.cfg.TimeoutRemoveRefreshTokens):
				count, err := fn(time.Now())
				if err != nil {
					s.lg.Error("Task scheduler 'RemoveRefreshTokens' error", slog.String("op", op), slog.Any("error", err))
				}
				s.lg.Info("Task 'RemoveRefreshTokens' has been successfully completed", slog.String("op", op), slog.Any("rows affected", count))
			}
		}
	}()
}
func (s *Scheduler) Stop() {
	close(s.chStop)
	s.wg.Wait()
}
