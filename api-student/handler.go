package main

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"backend-go/api-student/app/model"
	"github.com/gofiber/fiber/v2"
)

var students []model.Student
var nextID = 1

func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}

	return -1
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

func nimExists(nim, exceptID int) bool {
	for _, student := range students {
		if student.ID != exceptID && student.NIM == nim {
			return true
		}
	}

	return false
}

func cocokPencarian(student model.Student, kata string) bool {
	return strings.Contains(strings.ToLower(student.Name), strings.ToLower(kata))
}

func lessStudent(a, b model.Student, field string) bool {
	switch field {
	case "nim":
		if a.NIM != b.NIM {
			return a.NIM < b.NIM
		}
	case "name":
		if a.Name != b.Name {
			return a.Name < b.Name
		}
	case "grade":
		if a.Grade != b.Grade {
			return a.Grade < b.Grade
		}
	}

	return a.ID < b.ID
}

func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)
	filtered := []model.Student{}
	for _, student := range students {
		if q.IsActive != nil && student.IsActive != *q.IsActive {
			continue
		}
		if q.Search != "" && !cocokPencarian(student, q.Search) {
			continue
		}
		filtered = append(filtered, student)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if q.Order == "desc" {
			return lessStudent(filtered[j], filtered[i], q.Sort)
		}
		return lessStudent(filtered[i], filtered[j], q.Sort)
	})

	total := len(filtered)
	totalPages := 0
	if total > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	start := (q.Page - 1) * q.Limit
	if start > total {
		start = total
	}

	end := start + q.Limit
	if end > total {
		end = total
	}

	return okList(c, "daftar student berhasil diambil", filtered[start:end], &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	return ok(c, "student ditemukan", students[i])
}

func createStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	errs := validateStudentFields(req.NIM, req.Name, req.Grade)
	if len(errs) > 0 {
		return failValidation(c, errs)
	}
	if nimExists(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
	}

	student := model.Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	students = append(students, student)
	nextID++

	return created(c, "student berhasil dibuat", student, "/api/v1/students/"+strconv.Itoa(student.ID))
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

func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
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
	if nimExists(req.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
	}

	students[i] = model.Student{
		ID:        id,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  *req.IsActive,
		CreatedAt: students[i].CreatedAt,
	}

	return ok(c, "student berhasil diganti seluruhnya", students[i])
}

func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	updated := students[i]
	errs := map[string]string{}
	if req.NIM != nil {
		if *req.NIM < 1 {
			errs["nim"] = "harus berupa angka positif"
		} else {
			updated.NIM = *req.NIM
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			updated.Name = name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus berada di antara 0 dan 100"
		} else {
			updated.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}
	if req.NIM != nil && nimExists(updated.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah digunakan")
	}

	students[i] = updated

	return ok(c, "student berhasil diperbarui sebagian", students[i])
}

func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	students = append(students[:i], students[i+1:]...)

	return noContent(c)
}
