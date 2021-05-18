package models

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

func StartTransaction() (*sql.Tx, error) {
	tx, err := dbm.Begin()
	return tx, errors.Wrap(err, "transaction: begin failed")
}

func FinishTransaction(tx *sql.Tx, err error) error {
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("[INFO] Rollback error:", err)
		}
		return err
	}
	return errors.Wrap(tx.Commit(), "transaction: commit failed")
}
