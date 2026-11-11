package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"backend-go/api-student/helper"

	"github.com/gofiber/fiber/v2"
)

type PortfolioService struct {
	students   repository.StudentRepository
	portfolios repository.PortfolioRepository
}

func NewPortfolioService(students repository.StudentRepository, portfolios repository.PortfolioRepository) *PortfolioService {
	return &PortfolioService{students: students, portfolios: portfolios}
}

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	studentID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translateError(err, "student")
	}

	portfolio, err := s.portfolios.FindByStudentID(ctx, studentID)

	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Ok(c, "portofolio ditemukan", portfolio)
}

func (s *PortfolioService) Save(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	studentID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translateError(err, "student")
	}

	var req model.SavePortfolioRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio := model.Portfolio{
		StudentID:    studentID,
		Headline:     req.Headline,
		Bio:          req.Bio,
		Projects:     req.Projects,
		Certificates: req.Certificates,
	}

	if err := s.portfolios.Upsert(ctx, portfolio); err != nil {
		return translatePortfolioError(err)
	}

	return helper.Ok(c, "portofolio berhasil disimpan", portfolio)
}

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	studentID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translateError(err, "student")
	}

	var req model.AddProjectRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	project := model.Project{
		ID:           randomID(),
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	}

	if err := s.portfolios.AddProject(ctx, studentID, project); err != nil {
		return translatePortfolioError(err)
	}

	return helper.Created(c, "proyek berhasil ditambahkan", project,
		"/api/v1/students/"+strconv.Itoa(studentID)+"/portfolio")
}

func (s *PortfolioService) RemoveProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	studentID, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	projectID := c.Params("projectId")

	if projectID == "" {
		return helper.BadRequest("id proyek wajib diisi")
	}

	if err := s.portfolios.RemoveProject(ctx, studentID, projectID); err != nil {
		return translatePortfolioError(err)
	}

	return helper.NoContent(c)
}

func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)

	defer cancel()

	stats, err := s.portfolios.TopTechnologies(ctx, 10)

	if err != nil {
		return translatePortfolioError(err)
	}

	return helper.Ok(c, "statistik teknologi berhasil diambil", stats)
}

func randomID() string {
	buf := make([]byte, 12)

	_, err := rand.Read(buf)

	if err != nil {
		return ""
	}

	return hex.EncodeToString(buf)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("portofolio tidak ditemukan")
	case errors.Is(err, repository.ErrAlreadyExists):
		return helper.Conflict("portofolio untuk mahasiswa ini sudah ada")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT", "jumlah proyek sudah mencapai batas 20")
	case errors.Is(err, repository.ErrProjectMissing):
		return helper.NewError(fiber.StatusNotFound, "PROJECT_NOT_FOUND", "proyek tidak ditemukan")
	case errors.Is(err, repository.ErrInvalidObjectID):
		return helper.BadRequest("format id proyek tidak valid")
	default:
		return helper.Internal(err)
	}
}
