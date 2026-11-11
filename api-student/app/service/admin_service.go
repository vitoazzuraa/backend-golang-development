package service

import (
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
)

type AdminService struct {
	outbox repository.OutboxRepository
}

func NewAdminService(outbox repository.OutboxRepository) *AdminService {
	return &AdminService{outbox: outbox}
}

func (s *AdminService) Queue(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	status, err := s.outbox.QueueStatus(ctx)

	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, "status antrean berhasil diambil", status)
}
