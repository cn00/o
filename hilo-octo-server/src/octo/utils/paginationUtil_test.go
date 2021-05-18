package utils

import (
	"reflect"
	"testing"
)

func TestGetPagenation(t *testing.T) {
	patterns := []struct {
		list     List
		page     int
		limit    int
		expected Pagination
	}{
		{
			list:  List{1},
			page:  1,
			limit: 1,
			expected: Pagination{
				ResultCount: 1,
				MaxPage:     1,
				CurrentPage: 1,
				Limit:       1,
				ShowList:    List{1},
			},
		},
		{
			list:  List{1, 2},
			page:  1,
			limit: 1,
			expected: Pagination{
				ResultCount: 2,
				MaxPage:     2,
				CurrentPage: 1,
				Limit:       1,
				ShowList:    List{1},
			},
		},
		{
			list:  List{1, 2},
			page:  2,
			limit: 1,
			expected: Pagination{
				ResultCount: 2,
				MaxPage:     2,
				CurrentPage: 2,
				Limit:       1,
				ShowList:    List{2},
			},
		},
		{
			list:  List{1, 2, 3},
			page:  1,
			limit: 2,
			expected: Pagination{
				ResultCount: 3,
				MaxPage:     2,
				CurrentPage: 1,
				Limit:       2,
				ShowList:    List{1, 2},
			},
		},
		{
			list:  List{1, 2, 3},
			page:  2,
			limit: 2,
			expected: Pagination{
				ResultCount: 3,
				MaxPage:     2,
				CurrentPage: 2,
				Limit:       2,
				ShowList:    List{3},
			},
		},
		{
			list:  List{1, 2, 3},
			page:  1,
			limit: 0,
			expected: Pagination{
				Limit:    0,
				ShowList: List{3},
			},
		},
	}
	var pagenationUtil PaginationUtil
	for i, pattern := range patterns {
		p, err := pagenationUtil.GetPagenation(pattern.list, pattern.page, pattern.limit)
		if err == nil && !reflect.DeepEqual(p, pattern.expected) {
			t.Fatalf("#%d: GetPagenation is wrong: %+v", i, p)
		}
	}
}
