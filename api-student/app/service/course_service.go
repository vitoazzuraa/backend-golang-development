package service

import (
	"errors"
	"strconv"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
)

type CourseService struct {
	courses repository.CourseRepository
	tx      *repository.TxManager
	outbox  repository.OutboxRepository
}

func NewCourseService(courses repository.CourseRepository, tx *repository.TxManager, outbox repository.OutboxRepository) *CourseService {
	return &CourseService{courses: courses, tx: tx, outbox: outbox}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	courses, err := s.courses.FindAll(ctx)

	if err != nil {
		return helper.Internal(err)
	}

	return helper.Ok(c, "daftar kelas berhasil diambil", courses)
}

func (s *CourseService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courses, err := s.courses.FindByStudent(ctx, current.UserID)

	if err != nil {
		return helper.Internal(err)
	}

	total := 0

	for _, course := range courses {
		total += course.Credits
	}

	return helper.Ok(c, "KRS berhasil diambil", fiber.Map{
		"courses":       courses,
		"total_credits": total,
		"max_credits":   MaxCreditsPerSemester,
	})
}

func (s *CourseService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	course, err := s.courses.FindByID(ctx, id)

	if err != nil {
		return translateCourseError(err)
	}

	return helper.Ok(c, "kelas ditemukan", course)
}

var errCreditLimit = errors.New("batas SKS terlampaui")

func (s *CourseService) Enroll(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courseID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var enrolled model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		if err := repo.LockStudent(ctx, current.UserID); err != nil {
			return err
		}

		course, err := repo.FindByID(ctx, courseID)

		if err != nil {
			return err
		}

		taken, err := repo.SumCredits(ctx, current.UserID)

		if err != nil {
			return err
		}

		if !WithinCreditLimit(taken, course.Credits) {
			return errCreditLimit
		}

		enrolled, err = repo.IncrementEnrolled(ctx, courseID)

		if err != nil {
			return err
		}

		event := model.EnrollmentEvent{
			UserID:     current.UserID,
			CourseID:   enrolled.ID,
			CourseCode: enrolled.Code,
			CourseName: enrolled.Name,
		}

		if err := s.outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentCreated, event); err != nil {
			return err
		}

		return repo.InsertEnrollment(ctx, courseID, current.UserID)
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.Created(c, "berhasil mendaftar kelas", enrolled,
		"/api/v1/courses/"+strconv.Itoa(courseID))
}

func (s *CourseService) Drop(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courseID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		if err := repo.LockStudent(ctx, current.UserID); err != nil {
			return err
		}

		if err := repo.DeleteEnrollment(ctx, courseID, current.UserID); err != nil {
			return err
		}

		course, err := repo.DecrementEnrolled(ctx, courseID)

		if err != nil {
			return err
		}

		event := model.EnrollmentEvent{
			UserID:     current.UserID,
			CourseID:   course.ID,
			CourseCode: course.Code,
			CourseName: course.Name,
		}

		return s.outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentDropped, event)
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.NoContent(c)
}

func (s *CourseService) Swap(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	current, ok := helper.CurrentUser(c)

	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.SwapRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if req.FromCourseID == req.ToCourseID {
		return helper.Validation(map[string]string{
			"to_course_id": "tidak boleh sama dengan from_course_id",
		})
	}

	var joined model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		if err := repo.LockStudent(ctx, current.UserID); err != nil {
			return err
		}

		if err := repo.LockCourses(ctx, req.FromCourseID, req.ToCourseID); err != nil {
			return err
		}

		from, err := repo.FindByID(ctx, req.FromCourseID)

		if err != nil {
			return err
		}

		to, err := repo.FindByID(ctx, req.ToCourseID)

		if err != nil {
			return err
		}

		taken, err := repo.SumCredits(ctx, current.UserID)

		if err != nil {
			return err
		}

		if CreditsAfterSwap(taken, from.Credits, to.Credits) > MaxCreditsPerSemester {
			return errCreditLimit
		}

		if err := repo.DeleteEnrollment(ctx, req.FromCourseID, current.UserID); err != nil {
			return err
		}

		if _, err := repo.DecrementEnrolled(ctx, req.FromCourseID); err != nil {
			return err
		}

		joined, err = repo.IncrementEnrolled(ctx, req.ToCourseID)

		if err != nil {
			return err
		}

		if err := repo.InsertEnrollment(ctx, req.ToCourseID, current.UserID); err != nil {
			return err
		}

		event := model.EnrollmentEvent{
			UserID:         current.UserID,
			CourseID:       to.ID,
			CourseCode:     to.Code,
			CourseName:     to.Name,
			FromCourseID:   from.ID,
			FromCourseCode: from.Code,
			FromCourseName: from.Name,
		}

		return s.outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentSwapped, event)
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.Ok(c, "berhasil pindah kelas", joined)
}

func (s *CourseService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.UpdateQuotaRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var updated model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		course, err := repo.UpdateQuota(ctx, id, req.Quota, req.Version)

		if err != nil {
			return err
		}

		updated = course

		event := model.EnrollmentEvent{
			CourseID:   course.ID,
			CourseCode: course.Code,
			CourseName: course.Name,
			Quota:      course.Quota,
		}

		return s.outbox.WithTx(tx).Insert(ctx, model.EventQuotaChanged, event)
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.Ok(c, "kuota berhasil diubah", updated)
}

func translateCourseError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("kelas tidak ditemukan")
	case errors.Is(err, repository.ErrCourseFull):
		return helper.NewError(fiber.StatusConflict, "COURSE_FULL", "kuota kelas sudah penuh")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.NewError(fiber.StatusConflict, "ALREADY_ENROLLED", "Anda sudah terdaftar di kelas ini")
	case errors.Is(err, repository.ErrNotEnrolled):
		return helper.NewError(fiber.StatusConflict, "NOT_ENROLLED", "Anda tidak terdaftar di kelas tersebut")
	case errors.Is(err, errCreditLimit):
		return helper.NewError(fiber.StatusConflict, "CREDIT_LIMIT_EXCEEDED",
			"batas "+strconv.Itoa(MaxCreditsPerSemester)+" SKS terlampaui")
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data kelas sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_ENROLLED",
			"kuota tidak boleh lebih kecil dari jumlah peserta saat ini")
	default:
		return helper.Internal(err)
	}
}
