package models

import "github.com/pkg/errors"

type User struct {
	UserId   string
	Password string
	Email    string
	AuthType int
}

type UserAuthType uint

const (
	UserAuthTypeNormal UserAuthType = iota + 1
	UserAuthTypeLdap
	UserAuthTypeOauthGoogle
)

type UserDao struct {
}

func (*UserDao) Get(userId string) (User, error) {
	rows, err := dbs.Query("SELECT user_id,password,email,auth_type from users where user_id=?", userId)
	if err != nil {
		return User{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		var rec User
		if err := rows.Scan(&rec.UserId, &rec.Password, &rec.Email, &rec.AuthType); err != nil {
			return User{}, errors.Wrap(err, "scan error")
		}
		return rec, nil
	}
	return User{}, nil
}

func (*UserDao) Insert(user *User) error {
	_, err := dbm.Exec("INSERT INTO users values(?,?,?,?)", user.UserId, "", user.Email, user.AuthType)
	return errors.Wrap(err, "exec error")
}
