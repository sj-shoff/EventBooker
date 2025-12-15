package scheduler

import (
	"context"

	"event-booker/internal/config"

	"github.com/robfig/cron/v3"
	"github.com/wb-go/wbf/zlog"
)

type Scheduler struct {
	bookingUsecase bookingUsecase
	cfg            *config.Config
	logger         *zlog.Zerolog
	cron           *cron.Cron
}

func NewScheduler(bookingUsecase bookingUsecase, cfg *config.Config, logger *zlog.Zerolog) *Scheduler {
	return &Scheduler{
		bookingUsecase: bookingUsecase,
		cfg:            cfg,
		logger:         logger,
		cron:           cron.New(),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	_, err := s.cron.AddFunc("@every "+s.cfg.Scheduler.CleanupInterval.String(), func() {
		s.cleanupExpiredBookings(ctx)
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to add cron job")
	}
	s.cron.Start()
	s.logger.Info().Msg("Scheduler started")
}

func (s *Scheduler) cleanupExpiredBookings(ctx context.Context) {
	expired, err := s.bookingUsecase.GetExpiredBookings(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expired bookings")
		return
	}
	for _, b := range expired {
		if err := s.bookingUsecase.CancelBooking(ctx, b.ID); err != nil {
			s.logger.Error().Err(err).Str("booking_id", b.ID).Msg("Failed to cancel expired booking")
		} else {
			s.logger.Info().Str("booking_id", b.ID).Msg("Expired booking cancelled")
		}
	}
}
