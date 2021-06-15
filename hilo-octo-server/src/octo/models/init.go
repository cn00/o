package models

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const tcpmultiDialTimeout = 2 * time.Second

type Config struct {
	ReadOnly               bool
	DatabaseMasterAddrs    string
	DatabaseMasterDbname   string
	DatabaseMasterUser     string
	DatabaseMasterPassword string
	DatabaseSlaveAddrs     string
	DatabaseSlaveDbname    string
	DatabaseSlaveUser      string
	DatabaseSlavePassword  string
}

type Dao interface {
	table() string
}

var (
	readOnly bool
	dbm      *sql.DB
	dbs      *sql.DB
)

var (
	dbmMock sqlmock.Sqlmock
	dbsMock sqlmock.Sqlmock
)

var versionDao = &VersionDao{}

func init() {
	mysql.RegisterDial("tcpmulti", func(addrsStr string) (net.Conn, error) {
		addrs := strings.Split(addrsStr, ",")
		if len(addrs) == 0 {
			return nil, errors.New("tcpmulti: addrs is empty")
		}
		shuffle(addrs)
		log.Println("[INFO] tcpmulti: addrs:", addrs)
		var lastErr error
		for _, addr := range addrs {
			conn, err := net.DialTimeout("tcp", addr, tcpmultiDialTimeout)
			lastErr = err
			if err == nil {
				//log.Println("[INFO] tcpmulti: connected:", addr, conn)
				return conn, nil
			}
		}
		return nil, errors.Wrap(lastErr, "tcpmulti: dial failed")
	})
}

func shuffle(a []string) {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

func Setup(config Config) error {
	readOnly = config.ReadOnly

	const dsnFormat = "%s:%s@tcpmulti(%s)/%s?timeout=3000s&parseTime=true&loc=Asia%%2FTokyo"

	if !readOnly {
		var err error
		dsn := fmt.Sprintf(dsnFormat,
			config.DatabaseMasterUser,
			config.DatabaseMasterPassword,
			config.DatabaseMasterAddrs,
			config.DatabaseMasterDbname,
		)
		dbm, err = sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		if err := CheckDBM(); err != nil {
			panic(err)
		}
	}

	{
		var err error
		dsn := fmt.Sprintf(dsnFormat,
			config.DatabaseSlaveUser,
			config.DatabaseSlavePassword,
			config.DatabaseSlaveAddrs,
			config.DatabaseSlaveDbname,
		)
		dbs, err = sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		if err := CheckDBS(); err != nil {
			panic(err)
		}
	}

	return nil
}

// 测试用setup 在这里mock用的DB生成文件
func SetupEnvTest() {
	var err error

	dbmMock, dbsMock, err = setupMock()
	if err != nil {
		panic(err)
	}
}

func setupMock() (sqlmock.Sqlmock, sqlmock.Sqlmock, error) {
	var dbmMock sqlmock.Sqlmock
	var dbsMock sqlmock.Sqlmock
	var err error
	if !readOnly {
		// DBはmockのDB Object像是归还
		dbm, dbmMock, err = sqlmock.New()
		if err != nil {
			return nil, nil, err
		}
	}
	{
		// DBはmockのDB Object像是归还
		dbs, dbsMock, err = sqlmock.New()
		if err != nil {
			return nil, nil, err
		}
	}

	return dbmMock, dbsMock, nil
}

func CheckDBM() error {
	if readOnly {
		log.Println("[INFO] CheckDBM: checkDB skipped because readOnly is enabled")
		return nil
	}
	err := checkDB(dbm)
	return errors.Wrap(err, "dbm check failed")
}

func CheckDBS() error {
	err := checkDB(dbs)
	return errors.Wrap(err, "dbs check failed")
}

func checkDB(db *sql.DB) error {
	var res string
	return db.QueryRow("SELECT 1").Scan(&res)
}
