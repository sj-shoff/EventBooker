package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"event-booker/internal/config"
	"event-booker/internal/http-server/handler/booking"
	"event-booker/internal/http-server/handler/event"
	"event-booker/internal/http-server/handler/user"
	"event-booker/internal/http-server/router"
	"event-booker/internal/notification/composite"
	"event-booker/internal/notification/email"
	"event-booker/internal/notification/telegram"
	booking_repo "event-booker/internal/repository/booking/postgres"
	event_repo "event-booker/internal/repository/event/postgres"
	user_repo "event-booker/internal/repository/user/postgres"
	"event-booker/internal/scheduler"
	booking_uc "event-booker/internal/usecase/booking"
	event_uc "event-booker/internal/usecase/event"
	user_uc "event-booker/internal/usecase/user"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	cfg       *config.Config
	server    *http.Server
	logger    *zlog.Zerolog
	db        *dbpg.DB
	scheduler *scheduler.Scheduler
}

func NewApp(cfg *config.Config, logger *zlog.Zerolog) (*App, error) {
	retries := cfg.DefaultRetryStrategy()

	dbOpts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}

	db, err := dbpg.New(cfg.DBDSN(), []string{}, dbOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	emailNotifier := email.NewNotifier(cfg)
	telegramNotifier := telegram.NewNotifier(cfg.TelegramConfig.BotToken)
	compositeNotifier := composite.NewCompositeNotifier(emailNotifier, telegramNotifier)

	bookingRepo := booking_repo.NewBookingRepository(db, retries)
	eventRepo := event_repo.NewEventRepository(db, retries)
	userRepo := user_repo.NewUserRepository(db, retries)

	bookingUsecase := booking_uc.NewBookingUsecase(db, bookingRepo, eventRepo, userRepo, compositeNotifier, cfg, logger)
	eventUsecase := event_uc.NewEventUsecase(db, eventRepo, bookingRepo, userRepo, compositeNotifier, logger)
	userUsecase := user_uc.NewUserUsecase(userRepo)

	sch := scheduler.NewScheduler(bookingUsecase, cfg, logger)

	h := &router.Handler{
		EventHandler:   event.NewEventHandler(eventUsecase, logger),
		BookingHandler: booking.NewBookingHandler(bookingUsecase, logger),
		UserHandler:    user.NewUserHandler(userUsecase, logger),
	}

	mux := router.SetupRouter(h)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		cfg:       cfg,
		server:    server,
		logger:    logger,
		db:        db,
		scheduler: sch,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info().Str("addr", a.cfg.Server.Addr).Msg("Starting server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.handleSignals(cancel)

	serverErr := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		a.logger.Error().Err(err).Msg("Server error")
		return err
	case <-ctx.Done():
		a.logger.Info().Msg("Shutting down server")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
		defer shutdownCancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error().Err(err).Msg("Server shutdown failed")
		}

		if a.db != nil && a.db.Master != nil {
			a.db.Master.Close()
		}

		if a.scheduler != nil {
			a.scheduler.Stop()
		}

		a.logger.Info().Msg("Server stopped gracefully")
		return nil
	}
}

func (a *App) handleSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	a.logger.Info().Str("signal", sig.String()).Msg("Received signal")
	cancel()
}
