package models

import (
	"database/sql"
	"octo/utils"
	"time"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
)

type File struct {
	Item
	Crc        uint32
	Assets     sql.NullString
	Dependency sql.NullString
}

func (f *File) Sync(src File) {
	f.Item.Sync(src.Item)
	f.Crc = src.Crc
	f.Assets = src.Assets
	f.Dependency = src.Dependency
}

type FileDao struct {
	ItemDao
	fileScanner
}

func NewFileDao() *FileDao {
	dao := &FileDao{}
	dao.ItemDao = ItemDao{table: dao.table()}
	return dao
}

func (f File) GetAssets() []string {
	if !f.Assets.Valid {
		return []string{}
	}
	return utils.SplitAssets(f.Assets.String)
}

func (*FileDao) colums() string {
	return "app_id,version_id,id,revision_id,filename,object_name,size,crc,generation,md5,tag,assets,dependency,priority,state,build_number,upload_version_id,upd_datetime"
}

func (*FileDao) table() string {
	return "files"
}

type fileScanner struct {}

func (*fileScanner) scan(rows *sql.Rows) (File, error) {
	var rec File
	err := rows.Scan(&rec.AppId, &rec.VersionId, &rec.Id, &rec.RevisionId, &rec.Filename, &rec.ObjectName, &rec.Size, &rec.Crc, &rec.Generation, &rec.Md5, &rec.Tag, &rec.Assets, &rec.Dependency, &rec.Priority, &rec.State, &rec.BuildNumber, &rec.UploadVersionId, &rec.UpdDatetime)
	return rec, errors.Wrap(err, "scan error")
}

func (s *fileScanner) scanFile(rows *sql.Rows, err error) (File, error) {
	if err != nil {
		return File{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return s.scan(rows)
	}
	return File{}, nil
}

func (s *fileScanner) scanFiles(rows *sql.Rows, err error) ([]File, error) {
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res []File
	for rows.Next() {
		file, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, file)
	}
	return res, nil
}

func (dao *FileDao) GetList(appId int, versionId int, revisionId int) ([]File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
								FROM `+dao.table()+` where app_id=? and version_id=? and revision_id > ? `, appId, versionId, revisionId)
	return dao.scanFiles(rows, err)
}

func (dao *FileDao) GetRangeList(appId int, versionId int, revisionId int, fromDate string, toDate string) ([]File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
	FROM `+dao.table()+` where app_id=? and version_id=? and revision_id > ? and upd_datetime between ? and ?`, appId, versionId, revisionId, fromDate, toDate)
	return dao.scanFiles(rows, err)
}

func (dao *FileDao) GetDiffList(appId int, versionId int, revisionId int, targetRevisionId int) ([]File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
								FROM `+dao.table()+` where app_id=? and version_id=? and revision_id > ? and revision_id <= ? `, appId, versionId, revisionId, targetRevisionId)
	return dao.scanFiles(rows, err)
}

func (dao *FileDao) GetMaxRevisionId(appId int, versionId int) (int, error) {
	rows, err := dbs.Query(`SELECT max(revision_id) FROM `+dao.table()+` WHERE app_id=? and version_id=?`, appId, versionId)
	return scanInt(rows, err)
}

func (dao *FileDao) GetByName(appId int, versionId int, name string) (File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and filename=?`, appId, versionId, name)
	return dao.scanFile(rows, err)
}

func (dao *FileDao) GetByMd5(appId int, versionId int, md5 string) (File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and md5=?`, appId, versionId, md5)
	return dao.scanFile(rows, err)
}

func (dao *FileDao) GetByNameForUpdate(appId int, versionId int, name string, tx *sql.Tx) (File, error) {
	rows, err := tx.Query(`SELECT `+dao.colums()+
		` FROM `+dao.table()+
		` WHERE app_id=? and version_id=? and filename=?`+
		` FOR UPDATE`, appId, versionId, name)
	return dao.scanFile(rows, err)
}

func (dao *FileDao) GetByObjectName(appId int, versionId int, objectName string) (File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and object_name=?`, appId, versionId, objectName)
	return dao.scanFile(rows, err)
}

func (dao *FileDao) GetByGenIsNullOrUploadVerIdNull(appId int) ([]File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and (generation is null or upload_version_id is null) and state != ?`, appId, octo.Data_DELETE)
	return dao.scanFiles(rows, err)
}

func (dao *FileDao) Replace(file File, useOldTagFlg bool, tx *sql.Tx) (File, error) {
	rec, err := dao.GetByNameForUpdate(file.AppId, file.VersionId, file.Filename, tx)
	if err != nil {
		return File{}, err
	}

	if (rec == File{}) {
		return File{}, nil
	}
	file.Id = rec.Id
	file.ObjectName = rec.ObjectName
	file.State = int(octo.Data_UPDATE)
	if useOldTagFlg {
		file.Tag = rec.Tag
	}
	return dao.Update(file, tx)
}

func (dao *FileDao) Update(file File, tx *sql.Tx) (File, error) {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET
							revision_id=?,
							filename=?,
							object_name=?,
							size=?,
							crc=?,
							generation=?,
							md5=?,
							tag=?,
							assets=?,
							dependency=?,
							priority=?,
							state=?,
							build_number=?,
							upload_version_id=?,
							upd_datetime = ?
							WHERE app_id=? and version_id=? and id=?`,
		file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation.Int64, file.Md5, file.Tag, file.Assets, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.AppId, file.VersionId, file.Id)
	return file, errors.Wrap(err, "exec error")
}

func (dao *FileDao) Insert(file File, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,revision_id,filename,object_name,size,crc,generation,md5,tag,assets,dependency,priority,state,build_number,upload_version_id,upd_datetime,id) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation.Int64, file.Md5, file.Tag, file.Assets, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.Id)
	return errors.Wrap(err, "exec error")
}

func (dao *FileDao) InsertWithId(file File) error {
	_, err := dbm.Exec(`INSERT INTO `+dao.table()+` (app_id,version_id,revision_id,filename,object_name,size,crc,generation,md5,tag,assets,dependency,priority,state,build_number,upload_version_id,upd_datetime,id)
								SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CASE WHEN MAX(id) IS NULL THEN 1 ELSE MAX(id)+1 END FROM files WHERE app_id=? and version_id=?`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation.Int64, file.Md5, file.Tag, file.Assets, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Now(), file.AppId, file.VersionId)
	return errors.Wrap(err, "exec error")
}

func (dao *FileDao) UpdateGenerationAndUploadVersionId(appId int, versionId int, revisionId int, crc uint32, objectName string, generation uint64, uploadVersionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET generation=?, upload_version_id=? WHERE app_id=? and version_id=? and revision_id=? and crc=? and object_name=?`, generation, uploadVersionId, appId, versionId, revisionId, crc, objectName)
	return errors.Wrap(err, "UpdateGeneration exec error")
}

func (dao *FileDao) GetDeletedFile(appId, versionId int) ([]File, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+`
							FROM `+dao.table()+` where app_id=? and version_id=? and state != ?`, appId, versionId, octo.Data_DELETE)
	return dao.scanFiles(rows, err)
}

func (dao *FileDao) DeleteFile(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+
		dao.table()+` where app_id=? and version_id = ? and state != ?`, appId, versionId, octo.Data_DELETE)

	return errors.Wrap(err, "delete error")
}
