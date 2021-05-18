package models

import (
	"database/sql"
	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
	"time"
)

type Item struct {
	Id              int
	AppId           int
	VersionId       int
	RevisionId      int
	Filename        string
	ObjectName      sql.NullString
	Size            int
	Generation      sql.NullInt64
	Md5             sql.NullString
	Tag             sql.NullString
	Priority        int
	State           int
	BuildNumber     sql.NullString
	UploadVersionId sql.NullInt64
	UpdDatetime     time.Time
}

func (i *Item) Sync(src Item) {
	i.RevisionId = src.RevisionId
	i.ObjectName = src.ObjectName
	i.Size = src.Size
	i.Generation = src.Generation
	i.Md5 = src.Md5
	i.Tag = src.Tag
	i.Priority = src.Priority
	i.State = src.State
	i.BuildNumber = src.BuildNumber
	i.UploadVersionId = src.UploadVersionId
}

type ItemDao struct {
	table string
}

func (dao *ItemDao) CountByObjectName(objectName string) (int, error) {
	rows, err := dbs.Query(`SELECT count(object_name) FROM `+dao.table+` WHERE object_name=?`, objectName)
	return scanInt(rows, err)
}

func (dao *ItemDao) UpdateTag(appId int, versionId int, fileId int, revisionId int, tag string, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table+` SET
						revision_id=?,
						tag=?,
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		revisionId, tag, time.Now(), appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (dao *ItemDao) Delete(appId int, versionId int, fileId int, revisionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table+` SET
						revision_id=?,
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		revisionId, octo.Data_DELETE, time.Now(), appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (dao *ItemDao) DeleteAllByVersionId(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table+` SET
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=?`,
		octo.Data_DELETE, time.Now(), appId, versionId)
	return errors.Wrap(err, "exec error")
}

func (dao *ItemDao) HardDelete(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+dao.table+` WHERE
							state = ? and app_id = ? and version_id = ?`,
		octo.Data_DELETE, appId, versionId)
	return errors.Wrap(err, "Hard Delete error")
}
