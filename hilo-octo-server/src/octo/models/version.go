package models

import (
	"fmt"
	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"time"

	"database/sql"

	"github.com/pkg/errors"
)

type Version struct {
	AppId         int
	VersionId     int
	Description   string
	MaxRevision   int
	CopyVersionId sql.NullInt64
	CopyAppId     sql.NullInt64
	EnvId         sql.NullInt64
	State         int
	ApiAesKey     string
	UpdDatetime   time.Time
}

type VersionDao struct {
}

func (*VersionDao) colums() string {
	return "app_id,version_id,description,max_revision,copy_version_id,copy_app_id,env_id,state,api_aes_key,upd_datetime"
}

func (*VersionDao) table() string {
	return "versions"
}

func (*VersionDao) scan(rows *sql.Rows) (Version, error) {
	var rec Version
	err := rows.Scan(&rec.AppId, &rec.VersionId, &rec.Description, &rec.MaxRevision, &rec.CopyVersionId, &rec.CopyAppId, &rec.EnvId, &rec.State, &rec.ApiAesKey, &rec.UpdDatetime)
	return rec, errors.Wrap(err, "scan error")
}

func (dao *VersionDao) AddVersion(appId int, versionId int) error {

	description := fmt.Sprintf("version%d", versionId)
	_, updateErr := dbm.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,description,max_revision,state,upd_datetime) SELECT ?,?,?,0,?,? FROM DUAL WHERE NOT EXISTS (SELECT * FROM `+dao.table()+` where app_id=? and version_id=?)`,
		appId,
		versionId,
		description,
		octo.Data_ADD,
		time.Now(),
		appId,
		versionId,
	)
	return errors.Wrap(updateErr, "exec error")
}

func (dao *VersionDao) AddVersionWithEnvId(appId int, versionId int, envId int) error {
	description := fmt.Sprintf("version%d", versionId)
	_, updateErr := dbm.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,description,max_revision,env_id,state,upd_datetime) SELECT ?,?,?,0,?,?,? FROM DUAL WHERE NOT EXISTS (SELECT * FROM `+dao.table()+` where app_id=? and version_id=?)`,
		appId,
		versionId,
		description,
		envId,
		octo.Data_ADD,
		time.Now(),
		appId,
		versionId,
	)
	return errors.Wrap(updateErr, "exec error")
}

func (dao *VersionDao) Get(appId int, versionId int) (Version, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` from `+dao.table()+` where app_id=? and version_id=?`, appId, versionId)
	if err != nil {
		return Version{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return dao.scan(rows)
	}
	return Version{}, nil
}

func (dao *VersionDao) GetMaxRevision(appId int, versionId int) (int, error) {
	rows, err := dbs.Query(`SELECT max_revision FROM `+dao.table()+` WHERE app_id=? and version_id=?`, appId, versionId)
	return scanInt(rows, err)
}

func (dao *VersionDao) IncrementMaxRevision(appId int, versionId int, tx *sql.Tx) (int, error) {
	if _, err := tx.Exec(`SET @rev := 0`); err != nil {
		return 0, errors.Wrap(err, "exec error")
	}
	_, err := tx.Exec(`UPDATE versions
		SET max_revision = (@rev := max_revision + 1)
		WHERE app_id = ? AND version_id = ?`,
		appId,
		versionId,
	)
	if err != nil {
		return 0, errors.Wrap(err, "exec error")
	}
	rev := 0
	if err := tx.QueryRow("SELECT @rev").Scan(&rev); err != nil {
		return 0, errors.Wrap(err, "query row error")
	}
	return rev, nil
}

func (dao *VersionDao) UpdateMaxRevision(appId int, versionId int, maxRevision int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET max_revision=?, state=?, upd_datetime=? WHERE app_id=? and version_id=?`,
		maxRevision,
		octo.Data_UPDATE,
		time.Now(),
		appId,
		versionId,
	)
	return errors.Wrap(err, "exec error")
}

func (dao *VersionDao) GetListByAppIds(id int) ([]Version, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` from `+dao.table()+` where app_id = ? and state != ?`, id, octo.Data_DELETE)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	var res []Version
	for rows.Next() {
		rec, err := dao.scan(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scan error")
		}
		res = append(res, rec)
	}
	return res, nil
}

func (dao *VersionDao) Update(appId int, versionId int, description string, copyVersionId sql.NullInt64, copyAppId sql.NullInt64, envId int, apiAesKey string) error {
	_, err := dbm.Exec(`UPDATE `+dao.table()+` SET
						description=?,
						copy_version_id=?,
						copy_app_id=?,
						env_id=?,
						state=?,
						api_aes_key=?,
						upd_datetime=?
						WHERE app_id=? and version_id=?`,
		description, copyVersionId, copyAppId, envId, octo.Data_UPDATE, apiAesKey, time.Now(), appId, versionId)
	return errors.Wrap(err, "exec error")
}

func (dao *VersionDao) Delete(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET
					state=?, upd_datetime=?
					WHERE app_id = ? and version_id = ?`, octo.Data_DELETE, time.Now(), appId, versionId)

	return errors.Wrap(err, "update that version delete error")
}

func (dao *VersionDao) HardDelete(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+dao.table()+` WHERE
					state = ? and app_id = ? and version_id = ?`,
		octo.Data_DELETE, appId, versionId)
	return errors.Wrap(err, "Hard Delete error")
}

func (dao *VersionDao) GetSoftDeletedVersionByAppId(appId int, thirtyDaysAgo time.Time) ([]Version, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` from `+dao.table()+` where app_id = ? and state = ? and upd_datetime <= ?`, appId, octo.Data_DELETE, thirtyDaysAgo)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	var res []Version
	for rows.Next() {
		rec, err := dao.scan(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scan error")
		}
		res = append(res, rec)
	}
	return res, nil
}
