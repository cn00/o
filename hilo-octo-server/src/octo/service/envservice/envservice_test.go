package envservice

import (
	"database/sql"
	"testing"

	"octo/models"

	"octo/utils"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CheckEnvVersion struct {
	mock.Mock
}

func (dest *CheckEnvVersion) GetDestVersion(destAppId int, destVersionId int, envId int) models.Version {
	args := dest.Called(destAppId, destVersionId, envId)
	v := args.Get(0).(models.Version)
	v.AppId = destAppId
	v.VersionId = destVersionId
	v.EnvId.Valid = true
	v.EnvId.Int64 = int64(envId)
	return v
}

func (src *CheckEnvVersion) GetSrcVersion(srcAppId int, srcVersionId int, envId int) models.Version {
	args := src.Called(srcAppId, srcVersionId, envId)
	v := args.Get(0).(models.Version)
	v.AppId = srcAppId
	v.VersionId = srcVersionId
	v.EnvId.Valid = true
	v.EnvId.Int64 = int64(envId)
	return v
}

func TestEnvService_CheckSameEnvironment(t *testing.T) {
	//setupEnv()
	obj := new(CheckEnvVersion)
	obj.On("GetDestVersion", 1, 1, 2).Return(models.Version{})
	obj.On("GetSrcVersion", 1, 2, 3).Return(models.Version{})

	destVersion := obj.GetDestVersion(1, 1, 2)
	srcVersion := obj.GetSrcVersion(1, 2, 3)
	envDaoMock := new(models.EnvDaoMock)
	// 渡したメソッド名が呼ばれたときに定義したリターンの値を返すように
	srcEnv := models.Env{EnvId: 2, AppId: 1, Name: "Win", Detail: sql.NullString{String: "Win"}}
	desEnv := models.Env{EnvId: 3, AppId: 1, Name: "Win2", Detail: sql.NullString{String: "Win2"}}
	envDaoMock.On("Get", 1, 2).Return(srcEnv)
	envDaoMock.On("Get", 1, 3).Return(desEnv)
	var e = EnvServiceImpl{envDao: envDaoMock}
	err := e.CheckSameEnvironmentVersion(srcVersion, destVersion, true, false)
	assert.Error(t, err, "No Error is unnormal")
}

func TestEnvServiceImpl_GetEnvList(t *testing.T) {
	assert := assert.New(t)
	envDaoMock := new(models.EnvDaoMock)
	es := EnvServiceImpl{envDao: envDaoMock}
	returnValue := utils.List{&models.Env{}, &models.Env{}}
	// 渡したメソッド名が呼ばれたときに定義したリターンの値を返すように
	envDaoMock.On("GetList", 1).Return(returnValue)
	result, err := es.GetEnvList(1)
	assert.Equal(returnValue, result)
	assert.NoError(err)

	returnValue = utils.List{}
	envDaoMock.On("GetList", 2).Return(returnValue, errors.New("Has Err"))
	result, err = es.GetEnvList(2)

	assert.Equal(returnValue, result)
	assert.Error(err)

}

func TestEnvServiceImpl_Insert(t *testing.T) {
	assert := assert.New(t)
	m := new(models.EnvDaoMock)
	es := EnvServiceImpl{envDao: m}
	rvName := models.Env{}
	// 渡したメソッド名が呼ばれたときに定義したリターンの値を返すように
	m.On("GetByName", 1, "Win").Return(rvName)
	insertValue := models.Env{EnvId: 1, AppId: 1, Name: "Win", Detail: sql.NullString{String: "Win"}}
	m.On("Insert", insertValue).Return(nil)
	result := es.CreateEnv(insertValue)
	assert.Equal(result, nil)

}

func TestEnvServiceImpl_CreateEnv(t *testing.T) {
	assert := assert.New(t)
	m := new(models.EnvDaoMock)
	es := EnvServiceImpl{envDao: m}
	rvName := models.Env{
		AppId:  1,
		EnvId:  2,
		Name:   "Win",
		Detail: sql.NullString{String: "Win"},
	}

	m.On("GetByName", 1, "Win").Return(rvName)

	insertValue := models.Env{EnvId: 1, AppId: 1, Name: "Win", Detail: sql.NullString{String: "Win"}}
	result := es.CreateEnv(insertValue)
	assert.Error(result)
}

//TODO 各ゲームがEnvの設定対応が終わるまでコメントアウト
func TestEnvServiceImpl_CheckSameEnvironmentVersion(t *testing.T) {
	var envCheck = true
	assert := assert.New(t)
	envDaoMock := new(models.EnvDaoMock)
	es := EnvServiceImpl{envDao: envDaoMock}
	srcEnv := models.Env{EnvId: 1, AppId: 1, Name: "Win", Detail: sql.NullString{String: "Win"}}
	desEnv := models.Env{EnvId: 1, AppId: 1, Name: "Win", Detail: sql.NullString{String: "Win"}}
	envDaoMock.On("Get", 1, 1).Return(srcEnv)
	envDaoMock.On("Get", 1, 1).Return(desEnv)

	vDest := &models.Version{
		AppId:         1,
		VersionId:     1,
		Description:   "",
		MaxRevision:   0,
		CopyVersionId: sql.NullInt64{},
		CopyAppId:     sql.NullInt64{},
		EnvId:         sql.NullInt64{Int64: 1, Valid: true},
	}
	vSrc := &models.Version{
		AppId:         1,
		VersionId:     2,
		Description:   "",
		MaxRevision:   0,
		CopyVersionId: sql.NullInt64{},
		CopyAppId:     sql.NullInt64{},
		EnvId:         sql.NullInt64{Int64: 1, Valid: true},
	}
	err := es.CheckSameEnvironmentVersion(*vSrc, *vDest, envCheck, false)
	assert.NoError(err)

	srcEnv = models.Env{EnvId: 1, AppId: 1, Name: "Win", Detail: sql.NullString{String: "Win"}}
	desEnv = models.Env{EnvId: 2, AppId: 1, Name: "Win2", Detail: sql.NullString{String: "Win2"}}
	envDaoMock.On("Get", 1, 1).Return(srcEnv)
	envDaoMock.On("Get", 1, 2).Return(desEnv)
	vDest = &models.Version{
		AppId:         1,
		VersionId:     1,
		Description:   "",
		MaxRevision:   0,
		CopyVersionId: sql.NullInt64{},
		CopyAppId:     sql.NullInt64{},
		EnvId:         sql.NullInt64{Int64: 2, Valid: true},
	}
	vSrc = &models.Version{
		AppId:         1,
		VersionId:     2,
		Description:   "",
		MaxRevision:   0,
		CopyVersionId: sql.NullInt64{},
		CopyAppId:     sql.NullInt64{},
		EnvId:         sql.NullInt64{Int64: 1, Valid: true},
	}
	err = es.CheckSameEnvironmentVersion(*vSrc, *vDest, envCheck, false)
	assert.Error(err)

}
