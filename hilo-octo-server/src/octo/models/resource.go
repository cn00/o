package models

import (
	"database/sql"
	"time"

	"hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
)

type Resource struct {
	Item
}

type ResourceDao struct {
	ItemDao
	resourceScanner
}

func NewResourceDao() *ResourceDao {
	dao := &ResourceDao{}
	dao.ItemDao = ItemDao{table: dao.table()}
	return dao
}

func (*ResourceDao) colums() string {
	return "app_id,version_id,id,revision_id,filename,object_name,size,generation,md5,tag,priority,state,build_number,upload_version_id,upd_datetime"
}

func (*ResourceDao) table() string {
	return "resources"
}

type resourceScanner struct {}

func (*resourceScanner) scan(rows *sql.Rows) (Resource, error) {
	var rec Resource
	err := rows.Scan(&rec.AppId, &rec.VersionId, &rec.Id, &rec.RevisionId, &rec.Filename, &rec.ObjectName, &rec.Size, &rec.Generation, &rec.Md5, &rec.Tag, &rec.Priority, &rec.State, &rec.BuildNumber, &rec.UploadVersionId, &rec.UpdDatetime)
	return rec, errors.Wrap(err, "scan error")
}

func (s *resourceScanner) scanResource(rows *sql.Rows, err error) (Resource, error) {
	if err != nil {
		return Resource{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return s.scan(rows)
	}
	return Resource{}, nil
}

func (s *resourceScanner) scanResources(rows *sql.Rows, err error) ([]Resource, error) {
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res []Resource
	for rows.Next() {
		file, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, file)
	}
	return res, nil
}

func (dao *ResourceDao) GetList(appId int, versionId int, revisionId int) ([]Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
								FROM `+dao.table()+` where app_id=? and version_id=? and revision_id > ? `, appId, versionId, revisionId)
	return dao.scanResources(rows, err)
}

func (dao *ResourceDao) GetDiffList(appId int, versionId int, revisionId int, targetRevisionId int) ([]Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
								FROM `+dao.table()+` where app_id=? and version_id=? and revision_id > ? and revision_id <= ? `, appId, versionId, revisionId, targetRevisionId)
	return dao.scanResources(rows, err)
}

func (dao *ResourceDao) GetByName(appId int, versionId int, name string) (Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and filename=?`, appId, versionId, name)
	return dao.scanResource(rows, err)
}

func (dao *ResourceDao) GetByMd5(appId int, versionId int, md5 string) (Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and md5=?`, appId, versionId, md5)
	return dao.scanResource(rows, err)
}

func (dao *ResourceDao) GetByNameForUpdate(appId int, versionId int, name string, tx *sql.Tx) (Resource, error) {
	rows, err := tx.Query(`SELECT `+dao.colums()+
		` FROM `+dao.table()+
		` WHERE app_id=? and version_id=? and filename=?`+
		` FOR UPDATE`, appId, versionId, name)
	return dao.scanResource(rows, err)
}

func (dao *ResourceDao) GetByObjectName(appId int, versionId int, objectName string) (Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and object_name=?`, appId, versionId, objectName)
	return dao.scanResource(rows, err)
}

func (dao *ResourceDao) GetByGenIsNullOrUploadVerIdNull(appId int) ([]Resource, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and (generation is null or upload_version_id is null)`, appId)
	return dao.scanResources(rows, err)
}

func (dao *ResourceDao) Replace(file Resource, useOldTagFlg bool, tx *sql.Tx) (Resource, error) {
	rec, err := dao.GetByNameForUpdate(file.AppId, file.VersionId, file.Filename, tx)
	if err != nil {
		return Resource{}, err
	}

	if (rec == Resource{}) {
		return Resource{}, nil
	}
	file.Id = rec.Id
	file.ObjectName = rec.ObjectName
	file.State = int(octo.Data_UPDATE)
	if useOldTagFlg {
		file.Tag = rec.Tag
	}
	return dao.Update(file, tx)
}

func (dao *ResourceDao) Update(file Resource, tx *sql.Tx) (Resource, error) {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET
							revision_id=?,
							filename=?,
							object_name=?,
							size=?,
							generation=?,
							md5=?,
							tag=?,
							priority=?,
							state=?,
							build_number=?,
							upload_version_id=?,
							upd_datetime=?
							WHERE app_id=? and version_id=? and id=?`,
		file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.AppId, file.VersionId, file.Id)
	return file, errors.Wrap(err, "exec error")
}

func (dao *ResourceDao) Insert(file Resource, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,revision_id,filename,object_name,size,generation,md5,tag,priority,state,build_number,upload_version_id,upd_datetime,id) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.Id)
	return errors.Wrap(err, "exec error")
}

func (dao *ResourceDao) InsertWithId(file Resource) error {
	_, err := dbm.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,revision_id,filename,object_name,size,generation,md5,tag,priority,state,build_number,upload_version_id,upd_datetime,id)
								SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,CASE WHEN MAX(id) IS NULL THEN 1 ELSE MAX(id)+1 END FROM `+dao.table()+` WHERE app_id=? and version_id=?`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.AppId, file.VersionId)
	return errors.Wrap(err, "exec error")
}

func (dao *ResourceDao) UpdateGenerationAndUploadVersionId(appId int, versionId int, revisionId int, objectName string, generation uint64, uploadVersionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET generation=?, upload_version_id=? WHERE app_id=? and version_id=? and revision_id=? and object_name=?`, generation, uploadVersionId, appId, versionId, revisionId, objectName)
	return errors.Wrap(err, "UpdateGeneration exec error")
}
