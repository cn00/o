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

var fileDao = NewFileDao()

type AdminFileDao struct {
	AdminItemDao
	fileScanner
}

func NewAdminFileDao() *AdminFileDao {
	dao := &AdminFileDao{}
	dao.AdminItemDao = AdminItemDao{table: dao.table()}
	return dao
}

func (dao *AdminFileDao) colums() string {
	return fileDao.colums()
}

func (dao *AdminFileDao) table() string {
	return fileDao.table()
}

func (dao *AdminFileDao) GetList(appId int, versionId int, name string, objectName string, md5 string, tags []string, ids []int, revisionIds []int, overRevisionId int, fromDate string, toDate string, showDeleted bool) (utils.List, error) {
	query := `SELECT ` + dao.colums() + `
								FROM files where app_id=? and version_id=?`
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
		if overRevisionId > 0 {
			over := strconv.Itoa(overRevisionId)
			query += ` and (revision_id IN (` + strings.Join(idsStr, ",") + `) OR revision_id >= ` + over + `)`
		} else {
			query += ` and revision_id IN (` + strings.Join(idsStr, ",") + `)`
		}
	} else if overRevisionId > 0 {
		over := strconv.Itoa(overRevisionId)
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
		file, err := dao.fileScanner.scan(rows)
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

func (dao *AdminFileDao) GetById(appId int, versionId int, fileId int) (File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM files where app_id=? and version_id=? and id=?`, appId, versionId, fileId)
	return dao.scanFile(rows, err)
}

func (dao *AdminFileDao) GetByIdFromTx(appId int, versionId int, fileId int, tx *sql.Tx) (File, error) {
	rows, err := tx.Query(`SELECT `+dao.colums()+`
							FROM files where app_id=? and version_id=? and id=?`, appId, versionId, fileId)
	return dao.scanFile(rows, err)
}

func (dao *AdminFileDao) InsertWithId(file File, tx *sql.Tx) (int, error) {
	var newID int
	if err := tx.QueryRow(`SELECT CASE WHEN MAX(id) IS NULL THEN 1 ELSE MAX(id)+1 END FROM files WHERE app_id=? and version_id=?`, file.AppId, file.VersionId).Scan(&newID); err != nil {
		return 0, err
	}
	_, err := tx.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,revision_id,filename,object_name,size,crc,generation,md5,tag,dependency,priority,state,build_number,upload_version_id,upd_datetime,id)
								SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation.Int64, file.Md5, file.Tag, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), newID)
	return newID, errors.Wrap(err, "exec error")
}

func (dao *AdminFileDao) Replace(file File, tx *sql.Tx) error {
	_, err := tx.Exec(`REPLACE INTO `+dao.table()+` (`+dao.colums()+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		file.AppId, file.VersionId, file.Id, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation, file.Md5, file.Tag, file.Assets, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now())
	return errors.Wrap(err, "exec error")
}

func (*AdminFileDao) UpdateObjectName(appId int, versionId int, fileId int, objectName string) error {
	_, err := dbm.Exec(`UPDATE files SET
						object_name=?
						WHERE app_id=? and version_id=? and id=?`,
		objectName, appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (*AdminFileDao) UpdateUrl(appId int, versionId int, fileId int, revisionId int, md5 string) error {
	_, err := dbm.Exec(`UPDATE files SET
						revision_id=?,
						md5=?
						WHERE app_id=? and version_id=? and id=?`,
		revisionId, md5, appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}

func (*AdminFileDao) UpdateDatetime(appId int, versionId int, fileId int, datetime time.Time) error {
	_, err := dbm.Exec(`UPDATE files SET
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		datetime, appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}
