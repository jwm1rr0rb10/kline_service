package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 45 * time.Second

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация приложения
	newApp, err := app.NewApp(ctx)
	if err != nil {
		logging.L(ctx).Error("не удалось инициализировать приложение", logging.ErrAttr(err))
		os.Exit(1)
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Запускаем основную логику приложения
	g.Go(func() error {
		return newApp.Run(gCtx)
	})

	// Обработка сигналов завершения
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		<-sigChan
		logging.L(gCtx).Info("получен сигнал завершения, начинаем graceful shutdown...")

		cancel()

		// Принудительный выход по таймауту
		time.AfterFunc(shutdownTimeout, func() {
			logging.L(gCtx).Error("превышен таймаут graceful shutdown, принудительный выход")
			os.Exit(130)
		})
	}()

	// Ожидаем завершения всех горутин
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logging.L(ctx).Error("приложение завершилось с ошибкой", logging.ErrAttr(err))
		os.Exit(1)
	}

	logging.L(ctx).Info("приложение успешно остановлено")
}
