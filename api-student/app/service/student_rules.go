package service

import (
	"strings"

	"backend-go/api-student/app/model"
)

func ValidateStudentFields(nim int, name string, grade float64) map[string]string {
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

func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := ValidateStudentFields(req.NIM, req.Name, req.Grade)

	return errs
}

func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := ValidateStudentFields(req.NIM, req.Name, req.Grade)
	
	if req.IsActive == nil {
		errs["is_active"] = "wajib dikirim pada PUT"
	}
	
	return errs
}

func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}
	
	if req.NIM != nil {
		if *req.NIM < 1 {
			errs["nim"] = "harus berupa angka positif"
		} else {
			current.NIM = *req.NIM
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus berada di antara 0 dan 100"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}

	return (total + limit - 1) / limit
}