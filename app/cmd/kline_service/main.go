package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jwm1rr0rb10/go-logging"
	"github.com/jwm1rr0rb10/kline_service/app/internal/config"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("не удалось инициализировать приложение", logging.ErrAttr(err))
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
		logging.L(gCtx).Info("получен сигнал завершения, начинаем graceful shutdown...")

		cancel()

		time.AfterFunc(config.ShutdownTimeout, func() {
			logging.L(gCtx).Error("превышен таймаут graceful shutdown, принудительный выход")
			os.Exit(130)
		})
	}()

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logging.L(ctx).Error("приложение завершилось с ошибкой", logging.ErrAttr(err))
		os.Exit(1)
	}

	logging.L(ctx).Info("приложение успешно остановлено")
}
