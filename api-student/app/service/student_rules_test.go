package service

import (
	"testing"

	"backend-go/api-student/app/model"
)

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{10, 10, 1},
		{11, 10, 2},
	}

	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: want %d, got %d", tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestValidateCreate(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM: 0, Name: "",
		Grade: 200,
	}

	errs := ValidateCreate(req)

	if len(errs) == 0 {
		t.Error("want validation errors, got none")
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

	result, errs := ApplyPatch(initial, model.PatchStudentRequest{
		IsActive: &inactive,
	})

	if len(errs) != 0 {
		t.Fatalf("want no error, got %v", errs)
	}
	if result.IsActive {
		t.Error("want IsActive false")
	}
	if result.Name != "Vito" {
		t.Error("want Name unchanged")
	}
}
