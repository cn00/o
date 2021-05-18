package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

func scanInt(rows *sql.Rows, err error) (int, error) {
	if err != nil {
		return 0, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var ret int
	for rows.Next() {
		if err := rows.Scan(&ret); err != nil {
			return 0, errors.Wrap(err, "scan error")
		}
	}
	return ret, nil
}
