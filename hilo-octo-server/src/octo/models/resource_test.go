package models

import (
	"database/sql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
	"time"
)

func TestResourceDao_InsertWithId(t *testing.T) {
	SetupEnvTest()

	file := Resource{
		Item{
			AppId:           1,
			VersionId:       1,
			RevisionId:      0,
			Filename:        "Test",
			ObjectName:      sql.NullString{String: "TestObject", Valid: true},
			Size:            123,
			Generation:      sql.NullInt64{Int64: 123, Valid: true},
			Md5:             sql.NullString{String: "md5", Valid: true},
			Tag:             sql.NullString{String: "tag", Valid: true},
			Priority:        0,
			State:           2,
			BuildNumber:     sql.NullString{String: "TestBuild", Valid: true},
			UploadVersionId: sql.NullInt64{Int64: 1, Valid: true},
			UpdDatetime:     time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local),
		},
	}

	dbmMock.ExpectExec("INSERT INTO resources").WithArgs(file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local), file.AppId, file.VersionId).WillReturnResult(sqlmock.NewResult(1, 1))

	dbm.Exec(`INSERT INTO `+resourceDao.table()+` (app_id,version_id,revision_id,filename,object_name,size,generation,md5,tag,priority,state,build_number,upload_version_id,upd_datetime,id)
								SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,CASE WHEN MAX(id) IS NULL THEN 1 ELSE MAX(id)+1 END FROM `+resourceDao.table()+` WHERE app_id=? and version_id=?`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local), file.AppId, file.VersionId)

	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
