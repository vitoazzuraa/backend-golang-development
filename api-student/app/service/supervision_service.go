package service

import (
	"errors"
	"strconv"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
)

type SupervisionService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisionService(supervisors repository.SupervisorRepository, tx *repository.TxManager) *SupervisionService {
	return &SupervisionService{supervisors: supervisors, tx: tx}
}

func (s *SupervisionService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	supervisors, err := s.supervisors.FindAll(ctx)

	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, "daftar pembimbing berhasil diambil", supervisors)
}

func (s *SupervisionService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	supervisorID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.AssignRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var assigned model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		updated, err := repo.IncrementSupervised(ctx, supervisorID)

		if err != nil {
			return err
		}

		if err := repo.InsertSupervision(ctx, supervisorID, req.StudentID); err != nil {
			return err
		}

		assigned = updated

		return nil
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(c, "mahasiswa berhasil ditetapkan", assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID))
}

func (s *SupervisionService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	var req model.SwitchSupervisionRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if req.FromSupervisorID == req.ToSupervisorID {
		return helper.Validation(map[string]string{
			"to_supervisor_id": "tidak boleh sama dengan from_supervisor_id",
		})
	}

	var joined model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		if err := repo.LockSupervisors(ctx, req.FromSupervisorID, req.ToSupervisorID); err != nil {
			return err
		}

		if err := repo.RemoveSupervision(ctx, req.FromSupervisorID, req.StudentID); err != nil {
			return err
		}

		if _, err := repo.DecrementSupervised(ctx, req.FromSupervisorID); err != nil {
			return err
		}

		updated, err := repo.IncrementSupervised(ctx, req.ToSupervisorID)

		if err != nil {
			return err
		}

		if err := repo.InsertSupervision(ctx, req.ToSupervisorID, req.StudentID); err != nil {
			return err
		}

		joined = updated

		return nil
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Ok(c, "berhasil ganti pembimbing", joined)
}

func (s *SupervisionService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateSupervisorQuotaRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Ok(c, "kuota pembimbing berhasil diubah", updated)
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("pembimbing tidak ditemukan")
	case errors.Is(err, repository.ErrCourseFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa tidak dibimbing oleh pembimbing tersebut")
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data pembimbing sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_ENROLLED",
			"kuota tidak boleh lebih kecil dari jumlah bimbingan saat ini")
	default:
		return helper.Internal(err)
	}
}
