package service

import (
	"backend-go/api-student/app/model"
	"backend-go/api-student/helper"
)

func CanAccessStudent(
	current model.AuthUser,
	ownerID *int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if ownerID != nil && current.UserID == *ownerID {
		return true
	}

	return perms.Can(current.Role, anyPermission)
}
