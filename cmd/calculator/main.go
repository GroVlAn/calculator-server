package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GroVlAn/calculator-server/internal/config"
	"github.com/GroVlAn/calculator-server/internal/handler"
	"github.com/GroVlAn/calculator-server/internal/infrastructure/calculator"
	prometheus_metrics "github.com/GroVlAn/calculator-server/internal/infrastructure/metrics"
	"github.com/GroVlAn/calculator-server/internal/server"
	"github.com/GroVlAn/calculator-server/internal/service"
	"github.com/rs/zerolog"
)

const (
	localConfigPath = "configs/config.yml"

	labelFinalCalculate = "final"
)

func main() {
	timeStart := time.Now()

	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configPath := flag.String("config", localConfigPath, "Path to the configuration file")
	flag.Parse()

	cfg, err := config.New(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load configuration")
	}

	calc := calculator.New()

	pr := prometheus_metrics.New()

	s := service.New(l, pr, calc)

	h := handler.New(
		l,
		s,
		pr,
		handler.Deps{
			DefaultTimeout: cfg.HTTP.DefaultTimeout,
			BasePath:       cfg.HTTP.BaseHTTPPath,
		},
	)

	srv := server.New(
		h.Handler(),
		server.Settings{
			Port:              cfg.HTTP.Port,
			MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
			ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
		},
	)

	errCh := make(chan error, 1)

	go func() {
		l.Info().Msgf("starting http server on port: %s", cfg.HTTP.Port)

		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		s.PeriodicPrinter(ctx, cfg.Settings.LoggingInterval)
	}()

	l.Info().
		Dur("startup_time", time.Since(timeStart)).
		Str("http_port", cfg.HTTP.Port).
		Msg("server started")

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		s.LogTotals(labelFinalCalculate)
		if err := srv.Shutdown(shutdownCtx); err != nil {
			l.Error().Err(err).Msg("failed to shutdown server")
		} else {
			l.Info().Msg("server shutdown gracefully")
		}
	}

	select {
	case <-ctx.Done():
		shutdown()
	case err := <-errCh:
		if err != nil {
			l.Error().Err(err).Msg("server exited with error")

			shutdown()
		}
	}
}
