package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students []Student
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

func cocokPencarian(student Student, kata string) bool {
	return strings.Contains(strings.ToLower(student.Name), strings.ToLower(kata))
}

func lessStudent(a, b Student, field string) bool {
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
	filtered := []Student{}
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

	return okList(c, "daftar student berhasil diambil", filtered[start:end], &Meta{
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
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.NIM < 1 || req.Name == "" {
		return fail(c, fiber.StatusBadRequest, "nim dan name wajib diisi")
	}

	student := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	}
	students = append(students, student)
	nextID++

	return created(c, "student berhasil dibuat", student, "/api/v1/students/"+strconv.Itoa(student.ID))
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

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.NIM < 1 || req.Name == "" || req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "nim, name, dan is_active wajib dikirim pada PUT")
	}

	students[i] = Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: *req.IsActive,
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

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	updated := students[i]
	if req.NIM != nil {
		updated.NIM = *req.NIM
	}
	if req.Name != nil {
		updated.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		updated.Grade = *req.Grade
	}
	if req.IsActive != nil {
		updated.IsActive = *req.IsActive
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
