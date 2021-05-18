package utils

import (
	"github.com/pkg/errors"
)

type Pagination struct {
	ResultCount int
	MaxPage     int
	CurrentPage int
	Limit       int
	ShowList    List
}

type Element interface{}
type List []Element

type PaginationUtil struct{}

func (*PaginationUtil) GetPagenation(list List, page int, limit int) (Pagination, error) {

	res := Pagination{Limit: limit, CurrentPage: page}
	res.ResultCount = len(list)

	if limit == 0 {
		errMsg := "limit can't be 0"
		return res, errors.New(errMsg)
	}

	res.MaxPage = res.ResultCount / limit
	if res.ResultCount%limit > 0 {
		res.MaxPage++
	}

	var showList List
	firstRecord := 0
	if page > 1 {
		firstRecord = limit * (page - 1)
	}
	lastRecord := res.ResultCount - 1
	if lastRecord+1 > limit {
		lastRecord = limit*page - 1
	}

	for i, record := range list {
		if i >= firstRecord && i <= lastRecord {
			showList = append(showList, record)
		}
	}
	res.ShowList = showList

	return res, nil
}
