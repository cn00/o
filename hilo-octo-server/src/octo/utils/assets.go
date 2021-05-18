package utils

import (
	"strings"
)

func SplitAssets(s string) []string {
	if len(s) == 0 {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func JoinAssets(assets []string) string {
	return strings.Join(assets, "\n")
}
