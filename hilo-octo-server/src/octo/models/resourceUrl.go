package models

import (
	"database/sql"
	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"time"

	"github.com/pkg/errors"
)

type ResourceUrl struct {
	AppId       int
	VersionId   int
	RevisionId  int
	ObjectName  string
	Md5         sql.NullString
	Url         string
	UpdDatetime time.Time
}

type ResourceUrlDao struct {
	ItemUrlDao
	resourceUrlScanner
}

func NewResourceUrlDao() *ResourceUrlDao {
	dao := &ResourceUrlDao{}
	dao.ItemUrlDao = ItemUrlDao{table: dao.table()}
	return dao
}

func (*ResourceUrlDao) colums() string {
	return "app_id,version_id,revision_id,object_name,md5,url,upd_datetime"
}

func (*ResourceUrlDao) table() string {
	return "resource_urls"
}

type resourceUrlScanner struct {}

func (*resourceUrlScanner) scan(rows *sql.Rows) (ResourceUrl, error) {
	var rec ResourceUrl
	err := rows.Scan(&rec.AppId, &rec.VersionId, &rec.RevisionId, &rec.ObjectName, &rec.Md5, &rec.Url, &rec.UpdDatetime)
	return rec, errors.Wrap(err, "scan error")
}

func (s *resourceUrlScanner) scanResourceUrl(rows *sql.Rows, err error) (ResourceUrl, error) {
	if err != nil {
		return ResourceUrl{}, errors.Wrap(err, "query error")
	}
	defer rows.Close()
	for rows.Next() {
		return s.scan(rows)
	}
	return ResourceUrl{}, nil
}

func (dao *ResourceUrlDao) AddUrl(url ResourceUrl, tx *sql.Tx) error {

	_, err := tx.Exec(`INSERT IGNORE INTO `+dao.table()+` (`+dao.colums()+`) VALUES (?,?,?,?,?,?,?)`,
		url.AppId,
		url.VersionId,
		url.RevisionId,
		url.ObjectName,
		url.Md5,
		url.Url,
		url.UpdDatetime,
	)
	return errors.Wrap(err, "exec error")
}

func (dao *ResourceUrlDao) GetListByAppId(appId int) ([]ResourceUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id =?`, appId)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res []ResourceUrl
	for rows.Next() {
		resourceUrl, err := dao.scan(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, resourceUrl)
	}
	return res, nil

}

func (dao *ResourceUrlDao) GetUrlByObjectNameAndRevisionId(appId int, versionId int, objectName string, revisionId int) (ResourceUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and version_id=? and object_name=? and revision_id=?`, appId, versionId, objectName, revisionId)
	return dao.scanResourceUrl(rows, err)
}

func (dao *ResourceUrlDao) GetUrlByObjectNameAndRevisionIdLatest(appId int, versionId int, objectName string, revisionId int) (ResourceUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and version_id=? and object_name=? and revision_id <= ? ORDER BY revision_id DESC LIMIT 1`, appId, versionId, objectName, revisionId)
	return dao.scanResourceUrl(rows, err)
}

func (dao *ResourceUrlDao) GetUrlByObjectNameLatest(appId int, versionId int, objectName string) (ResourceUrl, error) {
	rows, err := dbs.Query(`SELECT `+dao.colums()+` FROM `+dao.table()+` where app_id=? and version_id=? and object_name=? ORDER BY revision_id DESC LIMIT 1`, appId, versionId, objectName)
	return dao.scanResourceUrl(rows, err)
}

func (dao *ResourceUrlDao) UpdateMd5(appId int, versionId int, objectName string, revisionId int, md5 string) error {
	_, err := dbm.Exec(`UPDATE `+dao.table()+` SET md5=? where app_id=? and version_id=? and object_name=? and revision_id=?`, md5, appId, versionId, objectName, revisionId)
	return errors.Wrap(err, "exec error")
}

func (dao *ResourceUrlDao) Delete(appId int, versionId int, fileId int, revisionId int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE `+dao.table()+` SET
						revision_id=?,
						state=?,
						upd_datetime=?
						WHERE app_id=? and version_id=? and id=?`,
		revisionId, octo.Data_DELETE, time.Now(), appId, versionId, fileId)
	return errors.Wrap(err, "exec error")
}
