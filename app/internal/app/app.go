package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jwm1rr0rb10/go-errors"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/clock"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/closer"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/safe/errorgroup"
	"github.com/jwm1rr0rb10/libraries/backend/golang/core/uuid/google_uuid"
	"github.com/jwm1rr0rb10/libraries/backend/golang/logging"
	"github.com/jwm1rr0rb10/libraries/backend/golang/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jwm1rr0rb10/kline_service/app/internal/config"
	"github.com/jwm1rr0rb10/kline_service/app/internal/dal/postgres"
	serviceSpotKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/service"
	storageSpotKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/storage/postgres"
	"github.com/jwm1rr0rb10/kline_service/app/internal/policy"
	policySpotKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"

	// gRPC
	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service/v1"
	klineCtrl "github.com/jwm1rr0rb10/kline_service/app/internal/controller/grpc/v1/kline"

	// WebSocket
	wsKline "github.com/jwm1rr0rb10/kline_service/app/internal/adapter/websocket/v1/kline"
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
	grpcServer *grpc.Server
	klineCtrl  *klineCtrl.Controller

	adapterSpotKline *wsKline.OkxSpotWebSocket

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

	if err := initTraceServer(ctx, cfg); err != nil {
		return nil, errors.Wrap(err, "initTraceServer")
	}

	postgresClient, err := app.initPostgresClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can't create postgres Client")
	}

	// === Policy + Storage ===
	storageSpot := storageSpotKlineOKX.NewStorage(postgresClient)
	serviceSpot := serviceSpotKlineOKX.NewService(storageSpot)

	basePolicy := policy.NewBasePolicy(
		google_uuid.NewGoogleUUIDGenerator(),
		clock.New(),
	)
	app.policySpot = policySpotKlineOKX.NewPolicy(basePolicy, serviceSpot)

	// === gRPC Controller ===
	app.klineCtrl = klineCtrl.NewController(app.policySpot)

	// === WebSocket Adapter ===
	symbols := []string{"BTC-USDT", "ETH-USDT", "SOL-USDT"} // TODO: вынести в конфиг
	intervals := []string{"1m", "5m", "15m", "1H"}
	app.adapterSpotKline = wsKline.NewOkxSpotWebSocket(
		app.policySpot,
		symbols,
		intervals,
		app.cfg.WebSocket.ReconnectTimes,
		app.cfg.WebSocket.ReconnectDelay,
	)

	// HTTP
	app.httpRouter = chi.NewRouter()
	if err := app.setupHTTPServer(ctx); err != nil {
		return nil, errors.Wrap(err, "setupHTTPServer")
	}

	// gRPC
	if err := app.setupGRPCServer(ctx); err != nil {
		return nil, errors.Wrap(err, "setupGRPCServer")
	}

	return &app, nil
}

func (a *App) setupGRPCServer(ctx context.Context) error {
	logging.L(ctx).Info("gRPC server initializing",
		logging.StringAttr("host", a.cfg.GRPC.Host),
		logging.IntAttr("port", a.cfg.GRPC.Port),
	)

	a.grpcServer = grpc.NewServer()
	gRPCKlineService.RegisterKlineServiceServer(a.grpcServer, a.klineCtrl)

	if a.cfg.App.IsDevelopment {
		reflection.Register(a.grpcServer)
		logging.L(ctx).Info("gRPC reflection enabled (development)")
	}

	return nil
}

func (a *App) Run(ctx context.Context) error {
	if err := postgres.RunMigrations(ctx, &a.cfg.Postgres); err != nil {
		return errors.Wrap(err, "migrations failed")
	}

	g, ctx := errorgroup.WithContext(ctx, errorgroup.WithRecover(a.recover))

	// Graceful shutdown
	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		if a.httpServer != nil {
			_ = a.httpServer.Shutdown(context.Background())
		}
		if a.grpcServer != nil {
			a.grpcServer.GracefulStop()
		}
		if a.adapterSpotKline != nil {
			a.adapterSpotKline.Close()
		}
		return nil
	})

	// HTTP
	g.Go(func(ctx context.Context) error {
		logging.L(ctx).Info("HTTP server starting", "addr", a.httpServer.Addr)
		return a.httpServer.ListenAndServe()
	})

	// gRPC
	g.Go(func(ctx context.Context) error {
		addr := fmt.Sprintf("%s:%d", a.cfg.GRPC.Host, a.cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return errors.Wrap(err, "failed to listen gRPC")
		}
		logging.L(ctx).Info("gRPC server starting", "addr", addr)
		return a.grpcServer.Serve(lis)
	})

	// WebSocket
	g.Go(func(ctx context.Context) error {
		logging.L(ctx).Info("WebSocket adapter starting...")
		if err := a.adapterSpotKline.Connect(ctx); err != nil {
			return errors.Wrap(err, "websocket connect")
		}
		a.adapterSpotKline.Listen(ctx) // blocking
		return nil
	})

	for _, r := range a.runners {
		runner := r
		g.Go(func(ctx context.Context) error { return runner.Run(ctx) })
	}

	logging.L(ctx).Info("application started")
	return g.Wait()
}
