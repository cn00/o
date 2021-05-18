package models

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"octo/utils"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
)

var resourceDao = NewResourceDao()

type AdminResourceDao struct {
	AdminItemDao
	resourceScanner
}

func NewAdminResourceDao() *AdminResourceDao {
	dao := &AdminResourceDao{}
	dao.AdminItemDao = AdminItemDao{table: dao.table()}
	return dao
}

func (dao *AdminResourceDao) colums() string {
	return resourceDao.colums()
}

func (dao *AdminResourceDao) table() string {
	return resourceDao.table()
}

func (dao *AdminResourceDao) GetList(appId int, versionId int, name string, objectName string, md5 string, tags []string, ids []int, revisionIds []int, revisionIdOver int, fromDate string, toDate string, showDeleted bool) (utils.List, error) {
	query := `SELECT ` + dao.colums() + `
								FROM ` + dao.table() + ` where app_id=? and version_id=?`
	args := []interface{}{appId, versionId}
	for _, s := range strings.Fields(name) {
		query += ` and filename like ? ESCAPE '$'`
		args = append(args, "%"+s+"%")
	}
	if len(objectName) > 0 {
		query += ` and object_name like ? ESCAPE '$'`
		args = append(args, "%"+objectName+"%")
	}
	if len(md5) > 0 {
		query += ` and md5 like ? ESCAPE '$'`
		args = append(args, "%"+md5+"%")
	}
	if len(ids) > 0 {
		idsStr := make([]string, len(ids))
		for n := range ids {
			idsStr[n] = strconv.Itoa(ids[n])
		}
		query += ` and id IN (` + strings.Join(idsStr, ",") + `)`
	}
	if len(revisionIds) > 0 {
		idsStr := make([]string, len(revisionIds))
		for n := range revisionIds {
			idsStr[n] = strconv.Itoa(revisionIds[n])
		}
		if revisionIdOver > 0 {
			over := strconv.Itoa(revisionIdOver)
			query += ` and (revision_id IN (` + strings.Join(idsStr, ",") + `) OR revision_id >= ` + over + `)`
		} else {
			query += ` and revision_id IN (` + strings.Join(idsStr, ",") + `)`
		}
	} else if revisionIdOver > 0 {
		over := strconv.Itoa(revisionIdOver)
		query += ` and revision_id >= ` + over
	}

	if len(fromDate) > 0 && len(toDate) > 0 {
		query += " and upd_datetime between ? and ?"
		args = append(args, fromDate, toDate)
	} else {
		if len(fromDate) > 0 {
			query += " and upd_datetime >= ?"
			args = append(args, fromDate)
		} else if len(toDate) > 0 {
			query += " and upd_datetime <= ?"
			args = append(args, toDate)
		}
	}
	if !showDeleted {
		query += " and state != ?"
		args = append(args, octo.Data_DELETE)
	}
	query += ` order by revision_id DESC`
	rows, err := dbs.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res utils.List
	for rows.Next() {
		file, err := dao.resourceScanner.scan(rows)
		if err != nil {
			return nil, err
		}
		if len(tags) > 0 {
			ftagList := utils.SplitTags(file.Tag.String)
		L:
			for _, ftag := range ftagList {
				for _, tag := range tags {
					if ftag == tag {
						res = append(res, file)
						break L
					}
				}
			}
		} else {
			res = append(res, file)
		}
	}
	return res, nil

}

func (dao *AdminResourceDao) GetById(appId int, versionId int, fileId int) (Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and id=?`, appId, versionId, fileId)
	return dao.scanResource(rows, err)
}

func (dao *AdminResourceDao) GetByIdFromTx(appId int, versionId int, fileId int, tx *sql.Tx) (Resource, error) {
	rows, err := tx.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and id=?`, appId, versionId, fileId)
	return dao.scanResource(rows, err)
}

func (dao *AdminResourceDao) GetByFileName(appId int, versionId int, filename string) (Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and filename=?`, appId, versionId, filename)
	return dao.scanResource(rows, err)
}

func (dao *AdminResourceDao) UpdateObjectName(appId int, versionId int, fileId int, objectName string) error {
	_, err := dbm.Exec(`UPDATE `+dao.table()+` SET
						object_name=?
						WHERE app_id=? and version_id=? and id=?`,
		objectName, appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (dao *AdminResourceDao) UpdateUrl(appId int, versionId int, fileId int, revisionId int, url string, md5 string) error {
	_, err := dbm.Exec(`UPDATE `+dao.table()+` SET
						revision_id=?,
						md5=?,
						url=?
						WHERE app_id=? and version_id=? and id=?`,
		revisionId, md5, url, appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (dao *AdminResourceDao) UpdateDatetime(appId int, versionId int, fileId int, datetime time.Time) error {
	_, err := dbm.Exec(`UPDATE `+dao.table()+` SET
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		datetime, appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (dao *AdminResourceDao) InsertWithId(file Resource, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,revision_id,filename,object_name,size,generation,md5,tag,priority,state,build_number,upload_version_id,upd_datetime,id)
								SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,CASE WHEN MAX(id) IS NULL THEN 1 ELSE MAX(id)+1 END FROM `+dao.table()+` WHERE app_id=? and version_id=?`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.AppId, file.VersionId)
	return errors.Wrap(err, "exec error")
}

func (dao *AdminResourceDao) Replace(file Resource, tx *sql.Tx) error {
	_, err := tx.Exec(`REPLACE INTO `+dao.table()+` (`+dao.colums()+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		file.AppId, file.VersionId, file.Id, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now())
	return errors.Wrap(err, "exec error")
}

func (dao *AdminResourceDao) CountByObjectName(objectName string) (int, error) {
	rows, err := dbs.Query(`SELECT count(object_name) FROM `+dao.table()+` WHERE object_name=?`, objectName)
	return scanInt(rows, err)
}
