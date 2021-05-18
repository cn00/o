package utils

import (
	"strconv"
	"strings"
)

func SplitDependencies(s string) ([]int, error) {
	var deps []int
	for _, s := range strings.FieldsFunc(s, isComma) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		deps = append(deps, n)
	}
	return deps, nil
}

func JoinDependencies(deps []int) string {
	var ss []string
	for _, n := range deps {
		ss = append(ss, strconv.Itoa(n))
	}
	return strings.Join(ss, ",")
}

func IsDependent(deps []int, target int) bool {
	for _, d := range deps {
		if d == target {
			return true
		}
	}
	return false
}
