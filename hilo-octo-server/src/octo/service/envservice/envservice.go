package envservice

import (
	"errors"
	"octo/models"
	"octo/utils"
)

type EnvService interface {
	GetEnvList(appId int) (utils.List, error)
	CreateEnv(env models.Env) error
	UpdateEnv(env models.Env) error
	DeleteEnv(envId int, appId int) error
	CheckSameEnvironment(appId int, sourceVersionId int, destAppId int, destinationVersionId int, envCheck bool) error
	CheckSameEnvironmentVersion(srcVer models.Version, desVer models.Version, envCheck bool, sync bool) error
	CheckSameEnvironmentForSync(srcAppId int, srcVersionId int, dstAppId int, dstVersionId int, envCheck bool, sync bool) error
	GetSameEnv(srcAppId, dstAppId, srcVersionId int) (models.Env, error)
}
type EnvServiceImpl struct {
	versionDao *models.VersionDao
	envDao     models.EnvDao
}

func NewEnvService() EnvService {
	esi := EnvServiceImpl{envDao: models.NewEnvDao()}
	var es EnvService = &esi
	return es
}

func (e EnvServiceImpl) GetEnvList(appId int) (utils.List, error) {

	return e.envDao.GetList(appId)
}

func (e EnvServiceImpl) CreateEnv(env models.Env) error {
	getEnv, err := e.envDao.GetByName(env.AppId, env.Name)
	if err != nil {
		return err
	}
	if len(getEnv.Name) > 0 {
		return errors.New("EnvName '" + env.Name + "' is exist on App. Input the Another Env name.")
	}
	return e.envDao.Insert(env)
}

func (e EnvServiceImpl) UpdateEnv(env models.Env) error {
	return e.envDao.Update(env)
}

func (e EnvServiceImpl) DeleteEnv(envId int, appId int) error {
	return e.envDao.Delete(envId, appId)
}

func (e EnvServiceImpl) GetEnv(envId int, appId int) (models.Env, error) {
	return e.envDao.Get(appId, envId)
}

func (e EnvServiceImpl) CheckSameEnvironment(srcAppId int, srcVersionId int, dstAppId int, dstVersionId int, envCheck bool) error {

	version, err := e.versionDao.Get(srcAppId, srcVersionId)
	if err != nil {
		return err
	}
	desVersion, err := e.versionDao.Get(dstAppId, dstVersionId)
	if err != nil {
		return err
	}

	return e.CheckSameEnvironmentVersion(version, desVersion, envCheck, false)
}

func (e EnvServiceImpl) CheckSameEnvironmentForSync(srcAppId int, srcVersionId int, dstAppId int, dstVersionId int, envCheck bool, sync bool) error {

	version, err := e.versionDao.Get(srcAppId, srcVersionId)
	if err != nil {
		return err
	}
	desVersion, err := e.versionDao.Get(dstAppId, dstVersionId)
	if !sync {
		if err != nil {
			return err
		}
	}

	return e.CheckSameEnvironmentVersion(version, desVersion, envCheck, sync)
}

func (e EnvServiceImpl) CheckSameEnvironmentVersion(srcVer models.Version, dstVer models.Version, envCheck bool, sync bool) error {

	if !envCheck {
		return nil
	}

	//if !srcVer.EnvId.Valid || int(srcVer.EnvId.Int64) == 0 {
	//	return nil
	//}

	if !srcVer.EnvId.Valid {
		return errors.New("Source Version Environment is null")
	}

	if dstVer.AppId != 0 {
		if !dstVer.EnvId.Valid {
			return errors.New("Destination Version Environment is null")
		}
	}

	if dstVer.AppId != 0 {
		srcEnv, err := e.GetEnv(int(srcVer.EnvId.Int64), srcVer.AppId)
		if err != nil {
			return errors.New("Source Env is None")
		}

		dstEnv, err := e.GetEnv(int(dstVer.EnvId.Int64), dstVer.AppId)
		if err != nil {
			return errors.New("Destination Env is None")
		}
		if srcEnv.Name != dstEnv.Name {
			return errors.New("Environment Name of Destination Version is different from Environment Name of Source Version")
		}
	}

	return nil
}

func (e EnvServiceImpl) GetEnvByName(appId int, envName string) (models.Env, error) {
	env, err := e.envDao.GetByName(appId, envName)
	if err != nil {
		return models.Env{}, err
	}
	return env, nil
}

func (e EnvServiceImpl) GetSameEnv(srcAppId, dstAppId, srcVersionId int) (models.Env, error) {
	srcVersion, err := e.versionDao.Get(srcAppId, srcVersionId)
	if err != nil {
		return models.Env{}, err
	}
	srcEnv, err := e.GetEnv(int(srcVersion.EnvId.Int64), srcAppId)
	if err != nil {
		return models.Env{}, err
	}
	// 原来的APPのEnv名先的APPのEnv寻找名字
	dstEnv, err := e.GetEnvByName(dstAppId, srcEnv.Name)
	if err != nil {
		return models.Env{}, err
	}
	if dstEnv == (models.Env{}) {
		return models.Env{}, err

	}
	return dstEnv, nil
}
