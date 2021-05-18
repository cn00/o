package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type App struct {
	AppId           int
	AppName         string
	Description     string
	ImageUrl        sql.NullString
	Email           sql.NullString
	AppSecretKey    string
	ClientSecretKey string
	AesKey          string
	StorageType     int
}

type AppDao struct {
	appScanner
}

func (*AppDao) table() string {
	return "apps"
}

type appScanner struct {}

func (*appScanner) scan(rows *sql.Rows) (App, error) {
	var rec App
	err := rows.Scan(&rec.AppId, &rec.AppName, &rec.Description, &rec.ImageUrl, &rec.Email, &rec.AppSecretKey, &rec.ClientSecretKey, &rec.AesKey, &rec.StorageType)
	return rec, errors.Wrap(err, "scan error")
}

func (s *appScanner) scanApp(rows *sql.Rows, err error) (App, error) {
	if err != nil {
		return App{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return s.scan(rows)
	}
	return App{}, nil
}

func (s *appScanner) scanApps(rows *sql.Rows, err error) ([]App, error) {
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res []App
	for rows.Next() {
		file, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, file)
	}
	return res, nil
}

func (*AppDao) Get(a *App, appId int) error {
	err := dbs.QueryRow("SELECT app_id,app_name,description,image_url,email,app_secret_key,client_secret_key,aes_key,storage_type from apps where app_id=?", appId).Scan(&a.AppId, &a.AppName, &a.Description, &a.ImageUrl, &a.Email, &a.AppSecretKey, &a.ClientSecretKey, &a.AesKey, &a.StorageType)
	return errors.Wrap(err, "query row error")
}

func (dao *AppDao) GetByClientSecretKey(clientSecretKey string) (App, error) {
	rows, err := dbs.Query("SELECT app_id,app_name,description,image_url,email,app_secret_key,client_secret_key,aes_key,storage_type from apps where client_secret_key=?", clientSecretKey)
	return dao.scanApp(rows, err)
}

func (dao *AppDao) GetByAppSecretKey(appSecretKey string) (App, error) {
	rows, err := dbs.Query("SELECT app_id,app_name,description,image_url,email,app_secret_key,client_secret_key,aes_key,storage_type from apps where app_secret_key=?", appSecretKey)
	return dao.scanApp(rows, err)
}

func (dao *AppDao) GetAllList() ([]App, error) {
	rows, err := dbs.Query("SELECT app_id,app_name,description,image_url,email,app_secret_key,client_secret_key,aes_key,storage_type from apps")
	return dao.scanApps(rows, err)
}

func (dao *AppDao) GetListByIds(appIds []int) ([]App, error) {
	str := ""
	for _, v := range appIds {
		str += fmt.Sprintf("%d,", v)
	}
	str = strings.TrimRight(str, ",")
	rows, err := dbs.Query("SELECT app_id,app_name,description,image_url,email,app_secret_key,client_secret_key,aes_key,storage_type from apps where app_id in (" + str + ")")
	return dao.scanApps(rows, err)
}

func (*AppDao) Update(appId int, name string, description string, imageUrl string) error {
	_, err := dbm.Exec(`UPDATE apps SET
						app_name=?,
						description=?,
						image_url=?
						WHERE app_id=?`,
		name, description, imageUrl, appId)
	return errors.Wrap(err, "exec error")
}

func (*AppDao) UpdateApp(app App, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE apps SET app_name=?, description=?, image_url=? WHERE app_id=?`, app.AppName, app.Description, app.ImageUrl.String, app.AppId)
	return errors.Wrap(err, "update exec error")

}
func (*AppDao) Insert(insApp App, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT INTO apps(
	app_id, app_name, description, image_url, email,
	 app_secret_key, client_secret_key, aes_key, storage_type)
	 VALUES(?,?,?,?,?,?,?,?,?)`,
		insApp.AppId, insApp.AppName, insApp.Description, insApp.ImageUrl.String, insApp.Email.String,
		insApp.AppSecretKey, insApp.ClientSecretKey, insApp.AesKey, 1)
	if err != nil {
		return errors.Wrap(err, "insert exec error")
	}

	if err != nil {
		return err
	}
	return nil
}
