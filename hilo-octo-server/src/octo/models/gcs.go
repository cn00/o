package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

type Gcs struct {
	AppId     int
	ProjectId string
	Backet    string
	Location  string
}

type GcsDao struct {
}

func (*GcsDao) table() string {
	return "gcs"
}

func (*GcsDao) GetGcs(a *Gcs, appId int) error {
	var err error
	rows, err := dbs.Query("SELECT app_id,project_id,backet,location from gcs where app_id=?", appId)
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&a.AppId, &a.ProjectId, &a.Backet, &a.Location)
		if err != nil {
			return errors.Wrap(err, "query row error")
		}
		return nil
	}
	return nil
}

func (*GcsDao) Insert(a Gcs, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT INTO gcs(app_id, project_id, backet, location) VALUES(?,?,?,?)`, a.AppId, a.ProjectId, a.Backet, a.Location)
	return errors.Wrap(err, "insert exec query")
}

func (*GcsDao) Update(a Gcs, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE gcs SET project_id=?, backet=?, location=? WHERE app_id=?`, a.ProjectId, a.Backet, a.Location, a.AppId)
	return errors.Wrap(err, "update exec query")
}
