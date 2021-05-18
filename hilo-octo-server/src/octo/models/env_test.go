package models

import (
	"fmt"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"database/sql"
)

func TestEnvDaoImpl_Insert(t *testing.T) {
	SetupEnvTest()
	// InsertするObjectを生成しておく
	var insEnv = Env{1,1,"iOS", sql.NullString{String:"iOS2",Valid:true}, }
	// 上記で生成したobjectをInsertの結果として返すように
	dbmMock.ExpectExec("INSERT INTO envs").WithArgs(1,"iOS","iOS2").
		WillReturnResult(sqlmock.NewResult(1,1))

	envDao := NewEnvDao()
	if err := envDao.Insert(insEnv); err != nil {
		t.Errorf("Env Insert Error %v\n", err)
	}

	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestEnvDao_GetList(t *testing.T) {
	SetupEnvTest()
	// Rowsを作成しておく
	rows := sqlmock.NewRows([]string{"app_id", "env_id", "name", "detail"}).
		AddRow(1, 1, "Test", "Test").
		AddRow(1, 2, "Test2", "Test2")
	// Selectで作成したRowが返す用に設定
	dbsMock.ExpectQuery("^SELECT (.+) FROM envs").WillReturnRows(rows)

	envDao := NewEnvDao()
	env, err := envDao.GetList(1)
	if err != nil {
		t.Fatalf("failed to get list: %v ", err)
	}
	if len(env) == 0 {
		fmt.Println("result is zero")
	}

	e := env[0].(Env)
	fmt.Println("filename:" + e.Name)

}

func TestEnvDaoImpl_Update(t *testing.T) {
	SetupEnvTest()
	envDao := NewEnvDao()
	var updEnv = Env {1,1, "iOS", sql.NullString{String:"iOS2", Valid:true},}
	dbmMock.ExpectExec("UPDATE envs").WillReturnResult(sqlmock.NewResult(1,1))

	if err := envDao.Update(updEnv); err != nil {
		t.Errorf("Env Update Error %v\n", err)
	}

	if err := dbmMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}



func TestEnvDaoImpl_GetByName(t *testing.T) {
	SetupEnvTest()
	rows := sqlmock.NewRows([]string{"app_id", "env_id", "name", "detail"}).
		AddRow(1, 1, "Windows", "Windows").
		AddRow(1, 2, "iOS", "iOS")
	dbsMock.ExpectQuery("^SELECT (.+) FROM envs").WillReturnRows(rows)
	envDao := NewEnvDao()
	env, err := envDao.GetByName(1, "Windows")
	if err != nil {
		t.Fatalf("failed to get env by name: %v ", err)
	}
	if len(env.Name) > 0 {
		fmt.Println("env is exist")
		return
	}
	fmt.Println("env is non exist")

}
