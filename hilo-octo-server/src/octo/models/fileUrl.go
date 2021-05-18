package models

import (
	"database/sql"
	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"log"
	"time"

	"github.com/pkg/errors"
)

type FileUrl struct {
	AppId       int
	VersionId   int
	RevisionId  int
	ObjectName  string
	Crc         uint32
	Md5         sql.NullString
	Url         string
	UpdDatetime time.Time
}

type FileUrlDao struct {
	ItemUrlDao
	fileUrlScanner
}

func NewFileUrlDao() *FileUrlDao {
	dao := &FileUrlDao{}
	dao.ItemUrlDao = ItemUrlDao{table: dao.table()}
	return dao
}

func (*FileUrlDao) colums() string {
	return "app_id,version_id,revision_id,object_name,crc,md5,url,upd_datetime"
}

func (*FileUrlDao) table() string {
	return "file_urls"
}

type fileUrlScanner struct {}

func (*fileUrlScanner) scan(rows *sql.Rows) (FileUrl, error) {
	var rec FileUrl
	err := rows.Scan(&rec.AppId, &rec.VersionId, &rec.RevisionId, &rec.ObjectName, &rec.Crc, &rec.Md5, &rec.Url, &rec.UpdDatetime)
	return rec, errors.Wrap(err, "scan error")
}

func (s *fileUrlScanner) scanFileUrl(rows *sql.Rows, err error) (FileUrl, error) {
	if err != nil {
		return FileUrl{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return s.scan(rows)
	}
	return FileUrl{}, nil
}

func (dao *FileUrlDao) AddUrl(url FileUrl, tx *sql.Tx) error {

	_, err := tx.Exec(`INSERT IGNORE INTO `+dao.table()+` (`+dao.colums()+`) VALUES (?,?,?,?,?,?,?,?)`,
		url.AppId,
		url.VersionId,
		url.RevisionId,
		url.ObjectName,
		url.Crc,
		url.Md5,
		url.Url,
		url.UpdDatetime,
	)
	return errors.Wrap(err, "exec error")
}

func (dao *FileUrlDao) GetListByAppId(appId int) ([]FileUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id =?`, appId)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res []FileUrl
	for rows.Next() {
		fileUrl, err := dao.scan(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, fileUrl)
	}
	return res, nil

}

func (dao *FileUrlDao) GetUrlByObjectNameAndRevisionId(appId int, versionId int, objectName string, revisionId int) (FileUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and version_id=? and object_name=? and revision_id=?`, appId, versionId, objectName, revisionId)
	return dao.scanFileUrl(rows, err)
}

func (dao *FileUrlDao) GetUrlByObjectNameAndRevisionIdLatest(appId int, versionId int, objectName string, revisionId int) (FileUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and version_id=? and object_name=? and revision_id <= ? ORDER BY revision_id DESC LIMIT 1`, appId, versionId, objectName, revisionId)
	return dao.scanFileUrl(rows, err)
}

func (dao *FileUrlDao) GetUrlByObjectNameLatest(appId int, versionId int, objectName string) (FileUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and version_id=? and object_name=? ORDER BY revision_id DESC LIMIT 1`, appId, versionId, objectName)
	return dao.scanFileUrl(rows, err)
}

func (dao *FileUrlDao) UpdateCrcAndMd5(appId int, versionId int, objectName string, revisionId int, crc uint32, md5 string) error {
	log.Println("[DEBUG] UpdateCrcAndMd5:", crc, md5, appId, versionId, objectName, revisionId)
	_, err := dbm.Exec(`UPDATE `+dao.table()+` SET crc=? , md5=? where app_id=? and version_id=? and object_name=? and revision_id=?`, crc, md5, appId, versionId, objectName, revisionId)
	return errors.Wrap(err, "exec error")
}

func (dao *FileUrlDao) DeleteFileUrl(appId, versionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+
		dao.table()+` where app_id=? and version_id = ? and state != ?`, appId, versionId, octo.Data_DELETE)

	return errors.Wrap(err, "delete error")
}
