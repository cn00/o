package service

import (
	"testing"

	"octo/models"
)

func TestUserServiceCheckAuthority(t *testing.T) {
	userApps := models.UserApps{
		models.UserApp{
			AppId:    1,
			RoleType: int(models.UserRoleTypeReader),
		},
		models.UserApp{
			AppId:    2,
			RoleType: int(models.UserRoleTypeUser),
		},
		models.UserApp{
			AppId:    3,
			RoleType: int(models.UserRoleTypeAdmin),
		},
	}

	patterns := []struct {
		appId    int
		roleType models.UserAppRoleType
		expected bool
	}{
		{
			appId:    1,
			roleType: models.UserRoleTypeReader,
			expected: true,
		},
		{
			appId:    1,
			roleType: models.UserRoleTypeUser,
			expected: false,
		},
		{
			appId:    1,
			roleType: models.UserRoleTypeAdmin,
			expected: false,
		},
		{
			appId:    2,
			roleType: models.UserRoleTypeReader,
			expected: true,
		},
		{
			appId:    2,
			roleType: models.UserRoleTypeUser,
			expected: true,
		},
		{
			appId:    2,
			roleType: models.UserRoleTypeAdmin,
			expected: false,
		},
		{
			appId:    3,
			roleType: models.UserRoleTypeReader,
			expected: true,
		},
		{
			appId:    3,
			roleType: models.UserRoleTypeUser,
			expected: true,
		},
		{
			appId:    3,
			roleType: models.UserRoleTypeAdmin,
			expected: true,
		},
		{
			appId:    4,
			roleType: models.UserRoleTypeReader,
			expected: false,
		},
		{
			appId:    4,
			roleType: models.UserRoleTypeUser,
			expected: false,
		},
		{
			appId:    4,
			roleType: models.UserRoleTypeAdmin,
			expected: false,
		},
	}

	for i, pattern := range patterns {
		if userService.checkAuthority(userApps, pattern.appId, pattern.roleType) != pattern.expected {
			t.Fatalf("#%d: checkAuthority is wrong: %+v", i, pattern)
		}
	}
}
