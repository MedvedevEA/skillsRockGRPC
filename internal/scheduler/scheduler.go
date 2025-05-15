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
		s.lg.Info("SCHEDULER: task 'RemoveRefreshTokens' start", slog.Any("interval", s.cfg.TimeoutRemoveRefreshTokens))
		for {
			select {
			case <-s.chStop:
				s.lg.Info("SCHEDULER: task 'RemoveRefreshTokens' stop")
				s.wg.Done()
				return
			case <-time.After(s.cfg.TimeoutRemoveRefreshTokens):
				count, err := fn(time.Now())
				if err != nil {
					s.lg.Error("SCHEDULER: task 'RemoveRefreshTokens' exec error", slog.Any("error", err))
				}
				s.lg.Info("SCHEDULER: task 'RemoveRefreshTokens' exec success", slog.Any("rows affected", count))
			}
		}
	}()
}
func (s *Scheduler) Stop() {
	close(s.chStop)
	s.wg.Wait()
}
