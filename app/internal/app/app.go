package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jwm1rr0rb10/go-errors"
	"github.com/jwm1rr0rb10/go-logging"
	"github.com/jwm1rr0rb10/go-metrics"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/clock"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/closer"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/safe/errorgroup"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/uuid/google_uuid"

	"github.com/jwm1rr0rb10/kline_service/app/internal/config"
	"github.com/jwm1rr0rb10/kline_service/app/internal/dal/postgres"
	serviceSpotKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/service"
	storageSpotKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/storage/postgres"
	"github.com/jwm1rr0rb10/kline_service/app/internal/policy"
	policySpotKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

const cfgPath = "/Users/jwm1rr0rb/Desktop/kline_service/configs/config.local.yaml"

type Runner interface {
	Run(context.Context) error
}

type App struct {
	cfg *config.Config

	httpRouter         *chi.Mux
	httpServer         *http.Server
	metricsHTTTPServer *metrics.Server

	policySpot *policySpotKlineOKX.Policy
	// adapterSpotKline *wsKline.SpotWebSocket

	runners []Runner
	recover errorgroup.RecoverFunc
	closer  *closer.LIFOCloser
}

func (a *App) AddRunner(runner Runner) {
	a.runners = append(a.runners, runner)
}

func NewApp(ctx context.Context) (*App, error) {
	app := App{
		closer: closer.NewLIFOCloser(),
	}

	cfg := config.MustLoadConfig(cfgPath)
	app.cfg = cfg

	if cfg.App.LogLevel == "" {
		cfg.App.LogLevel = "info"
	}
	logger := logging.NewLogger(
		logging.WithLevel(cfg.App.LogLevel),
		logging.WithIsJSON(cfg.App.IsLogJSON),
	)
	ctx = logging.ContextWithLogger(ctx, logger)

	logging.L(ctx).Info("config loaded", "config", cfg)

	// Init Trace
	if err := initTraceServer(ctx, cfg); err != nil {
		return nil, errors.Wrap(err, "initTraceServer")
	}

	// Init Postgres
	postgresClient, err := app.initPostgresClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't create postgres Client")
	}

	// Init Policy + Storage
	storageSpot := storageSpotKlineOKX.NewStorage(postgresClient)
	serviceSpot := serviceSpotKlineOKX.NewService(storageSpot)

	basePolicy := policy.NewBasePolicy(
		google_uuid.NewGoogleUUIDGenerator(),
		clock.New(),
	)

	app.policySpot = policySpotKlineOKX.NewPolicy(basePolicy, serviceSpot)

	app.httpRouter = chi.NewRouter()
	if err := app.setupHTTPServer(ctx); err != nil {
		return nil, errors.Wrap(err, "setupHTTPServer")
	}

	return &app, nil
}

func (a *App) Run(ctx context.Context) error {
	// Migrations
	if err := postgres.RunMigrations(ctx, &a.cfg.Postgres); err != nil {
		return errors.Wrap(err, "migrations failed")
	}

	g, ctx := errorgroup.WithContext(ctx, errorgroup.WithRecover(a.recover))

	// Graceful shutdown
	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		if a.httpServer != nil {
			return a.httpServer.Shutdown(context.Background())
		}
		return nil
	})

	// HTTP Server
	g.Go(func(ctx context.Context) error {
		logging.L(ctx).Info("HTTP server starting", "addr", a.httpServer.Addr)
		return a.httpServer.ListenAndServe()
	})

	for _, r := range a.runners {
		runner := r
		g.Go(func(ctx context.Context) error { return runner.Run(ctx) })
	}

	logging.L(ctx).Info("application started")
	return g.Wait()
}
