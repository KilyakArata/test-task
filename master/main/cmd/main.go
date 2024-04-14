package main

import (
	"avito-testovoe/config"
	"avito-testovoe/handler"
	c "avito-testovoe/internal/cache"
	"avito-testovoe/internal/logger"
	"avito-testovoe/internal/storage"
	"context"
	"golang.org/x/sync/errgroup"
	"log/slog"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx))
}

func run(ctx context.Context) int {
	log := logger.SettUpLogger()

	log.Info("Логгер подключен")

	log.Info("Старт приложения avito-testovoe")

	cfg := config.MustLoad()

	log.Info("Конфиг прочитан")

	cache := c.New(cfg.DefaultExpiration, cfg.CleanupInterval)

	log.Info("Кэш контейнет создан")

	storage, err := sqlite.New(cfg.StoragePath, log, ctx)
	if err != nil {
		log.Error("Ошибка подключения к базе данных")
		os.Exit(1)
	}

	log.Info("База данных подключена")

	srv := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      handler.NewServer(log, storage, cache, ctx),
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("Запускаем сервер:", slog.String("server", cfg.Address))

		return srv.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		log.Info("Остановка сервера")

		shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		return srv.Shutdown(shutCtx)
	})

	err = g.Wait()
	if err != nil && err != http.ErrServerClosed {
		log.Error("Сервер остановился с ошибкой: ", err)
		return 1
	}

	log.Info("Сервер остановлен")

	return 0
}
