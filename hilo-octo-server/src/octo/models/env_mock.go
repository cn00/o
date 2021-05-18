package models

import (
	"octo/utils"

	"github.com/stretchr/testify/mock"
)

type EnvDaoMock struct {
	mock.Mock
}

func (e *EnvDaoMock) GetList(appId int) (utils.List, error) {
	args := e.Called(appId)
	if len(args.Get(0).(utils.List)) == 0 {
		return utils.List{}, args.Error(1)
	}
	return args.Get(0).(utils.List), nil
}

func (e *EnvDaoMock) Get(appId int, envId int) (Env, error) {
	args := e.Called(appId, envId)
	if args.Get(0).(Env).AppId == 0{
		return Env{}, args.Error(1)
	}
	return args.Get(0).(Env), nil
}

func (e *EnvDaoMock) GetByName(appId int, name string) (Env, error) {
	args := e.Called(appId, name)
	return args.Get(0).(Env), nil
}

func (e *EnvDaoMock) Insert(env Env) error {
	args := e.Called(env)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvDaoMock) Update(env Env) error {
	args := e.Called(env)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (e *EnvDaoMock) Delete(envId int, appId int) error {
	args := e.Called(envId, appId)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}
