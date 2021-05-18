package service

import "octo/models"

type BucketService struct {
}

func (bucketService *BucketService) GetBucket(appId int) (models.Bucket, error) {
	var bucket models.Bucket
	err := bucketDao.GetBucket(&bucket, appId)
	if err != nil {
		return models.Bucket{}, err
	}
	return bucket, nil
}

//func (*BucketService) CreateBucket(b models.Bucket) error {
//	err := bucketDao.Insert(b)
//	if err != nil {
//		return err
//	}
//	return nil
//}
