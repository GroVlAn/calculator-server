package service

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type calculator interface {
	Add(a, b int64) int64
	Sub(a, b int64) int64
}

type prometheus interface {
	CMetric(fn func(a, b int64) int64, a, b int64) int64
	RustMetric(fn func(a, b int64) int64, a, b int64) int64
}

type Service struct {
	l        zerolog.Logger
	pr       prometheus
	sumValue int64
	subValue int64
	sumMt    sync.Mutex
	subMt    sync.Mutex
	calc     calculator
}

func New(l zerolog.Logger, pr prometheus, calc calculator) *Service {
	return &Service{
		l:    l,
		pr:   pr,
		calc: calc,
	}
}

func (s *Service) Calculate(value int64) {
	var wg sync.WaitGroup

	wg.Go(
		func() {
			s.sumMt.Lock()
			defer s.sumMt.Unlock()

			s.sumValue = s.pr.CMetric(
				s.calc.Add,
				s.sumValue,
				value,
			)
		},
	)
	wg.Go(
		func() {
			s.subMt.Lock()
			defer s.subMt.Unlock()

			s.subValue = s.pr.RustMetric(
				s.calc.Sub,
				s.subValue,
				value,
			)
		},
	)

	wg.Wait()
}

func (s *Service) PeriodicPrinter(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.LogTotals("periodic")
		}
	}
}

func (s *Service) LogTotals(label string) {
	s.sumMt.Lock()
	s.subMt.Lock()

	sum := s.sumValue
	sub := s.subValue

	s.sumMt.Unlock()
	s.subMt.Unlock()

	s.l.Info().Msgf("[%s] sum=%d sub=%d", label, sum, sub)
}
