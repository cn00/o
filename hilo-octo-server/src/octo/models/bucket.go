package models

import (
	"database/sql"

	"github.com/pkg/errors"
)

type Bucket struct {
	AppId      int
	BucketName string
}

type BucketDao struct {
}

func (*BucketDao) table() string {
	return "buckets"
}

func (*BucketDao) GetBucket(b *Bucket, appId int) error {
	rows, err := dbs.Query("SELECT app_id, bucket_name FROM buckets WHERE app_id = ?", appId)
	defer rows.Close()
	if err != nil {
		return errors.Wrap(err, "query row error")
	}

	if rows.Next() {
		if err := rows.Scan(&b.AppId, &b.BucketName); err != nil {
			return errors.Wrap(err, "query row scan error")
		}
		return nil
	}
	return nil
}

func (*BucketDao) Insert(b Bucket, tx *sql.Tx) error {
	_, err := tx.Exec(`INSERT INTO buckets(app_id, bucket_name) VALUES(?, ?)`, b.AppId, b.BucketName)
	return errors.Wrap(err, "exec error")
}

func (*BucketDao) Update(b Bucket, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE buckets SET bucket_name=? WHERE app_id=?`, b.BucketName, b.AppId)
	return errors.Wrap(err, "update exec error")
}
