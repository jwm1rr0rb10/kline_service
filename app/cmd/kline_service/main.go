package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jwm1rr0rb10/go-logging"
	"golang.org/x/sync/errgroup"

	"github.com/jwm1rr0rb10/kline_service/app/internal/app"
	"github.com/jwm1rr0rb10/kline_service/app/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("failed to initialize the application", logging.ErrAttr(err))
		os.Exit(1)
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return newApp.Run(gCtx)
	})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		<-sigChan
		logging.L(gCtx).Info("received a termination signal and are initiating a graceful shutdown...")

		cancel()

		time.AfterFunc(config.ShutdownTimeout, func() {
			logging.L(gCtx).Error("graceful shutdown timeout exceeded, forced exit")
			os.Exit(130)
		})
	}()

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logging.L(ctx).Error("the application terminated with an error", logging.ErrAttr(err))
		os.Exit(1)
	}

	logging.L(ctx).Info("The application has been stopped successfully.")
}
