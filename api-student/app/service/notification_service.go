package service

import (
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
)

type NotificationService struct {
	notifications repository.NotificationRepository
}

func NewNotificationService(notifications repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

func (s *NotificationService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	notifications, err := s.notifications.FindByUser(ctx, current.UserID)

	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, "daftar notifikasi berhasil diambil", notifications)
}
