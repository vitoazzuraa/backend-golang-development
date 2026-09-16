package service

import (
	"testing"

	"backend-go/api-student/app/model"
)

func TestValidateRegisterWeak(t *testing.T) {
	req := model.RegisterRequest{
		Username: "vito",
		Email: "vito@example.com",
		Password: "password1",
	}
	
	errs := ValidateRegister(req)
	
	if _, ok := errs["password"]; !ok {
		t.Fatalf("want password error, got %v", errs)
	}
}

func TestValidateRegisterValid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "vito", 
		Email: "vito@example.com", 
		Password: "admin123123",
	}
	
	if errs := ValidateRegister(req); len(errs) != 0 {
		t.Fatalf("want no error, got %v", errs)
	}
}

func TestValidateLoginEmpty(t *testing.T) {
	errs := ValidateLogin(model.LoginRequest{})

	if len(errs) != 2 {
		t.Fatalf("want 2 errors, got %v", errs)
	}
}

func TestCheckPasswordStrengthShort(t *testing.T) {
	if got := checkPasswordStrength("a1b2"); got == "" {
		t.Fatal("want error for short password")
	}
}

func TestIsValidUsernameReject(t *testing.T) {
	if isValidUsername("sari!") {
		t.Fatal("want false for username with !")
	}
}
