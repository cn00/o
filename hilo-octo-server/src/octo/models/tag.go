package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

type Tag struct {
	AppId int
	TagId int
	Name  string
}

type TagDao struct {
}

func (*TagDao) table() string {
	return "tags"
}

func (*TagDao) GetList(appId int) ([]Tag, error) {
	rows, err := dbs.Query(`SELECT tag_id,name FROM tags where app_id=? ORDER BY tag_id`, appId)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close()

	var res []Tag
	for rows.Next() {
		tag := Tag{AppId: appId}
		if err := rows.Scan(&tag.TagId, &tag.Name); err != nil {
			return nil, errors.Wrap(err, "scan error")
		}
		res = append(res, tag)
	}
	return res, nil
}

func (*TagDao) AddTag(appId int, tagname string, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT IGNORE INTO tags (app_id,name,tag_id)
		SELECT ?,?,coalesce(max(tag_id),0)+1
		FROM tags
		WHERE app_id=?`,
		appId,
		tagname,
		appId,
	)
	return errors.Wrap(err, "exec error")
}
