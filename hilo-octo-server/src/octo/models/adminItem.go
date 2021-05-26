package models

import (
	"database/sql"
	"hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
	"time"
)

type AdminItemDao struct {
	table string
}

func (dao *AdminItemDao) scan(rows *sql.Rows) (Item, error) {
	var rec Item
	err := rows.Scan(&rec.AppId, &rec.VersionId, &rec.Id, &rec.RevisionId, &rec.Filename, &rec.ObjectName, &rec.Size, &rec.Generation, &rec.Md5, &rec.Tag, &rec.Priority, &rec.State, &rec.BuildNumber, &rec.UploadVersionId, &rec.UpdDatetime)
	return rec, errors.Wrap(err, "scan error")
}

func (dao *AdminItemDao) GetItemById(appId int, versionId int, fileId int) (Item, error) {
	rows, err := dbs.Query(`SELECT app_id,version_id,id,revision_id,filename,object_name,size,generation,md5,tag,priority,state,build_number,upload_version_id,upd_datetime
							FROM `+dao.table+` where app_id=? and version_id=? and id=?`, appId, versionId, fileId)
	if err != nil {
		return Item{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return dao.scan(rows)
	}
	return Item{}, nil
}

func (dao *AdminItemDao) HardDelete(appId int, versionId int, fileId int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+dao.table+` WHERE app_id=? and version_id=? and id=?`,
		appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (dao *AdminItemDao) Delete(appId int, versionId int, fileId int) error {
	tx, err := StartTransaction()
	if err != nil {
		return err
	}
	revision, err := versionDao.IncrementMaxRevision(appId, versionId, tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE `+dao.table+` SET
						revision_id=?,
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		revision, octo.Data_DELETE, time.Now(), appId, versionId, fileId)

	return FinishTransaction(tx, errors.Wrap(err, "exec error"))
}

func (dao *AdminItemDao) DeleteByList(appId int, versionId int, fileIdList []int) error {
			tx, err := StartTransaction()
			if err != nil {
			return err
		}
			revision, err := versionDao.IncrementMaxRevision(appId, versionId, tx)
			if err != nil {
			return err
		}

			for _, id := range fileIdList {
			_, err = tx.Exec(`UPDATE `+dao.table+` SET
						revision_id=?,
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
			revision, octo.Data_DELETE, time.Now(), appId, versionId, id)
		if err != nil {
			return errors.Wrap(err, "exec error")
		}
	}

	return FinishTransaction(tx, errors.Wrap(err, "exec error"))
}

func (dao *AdminItemDao) Update(appId int, versionId int, fileId int, priority int, tag string, tx *sql.Tx) error {
	revision, err := versionDao.IncrementMaxRevision(appId, versionId, tx)
	if err != nil {
		return err
	}

	_, err = dbm.Exec(`UPDATE `+dao.table+` SET
						revision_id=?,
						tag=?,
						priority=?,
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		revision, tag, priority, octo.Data_UPDATE, time.Now(), appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}
