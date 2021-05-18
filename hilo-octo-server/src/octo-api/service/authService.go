package service

import "octo/models"

var appDao = &models.AppDao{}

type AuthService struct{}

func (s *AuthService) GetAppByClientSecretKey(clientSecretKey string) (models.App, error) {
	return appDao.GetByClientSecretKey(clientSecretKey)
}

func (s *AuthService) GetAppByAppSecretKey(appSecretKey string) (models.App, error) {
	return appDao.GetByAppSecretKey(appSecretKey)
}
