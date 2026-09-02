package main

import (
	"errors"
	"strconv"
	"strings"

	"backend-go/api-student/app/model"
	"backend-go/api-student/app/repository"
	"github.com/gofiber/fiber/v2"
)

type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

func translateRepositoryError(c *fiber.Ctx, err error, message string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
	default:
		return fail(c, fiber.StatusInternalServerError, message)
	}
}

func validateStudentFields(nim int, name string, grade float64) map[string]string {
	errs := map[string]string{}

	if nim < 1 {
		errs["nim"] = "wajib berupa angka positif"
	}
	if strings.TrimSpace(name) == "" {
		errs["name"] = "wajib diisi"
	}
	if grade < 0 || grade > 100 {
		errs["grade"] = "harus berada di antara 0 dan 100"
	}

	return errs
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	q := parseListQuery(c)
	students, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar student")
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	return okList(c, "daftar student berhasil diambil", students, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func (h *StudentHandler) Get(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err, "gagal mengambil student")
	}

	return ok(c, "student ditemukan", student)
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	errs := validateStudentFields(req.NIM, req.Name, req.Grade)
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student, err := h.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	})
	if err != nil {
		return translateRepositoryError(c, err, "gagal menyimpan student")
	}

	return created(c, "student berhasil dibuat", student,
		"/api/v1/students/"+strconv.Itoa(student.ID))
}

func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	errs := validateStudentFields(req.NIM, req.Name, req.Grade)
	if req.IsActive == nil {
		errs["is_active"] = "wajib dikirim pada PUT"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student, err := h.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: *req.IsActive,
	})
	if err != nil {
		return translateRepositoryError(c, err, "gagal memperbarui student")
	}

	return ok(c, "student berhasil diganti seluruhnya", student)
}

func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err, "gagal mengambil student")
	}

	errs := map[string]string{}
	if req.NIM != nil {
		if *req.NIM < 1 {
			errs["nim"] = "harus berupa angka positif"
		} else {
			student.NIM = *req.NIM
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			student.Name = name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus berada di antara 0 dan 100"
		} else {
			student.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student, err = h.repo.Update(ctx, student)
	if err != nil {
		return translateRepositoryError(c, err, "gagal memperbarui student")
	}

	return ok(c, "student berhasil diperbarui sebagian", student)
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	ctx, cancel := reqCtx(c)
	defer cancel()

	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		return translateRepositoryError(c, err, "gagal menghapus student")
	}

	return noContent(c)
}
