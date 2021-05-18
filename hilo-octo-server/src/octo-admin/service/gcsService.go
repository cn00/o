package service

import "octo/models"

type GcsService struct {
}

func (*GcsService) GetGCS(appId int) (models.Gcs, error) {
	var gcs models.Gcs
	err := gcsDao.GetGcs(&gcs, appId)
	if err != nil {
		return models.Gcs{}, err
	}
	return gcs, nil
}

//func (*GcsService) CreateGCS(g models.Gcs) error {
//
//	err := gcsDao.Insert(g)
//	if err != nil {
//		return err
//	}
//	return nil
//}
