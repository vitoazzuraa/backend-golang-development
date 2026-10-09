package service

import (
	"testing"

	"backend-go/api-student/app/model"
	"backend-go/api-student/helper"
)

func TestValidateCreate(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM: 0, Name: "",
		Grade: 200,
	}

	errs := helper.ValidateStruct(req)

	if len(errs) == 0 {
		t.Error("want validation errors, got none")
	}
}

func TestPatchStudentEmptyNameRejected(t *testing.T) {
	empty := ""

	errs := helper.ValidateStruct(model.PatchStudentRequest{
		Name: &empty,
	})

	if _, ok := errs["name"]; !ok {
		t.Fatalf("want name error, got %v", errs)
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      1,
		Name:     "Vito",
		Grade:    90,
		IsActive: true,
	}

	inactive := false

	result := ApplyPatch(initial, model.PatchStudentRequest{
		IsActive: &inactive,
	})

	if result.IsActive {
		t.Error("want IsActive false")
	}
	if result.Name != "Vito" {
		t.Error("want Name unchanged")
	}
}
