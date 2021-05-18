package utils

import (
	"strconv"
	"strings"
)

func GetSearchRange(idsParam string) ([]int, int) {

	var ids []int
	var over int

	if idsParam != "" {
		for _, idStr := range strings.Split(idsParam, ",") {
			index := strings.Index(idStr, "-")
			if index == -1 {
				id, _ := strconv.Atoi(idStr)
				ids = append(ids, id)
			} else if index == 0 {
				value := strings.Replace(idStr, "-", "", -1)
				id, _ := strconv.Atoi(value)
				for i := 1; i <= id; i++ {
					ids = append(ids, i)
				}
			} else if index == len(idStr)-1 {
				value := strings.Replace(idStr, "-", "", -1)
				id, _ := strconv.Atoi(value)
				over = id
			} else {
				values := strings.Split(idStr, "-")
				start, _ := strconv.Atoi(values[0])
				end, _ := strconv.Atoi(values[1])
				for i := start; i <= end; i++ {
					ids = append(ids, i)
				}
			}
		}
	}

	return ids, over
}
