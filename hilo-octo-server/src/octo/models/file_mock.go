package models

import "github.com/stretchr/testify/mock"

type FileMockDao struct {
	mock.Mock
}

func (f *FileMockDao) GetList(appId int, versionId int, revisionId int) ([]File, error) {
	args := f.Called(appId, versionId, revisionId)
	if len(args.Get(0).([]File)) == 0 {
		return []File{}, args.Error(0)
	}
	return args.Get(0).([]File), nil
}

func (f *FileMockDao) GetRangeList(appId int, versionId int, revisionId int, fromDate string, toDate string) ([]File, error) {
	args := f.Called(appId, versionId, revisionId, fromDate, toDate)
	if len(args.Get(0).([]File)) == 0 {
		return []File{}, args.Error(0)
	}
	return args.Get(0).([]File), nil
}

func (f *FileMockDao) GetDiffList(appId int, versionId int, revisionId int, targetRevisionId int) ([]File, error) {
	args := f.Called(appId, versionId, revisionId, targetRevisionId)
	if len(args.Get(0).([]File)) == 0 {
		return []File{}, args.Error(0)
	}
	return args.Get(0).([]File), nil
}

func (f *FileMockDao) GetMaxRevisionId(appId int, versionId int) (int, error) {
	args := f.Called(appId, versionId)
	return args.Int(0), nil
}

func (f *FileMockDao) GetByName(appId int, versionId int, name string) (File, error) {
	args := f.Called(appId, versionId, name)
	return args.Get(0).(File), nil
}

func (f *FileMockDao) GetByNameForUpdate(appId int, versionId int, name string) (File, error) {
	args := f.Called(appId, versionId, name)
	return args.Get(0).(File), nil
}
