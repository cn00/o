package models

import (
	"octo/utils"

	"github.com/pkg/errors"
)

type UserApp struct {
	AppId    int
	UserId   string
	RoleType int
}

type UserApps []UserApp
type UserAppIds []int
type UserAppRoleType uint

const (
	UserRoleTypeUser UserAppRoleType = iota + 1
	UserRoleTypeAdmin
	UserRoleTypeReader
)

type UserAppDao struct {
}

func (*UserAppDao) table() string {
	return "user_apps"
}

func (*UserAppDao) GetByUserId(userId string) (UserApps, error) {
	rows, err := dbs.Query("SELECT app_id,user_id,role_type from user_apps where user_id=?", userId)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	var res UserApps
	for rows.Next() {
		var rec UserApp
		if err := rows.Scan(&rec.AppId, &rec.UserId, &rec.RoleType); err != nil {
			return UserApps{}, errors.Wrap(err, "scan error")
		}
		res = append(res, rec)
	}
	return res, nil
}

func (userApps UserApps) GetAppIds() UserAppIds {
	var res UserAppIds
	for _, userApp := range userApps {
		res = append(res, userApp.AppId)
	}
	return res
}

func (slice UserAppIds) Position(value int) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func (*UserAppDao) GetListByAppId(appId int) (utils.List, error) {
	rows, err := dbs.Query(`SELECT app_id,user_id,role_type from user_apps where app_id=?`, appId)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res utils.List
	for rows.Next() {
		var rec UserApp
		if err := rows.Scan(&rec.AppId, &rec.UserId, &rec.RoleType); err != nil {
			return nil, errors.Wrap(err, "scan error")
		}
		res = append(res, rec)
	}

	return res, nil

}

func (*UserAppDao) Add(appId int, userId string, roleType int) error {
	_, err := dbm.Exec(`INSERT INTO user_apps (app_id,user_id,role_type) values (?,?,?)`,
		appId,
		userId,
		roleType,
	)
	return errors.Wrap(err, "exec error")
}

func (*UserAppDao) Delete(appId int, userId string) error {
	_, err := dbm.Exec(`DELETE FROM user_apps WHERE app_id=? and user_id=?`,
		appId,
		userId,
	)
	return errors.Wrap(err, "exec error")
}
