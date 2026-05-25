package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jwm1rr0rb10/go-errors"
	"github.com/jwm1rr0rb10/go-pprof"
	"github.com/jwm1rr0rb10/libraries/backend/golang/logging"
	"github.com/jwm1rr0rb10/libraries/backend/golang/metrics"
	"github.com/jwm1rr0rb10/libraries/backend/golang/postgresql"
	"github.com/jwm1rr0rb10/libraries/backend/golang/tracing"

	"github.com/jwm1rr0rb10/kline_service/app/internal/config"
)

func (a *App) setupHTTPServer(ctx context.Context) error {
	logging.L(ctx).Info(
		"HTTP server initializing",
		logging.StringAttr("host", a.cfg.HTTP.Host),
		logging.IntAttr("port", a.cfg.HTTP.Port),
	)

	a.httpServer = &http.Server{
		Addr:        fmt.Sprintf("%s:%d", a.cfg.HTTP.Host, a.cfg.HTTP.Port),
		Handler:     a.httpRouter,
		ReadTimeout: a.cfg.HTTP.ReadHeaderTimeout,
	}

	return nil
}

func (a *App) setupDebug(ctx context.Context) error {
	if !a.cfg.App.IsDevelopment {
		logging.L(ctx).Info("debug service not started, because app is not in development mode")
		return nil
	}

	debugServer := pprof.NewServer(pprof.NewConfig(
		a.cfg.Profiler.Host,
		a.cfg.Profiler.Port,
		a.cfg.Profiler.ReadHeaderTimeout,
	))

	go func() {
		logging.L(ctx).Info(
			"pprof debug server started",
			logging.StringAttr("host", a.cfg.Profiler.Host),
			logging.IntAttr("port", a.cfg.Profiler.Port),
		)

		err := debugServer.Run(ctx)
		if err != nil {
			logging.L(ctx).Error("error listen debug server", logging.ErrAttr(err))
		}
	}()

	return nil
}

func (a *App) initPostgresClient(ctx context.Context) (*psql.Client, error) {
	logging.WithAttrs(
		ctx,
		logging.StringAttr("host", a.cfg.Postgres.Host),
		logging.IntAttr("port", a.cfg.Postgres.Port),
		logging.StringAttr("user", a.cfg.Postgres.User),
		logging.StringAttr("db", a.cfg.Postgres.Database),
		logging.StringAttr("password", "<REMOVED>"),
		logging.IntAttr("max-attempts", a.cfg.Postgres.MaxAttempt),
		logging.DurationAttr("max_delay", a.cfg.Postgres.MaxDelay),
	).Info("PostgreSQL initializing")

	pgDsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		a.cfg.Postgres.User,
		a.cfg.Postgres.Password,
		a.cfg.Postgres.Host,
		a.cfg.Postgres.Port,
		a.cfg.Postgres.Database,
		sslMode(a.cfg.Postgres.RequireSSL),
	)

	if a.cfg.Postgres.RequireSSL {
		pgDsn += "?sslmode=require"
	}

	postgresConfig, err := psql.NewConfig(
		pgDsn,
		psql.WithBinaryExecMode(a.cfg.Postgres.RequireSSL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "psql.NewConfig")
	}

	pgClient, err := psql.NewClient(ctx, postgresConfig)
	if err != nil {
		return nil, errors.Wrap(err, "psql.NewClient")
	}

	a.closer.AddNoErr(pgClient)

	return pgClient, nil
}

func sslMode(requireSSL bool) string {
	if requireSSL {
		return "require"
	}
	return "disable"
}

func (a *App) initMetricsServer(ctx context.Context, cfg config.MetricsConfig) (*metrics.Server, error) {
	if !cfg.Enabled {
		logging.L(ctx).Info("Metrics disabled")

		return nil, nil
	}

	logging.L(ctx).Info(
		"Metrics initializing",
		logging.StringAttr("host", cfg.Host),
		logging.IntAttr("port", cfg.Port),
	)

	metricsHTTTPServer, err := metrics.NewServer(metrics.NewConfig(
		metrics.WithHost(cfg.Host),
		metrics.WithPort(cfg.Port),
		metrics.WithReadTimeout(cfg.ReadTimeout),
		metrics.WithWriteTimeout(cfg.WriteTimeout),
		metrics.WithReadHeaderTimeout(cfg.ReadHeaderTimeout),
	))
	if err != nil {
		return nil, errors.Wrap(err, "metrics.NewServer")
	}

	a.closer.Add(metricsHTTTPServer)

	return metricsHTTTPServer, nil
}

func initTraceServer(ctx context.Context, cfg *config.Config) error {
	if !cfg.Tracing.Enabled {
		logging.L(ctx).Info("tracing is disabled")

		return nil
	}

	_, err := tracing.New(
		tracing.WithHost(cfg.Tracing.Host),
		tracing.WithPort(strconv.Itoa(cfg.Tracing.Port)),
		tracing.WithServiceID(cfg.App.ID),
		tracing.WithServiceName(cfg.App.Name),
		tracing.WithServiceVersion(cfg.App.Version),
	)
	if err != nil {
		return errors.Wrap(err, "can't initialize tracing")
	}

	return nil
}
