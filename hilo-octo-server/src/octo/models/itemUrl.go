package models

import (
	"time"

	"database/sql"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
)

type ItemUrlDao struct {
	table string
}

func (dao *ItemUrlDao) DeleteAllByVersionId(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table+` SET
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=?`,
		octo.Data_DELETE, time.Now(), appId, versionId)
	return errors.Wrap(err, "exec error")
}

func (dao *ItemUrlDao) HardDelete(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+dao.table+` WHERE
					state = ? and app_id = ? and version_id = ?`,
		octo.Data_DELETE, appId, versionId)
	return errors.Wrap(err, "Hard Delete error")
}

func (dao *ItemUrlDao) HardDeleteByObjectName(appId, versionId int, objectName string, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+dao.table+` WHERE
					app_id=? and version_id=? and object_name=?`,
		appId, versionId, objectName)
	return errors.Wrap(err, "Hard Delete error")
}
