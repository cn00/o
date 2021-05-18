package models

import (
	"database/sql"
	"fmt"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
	"time"
)

func TestFileDao_GetRangeList(t *testing.T) {

	SetupEnvTest()
	rows := sqlmock.NewRows([]string{"id", "app_id", "version_id", "revision_id", "filename",
		"object_name", "size", "crc", "generation", "md5", "tag", "dependency", "priority", "state", "build_number", "upload_version_id", "upd_datetime"}).
		AddRow(1, 1, 1, 0, "Test", "Test", 123, 123, 123, "md5", "tag", nil, 1, 2, "testnumber", 1, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local))
	dbsMock.ExpectQuery("^SELECT (.+) FROM files").WillReturnRows(rows)
	file, err := fileDao.GetRangeList(1, 1, 0, "2017-07-12 11:50:06", "2017-07-12 12:51:01")
	if err != nil {
		t.Fatalf("failed to range list: %v ", err)

	}

	if len(file) == 0 {
		fmt.Println("result is zero")
	}
	fmt.Println("filename:" + file[0].Filename)
}

func TestFileDao_InsertWithId(t *testing.T) {
	SetupEnvTest()
	file := File{
		Item: Item{
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
		Crc:             123,
		Assets:          sql.NullString{String: "", Valid: true},
		Dependency:      sql.NullString{String: "", Valid: true},
	}

	dbmMock.ExpectExec("INSERT INTO files").WithArgs(file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation.Int64, file.Md5, file.Tag, file.Assets, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local), file.AppId, file.VersionId).WillReturnResult(sqlmock.NewResult(1, 1))

	dbm.Exec(`INSERT INTO `+fileDao.table()+` (app_id,version_id,revision_id,filename,object_name,size,crc,generation,md5,tag,assets,dependency,priority,state,build_number,upload_version_id,upd_datetime,id)
								SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CASE WHEN MAX(id) IS NULL THEN 1 ELSE MAX(id)+1 END FROM files WHERE app_id=? and version_id=?`,
		file.AppId, file.VersionId, file.RevisionId, file.Filename, file.ObjectName, file.Size, file.Crc, file.Generation.Int64, file.Md5, file.Tag, file.Assets, file.Dependency, file.Priority, file.State, file.BuildNumber, file.UploadVersionId, time.Date(2017, 7, 12, 11, 51, 0, 0, time.Local), file.AppId, file.VersionId)

	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

func TestFile_Sync(t *testing.T) {
	type fields struct {
		Item       Item
		Crc        uint32
		Assets     sql.NullString
		Dependency sql.NullString
	}
	type args struct {
		src File
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "simple",
			fields: fields{
				Item:       Item{
					Id:              1,
					AppId:           2,
					VersionId:       3,
					RevisionId:      4,
					Filename:        "file",
					ObjectName:      sql.NullString{Valid: true, String: "object"},
					Size:            5,
					Generation:      sql.NullInt64{Valid: true, Int64: 6},
					Md5:             sql.NullString{Valid: true, String: "md5"},
					Tag:             sql.NullString{Valid: true, String: "tag"},
					Priority:        7,
					State:           8,
					BuildNumber:     sql.NullString{Valid: true, String: "build"},
					UploadVersionId: sql.NullInt64{Valid: true, Int64: 9},
					UpdDatetime:     time.Time{},
				},
				Crc:        10,
				Assets:     sql.NullString{Valid: true, String: "assets"},
				Dependency: sql.NullString{Valid: true, String: "dep"},
			},
			args:   args{
				src: File{
					Item:       Item{
						Id:              1, // will not sync
						AppId:           2, // will not sync
						VersionId:       3, // will not sync
						RevisionId:      40,
						Filename:        "file", // will not sync
						ObjectName:      sql.NullString{Valid: true, String: "new object"},
						Size:            50,
						Generation:      sql.NullInt64{Valid: true, Int64: 60},
						Md5:             sql.NullString{Valid: true, String: "new md5"},
						Tag:             sql.NullString{Valid: true, String: "new tag"},
						Priority:        70,
						State:           80,
						BuildNumber:     sql.NullString{Valid: true, String: "new build"},
						UploadVersionId: sql.NullInt64{Valid: true, Int64: 90},
						UpdDatetime:     time.Time{},
					},
					Crc:        100,
					Assets:     sql.NullString{Valid: true, String: "new assets"},
					Dependency: sql.NullString{Valid: true, String: "new dep"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := File{
				Item:       tt.fields.Item,
				Crc:        tt.fields.Crc,
				Assets:     tt.fields.Assets,
				Dependency: tt.fields.Dependency,
			}
			f.Sync(tt.args.src)
			if !reflect.DeepEqual(f, tt.args.src) {
				t.Errorf("Sync() want = %v, got %v", tt.args.src, f)
			}
		})
	}
}