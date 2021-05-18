package models

import (
	"database/sql"

	"octo/utils"

	"github.com/pkg/errors"
)

type Env struct {
	AppId  int
	EnvId  int
	Name   string
	Detail sql.NullString
}

type EnvDao interface {
	Dao
	Get(appId int, envId int) (Env, error)
	GetList(appId int) (utils.List, error)
	GetByName(appId int, name string) (Env, error)
	Insert(env Env) error
	Update(env Env) error
	Delete(envId int, appId int) error
}

type EnvDaoImpl struct {
	envScanner
}

func NewEnvDao() EnvDao {
	ei := EnvDaoImpl{}
	var edo EnvDao = &ei
	return edo
}

func (*EnvDaoImpl) columns() string {
	return "app_id, env_id, name, detail"
}

func (*EnvDaoImpl) table() string {
	return "envs"
}

type envScanner struct {}

func (*envScanner) scan(rows *sql.Rows) (Env, error) {
	var e Env
	err := rows.Scan(&e.AppId, &e.EnvId, &e.Name, &e.Detail)
	return e, errors.Wrap(err, "scan error")
}

func (s *envScanner) scanEnv(rows *sql.Rows, err error) (Env, error) {
	if err != nil {
		return Env{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return s.scan(rows)
	}
	return Env{}, nil
}

func (dao *EnvDaoImpl) Get(appId int, envId int) (Env, error) {
	rows, err := dbs.Query("SELECT "+dao.columns()+" FROM "+dao.table()+" WHERE app_id =? and env_id = ?", appId, envId)
	return dao.scanEnv(rows, err)
}

func (dao *EnvDaoImpl) GetList(appId int) (utils.List, error) {
	rows, err := dbs.Query("SELECT "+dao.columns()+" FROM "+dao.table()+" WHERE app_id =?", appId)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	var res utils.List
	for rows.Next() {
		env, err := dao.scan(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scan error")
		}
		res = append(res, env)
	}

	return res, nil
}

func (dao *EnvDaoImpl) GetByName(appId int, name string) (Env, error) {
	rows, err := dbs.Query("SELECT "+dao.columns()+" FROM "+dao.table()+" WHERE app_id = ? and name = ?", appId, name)
	return dao.scanEnv(rows, err)
}

func (*EnvDaoImpl) Insert(env Env) error {
	_, err := dbm.Exec("INSERT INTO envs(app_id, name, detail) VALUES(?, ?, ?)", env.AppId, env.Name, env.Detail.String)
	return errors.Wrap(err, "exec error")
}

func (*EnvDaoImpl) Update(env Env) error {
	_, err := dbm.Exec("UPDATE envs SET name = ?, detail = ? WHERE env_id = ? and app_id = ?",
		env.Name, env.Detail, env.EnvId, env.AppId)
	return errors.Wrap(err, "exec error")
}

func (*EnvDaoImpl) Delete(envId int, appId int) error {
	_, err := dbm.Exec("DELETE FROM envs WHERE env_id = ? and app_id = ?", envId, appId)
	return errors.Wrap(err, "exec error")
}
