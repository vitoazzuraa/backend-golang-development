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
}

func NewCourseService(courses repository.CourseRepository, tx *repository.TxManager) *CourseService {
	return &CourseService{courses: courses, tx: tx}
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

		_, err := repo.DecrementEnrolled(ctx, courseID)

		return err
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.NoContent(c)
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
