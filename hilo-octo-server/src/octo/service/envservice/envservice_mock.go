package envservice

import (
	"github.com/stretchr/testify/mock"

	"octo/models"
	"octo/utils"

	"github.com/pkg/errors"
)

type EnvServiceMock struct {
	mock.Mock
}

func (e *EnvServiceMock) GetEnvList(appId int) (utils.List, error) {
	args := e.Called(appId)
	if len(args.Get(0).(utils.List)) == 0 {
		return nil, errors.New("List size nul")
	}
	if args.Get(1).(error) != nil {
		return nil, args.Get(1).(error)
	}
	return args.Get(0).(utils.List), nil
}

func (e *EnvServiceMock) CreateEnv(env models.Env) error {
	args := e.Called(env)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvServiceMock) UpdateEnv(env models.Env) error {
	args := e.Called(env)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvServiceMock) DeleteEnv(envId int, appId int) error {
	args := e.Called(envId, appId)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvServiceMock) CheckSameEnvironment(appId int, sourceVersionId int, destAppId int, destVersionId int) error {
	args := e.Called(appId, sourceVersionId, destAppId, destVersionId)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvServiceMock) CheckSameEnvironmentVersion(srcVer models.Version, desVer models.Version, sync bool) error {
	args := e.Called(srcVer, desVer, false)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvServiceMock) CheckSameEnvironmentForSync(appId int, sourceVersionId int, destAppId int, destVersionId int, sync bool) error {
	args := e.Called(appId, sourceVersionId, destAppId, destVersionId, sync)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}
