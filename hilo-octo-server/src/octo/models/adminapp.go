package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

func DeleteApp(dao Dao, appID int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM `+dao.table()+` where app_id=?`, appID)
	return errors.Wrap(err, "delete error")
}
