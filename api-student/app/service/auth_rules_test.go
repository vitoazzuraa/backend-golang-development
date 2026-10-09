package service

import (
	"testing"

	"backend-go/api-student/app/model"
	"backend-go/api-student/helper"
)

func TestValidateRegisterWeak(t *testing.T) {
	req := model.RegisterRequest{
		Username: "vito",
		Email:    "vito@example.com",
		Password: "password1",
	}

	errs := helper.ValidateStruct(req)

	if _, ok := errs["password"]; !ok {
		t.Fatalf("want password error, got %v", errs)
	}
}

func TestValidateRegisterShortPassword(t *testing.T) {
	req := model.RegisterRequest{
		Username: "vito",
		Email:    "vito@example.com",
		Password: "a1b2",
	}

	errs := helper.ValidateStruct(req)

	if errs["password"] != "minimal 8 karakter" {
		t.Fatalf("want %q, got %v", "minimal 8 karakter", errs)
	}
}

func TestValidateRegisterValid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "vito",
		Email:    "vito@example.com",
		Password: "admin123123",
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		t.Fatalf("want no error, got %v", errs)
	}
}

func TestValidateRegisterInvalidUsername(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari!",
		Email:    "sari@example.com",
		Password: "admin123123",
	}

	errs := helper.ValidateStruct(req)

	if _, ok := errs["username"]; !ok {
		t.Fatalf("want username error, got %v", errs)
	}
}

func TestValidateLoginEmpty(t *testing.T) {
	errs := helper.ValidateStruct(model.LoginRequest{})

	if len(errs) != 2 {
		t.Fatalf("want 2 errors, got %v", errs)
	}
}
