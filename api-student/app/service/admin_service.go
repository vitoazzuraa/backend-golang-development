package service

import (
	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

type AdminService struct {
	outbox    repository.OutboxRepository
	inspector *asynq.Inspector
}

func NewAdminService(outbox repository.OutboxRepository, inspector *asynq.Inspector) *AdminService {
	return &AdminService{outbox: outbox, inspector: inspector}
}

func (s *AdminService) Queue(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	stats, err := s.outbox.Stats(ctx)

	if err != nil {
		return helper.Internal(err)
	}

	info, err := s.inspector.GetQueueInfo(model.QueueNotifications)

	if err != nil {
		stats.QueueError = "Redis tidak dapat dihubungi: " + err.Error()

		return helper.Ok(c, "status antrean berhasil diambil", stats)
	}

	stats.Queue = &model.QueueInfo{
		Name:      info.Queue,
		Pending:   info.Pending,
		Active:    info.Active,
		Scheduled: info.Scheduled,
		Retry:     info.Retry,
		Archived:  info.Archived,
		Completed: info.Completed,
		Processed: info.Processed,
		Failed:    info.Failed,
	}

	return helper.Ok(c, "status antrean berhasil diambil", stats)
}
