package models

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAdminResourceDao_GetList(t *testing.T) {

	adminResourceDao := &AdminResourceDao{}

	SetupEnvTest()
	rows := sqlmock.NewRows([]string{"id", "app_id", "version_id", "revision_id", "filename",
		"object_name", "size", "generation", "md5", "tag", "priority", "state", "build_number", "upload_version_id", "upd_datetime"}).
		AddRow(1, 1, 1, 0, "Test", "Test", 123, 123, "md5", "tag", 1, 2, "testnumber", 1, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local))
	dbsMock.ExpectQuery("^SELECT (.+) FROM resources").WillReturnRows(rows)
	file, err := adminResourceDao.GetList(1, 1, "Test", "", "", nil, []int{1, 2}, []int{1}, 0, "2017-07-12 11:50:06", "2017-07-12 12:51:01", false)
	if err != nil {
		t.Fatalf("failed to range list: %v ", err)

	}

	if len(file) == 0 {
		fmt.Println("result is zero")
	}
	//fmt.Println("filename:" + file[0])
}

func TestAdminResourceDao_Update(t *testing.T) {

	SetupEnvTest()

	//dbmMock.ExpectBegin()
	dbmMock.ExpectExec("UPDATE resources").WithArgs("Test", 1, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	//tx, err := dbm.Begin()
	dbm.Exec("UPDATE resources set build_name=?, update_version_id where app_id=> and version_id=?", "Test", 1, 1, 1)
	//dbmMock.ExpectCommit()
	// we make sure that all expectations were met
	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAdminResourceDao_Replace(t *testing.T) {
	SetupEnvTest()
	//dbmMock.ExpectBegin()

	dbmMock.ExpectExec("REPLACE INTO resources").WithArgs(1, 1, 1, 0, "Test", "Test", 123, 123, "md5", "tag", 0, 2, "testnumber", 1, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local)).WillReturnResult(sqlmock.NewResult(1, 1))

	dbm.Exec(`REPLACE INTO `+resourceDao.table()+` (`+resourceDao.colums()+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, 1, 1, 1, 0, "Test", "Test", 123, 123, "md5", "tag", 0, 2, "testnumber", 1, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local))
	//dbmMock.ExpectCommit()

	// we make sure that all expectations were met
	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
func TestAdminResourceDao_InsertWithId(t *testing.T) {
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

	dbmMock.ExpectBegin()
	dbmMock.ExpectExec("INSERT INTO resources").WithArgs(file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Generation.Int64, file.Md5, file.Tag, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, sqlmock.AnyArg(), file.AppId, file.VersionId).WillReturnResult(sqlmock.NewResult(1, 1))
	dbmMock.ExpectCommit()

	tx, err := dbm.Begin()
	if err != nil {
		t.Errorf("Failed to begin transaction: %s", err)
		return
	}
	dao := AdminResourceDao{}
	if err := dao.InsertWithId(file, tx); err != nil {
		t.Errorf("Failed to InsertWithId: %s", err)
		return
	}
	if err := tx.Commit(); err != nil {
		t.Errorf("Failed to Commit: %s", err)
		return
	}

	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
