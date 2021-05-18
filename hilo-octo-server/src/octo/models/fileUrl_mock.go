package models

import "github.com/stretchr/testify/mock"

type FileUrlDaoMock struct {
	mock.Mock
}

func (u *FileUrlDaoMock) AddUrl(url FileUrl) error {
	args := u.Called(url)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (u *FileUrlDaoMock) GetListByAppId(appId int) ([]FileUrl, error) {
	args := u.Called(appId)
	if len(args.Get(0).([]FileUrl)) == 0 {
		return []FileUrl{}, args.Error(0)
	}
	return args.Get(0).([]FileUrl), nil
}

func (u *FileUrlDaoMock) GetUrlByObjectNameAndRevisionId(appId int, versionId int, objectName string, revisionId int) (FileUrl, error) {
	args := u.Called(appId, versionId, objectName, revisionId)
	if args.Get(0) != nil {
		return FileUrl{}, args.Error(0)
	}
	return args.Get(0).(FileUrl), nil
}

func (u *FileUrlDaoMock) GetUrlByObjectNameAndRevisionIdLatest(appId int, versionId int, objectName string, revisionId int) (FileUrl, error) {
	args := u.Called(appId, versionId, objectName, revisionId)
	if args.Get(0) != nil {
		return FileUrl{}, args.Error(0)
	}
	return args.Get(0).(FileUrl), nil
}

func (u *FileUrlDaoMock) GetUrlByObjectNameLatest(appId int, versionId int, objectName string) (FileUrl, error) {
	args := u.Called(appId, versionId, objectName)
	if args.Get(0) != nil {
		return FileUrl{}, args.Error(0)
	}
	return args.Get(0).(FileUrl), nil
}

func (u *FileUrlDaoMock) UpdateCrcAndMd5(appId int, versionId int, objectName string, revisionId int, crc uint32, md5 string) error {
	args := u.Called(appId, versionId, objectName, revisionId, crc, md5)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}
